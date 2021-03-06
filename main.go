// arplogger is a tool for Linux systems that listens for arp packets on
// the specified interface(s) to discover new hosts appearing on the
// local IPv4 network.
//
// Copyright (c) 2021-2022 Johannes Heimansberg
// SPDX-License-Identifier: MIT
//
// To avoid running as root, it needs raw socket capabilities:
// chown root:arplogger ./arplogger
// chmod 750 ./arplogger
// setcap CAP_NET_RAW=p ./arplogger
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/mdlayher/arp"
	"kernel.org/pub/linux/libs/security/libcap/cap"
)

var (
	ifacesFlag   = flag.String("i", "eth0", "(comma-separated list of) network interface(s) to listen for ARP packets")
	logfileFlag  = flag.String("l", "/var/log/arplogger/arplogger.log", "logfile path")
	databaseFlag = flag.String("d", "/var/cache/arplogger/arplogger.db", "database path")
	versionFlag  = flag.Bool("v", false, "version information")
	version      = "undefined"
)

// checkEUID checks if the program is running with setuid or as root and
// warns about it
func checkEUID() error {
	euid := syscall.Geteuid()
	uid := syscall.Getuid()
	egid := syscall.Getegid()
	gid := syscall.Getgid()
	if uid != euid || gid != egid {
		return fmt.Errorf("Warning: Setuid detected: uids(%d vs %d), gids(%d vs %d)", uid, euid, gid, egid)
	}
	if uid == 0 {
		return fmt.Errorf("Warning: This program should not be run as root.")
	}
	return nil
}

// listen tries to open a raw socket on each of the supplied interfaces
// and returns a list of arp.Client objects; interfaces should be
// specified as a comma-separated list of valid network interface names
func listen(interfaces string) ([]*arp.Client, error) {
	var res []*arp.Client

	ocaps := cap.GetProc()
	defer ocaps.SetProc() // Restore original capabilities on return

	caps, err := ocaps.Dup()
	if err != nil {
		return nil, fmt.Errorf("Failed to dup caps: %v", err)
	}

	if on, _ := caps.GetFlag(cap.Permitted, cap.NET_RAW); !on {
		return nil, fmt.Errorf(
			"Insufficient privilege to open raw socket - want %q, have %q. Set with 'setcap CAP_NET_RAW=p %s'.",
			cap.NET_RAW,
			caps,
			os.Args[0])
	}

	if err := caps.SetFlag(cap.Effective, true, cap.NET_RAW); err != nil {
		return nil, fmt.Errorf("Unable to set capability: %v", err)
	}

	if err := caps.SetProc(); err != nil {
		return nil, fmt.Errorf("Unable to raise capabilities %q: %v", caps, err)
	}

	err = nil
	errorIfaces := ""
	// Loop over all supplied network interface names and try to open
	// a socket for listening for each interface. If at least one of
	// the supplied interfaces is invalid/cannot be used, err will
	// contain the last error encountered.
	for _, ifname := range strings.Split(interfaces, ",") {
		// Make sure the network interface exists
		iface, errIf := net.InterfaceByName(ifname)
		if errIf != nil {
			if errorIfaces != "" {
				errorIfaces = errorIfaces + ", "
			}
			errorIfaces = errorIfaces + ifname
			continue
		}

		c, errDial := arp.Dial(iface)
		if errDial != nil {
			if errorIfaces != "" {
				errorIfaces = errorIfaces + ", "
			}
			errorIfaces = errorIfaces + ifname
			continue
		}
		res = append(res, c)
	}
	if errorIfaces != "" {
		err = fmt.Errorf("Unable to use interface(s): %s", errorIfaces)
	}
	return res, err
}

// Go routine for reading ARP packets from a socket
func readSocket(c arp.Client, ifname string, lc chan string, db *DB) {
	for {
		arp, _, err := c.Read()
		if err != nil {
			return
		}
		res, err := db.CheckMAC(arp.SenderHardwareAddr.String())
		if !res {
			lc <- fmt.Sprintf("New host discovered on %s: %s (%s)", ifname, arp.SenderHardwareAddr, arp.SenderIP)
			db.Add(arp.SenderHardwareAddr.String(), arp.SenderIP.String())
		}
	}
}

// Go routine for writing ARP information collected by readSocket to the log
func writeLog(lc chan string) {
	for {
		log.Println(<-lc)
	}
}

func main() {
	var db DB
	flag.Parse()

	if *versionFlag {
		fmt.Printf("arplogger version: %v\n", version)
		fmt.Printf("Copyright (c) 2021-2022 Johannes Heimansberg\n")
		fmt.Printf("License: MIT\n")
		os.Exit(0)
	}

	// Prepare logfile, unless "-" was specified, which will be treated as stderr
	if *logfileFlag != "-" {
		f, err := os.OpenFile(*logfileFlag, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal("Error: ", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	err := checkEUID()
	if err != nil {
		fmt.Println(err.Error())
		log.Println(err.Error())
	}

	err = db.Init(*databaseFlag)
	if err != nil {
		log.Fatal("Error: Problem with database: ", err)
	}

	// Open socket(s)
	socks, err := listen(*ifacesFlag)
	if err != nil {
		if *logfileFlag != "-" {
			fmt.Fprintln(os.Stderr, "Error: Failed to open socket(s): ", err)
		}
		log.Println("Error: Failed to open socket(s): ", err)
	}
	if len(socks) == 0 {
		log.Fatal("Error: No valid interfaces found.")
	}

	logChan := make(chan string)

	go writeLog(logChan)

	ifaces := strings.Split(*ifacesFlag, ",")
	for i, s := range socks {
		defer s.Close()
		log.Printf("Starting reader thread for %s...\n", ifaces[i])
		go readSocket(*s, ifaces[i], logChan, &db)
	}

	for {
		time.Sleep(time.Second * 10)
	}
}
