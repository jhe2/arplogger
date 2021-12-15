// arplogger is a tool for Linux systems that listens for arp packets on
// the specified interface(s) to discover new hosts appearing on the
// local IPv4 network.
//
// Copyright (c) 2021 Johannes Heimansberg
// License: MIT
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type DB struct {
	mu           sync.RWMutex
	databasePath string
}

func (db *DB) Init(dbPath string) error {
	db.databasePath = dbPath
	// TODO: Create db file if it does not exist
	return nil
}

// CheckMAC checks if the supplied MAC address is found in the database
func (db *DB) CheckMAC(mac string) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	f, err := os.OpenFile(db.databasePath, os.O_RDONLY, 0644)
	defer f.Close()
	if err != nil {
		return false, err
	}
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		entry := strings.Split(s.Text(), " ")
		if entry[0] == mac { // Match found
			return true, nil
		}
	}
	return false, nil
}

// Add adds a new entry to the database
func (db *DB) Add(mac string, ip string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	f, err := os.OpenFile(db.databasePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "%s %s\n", mac, ip)
	return nil
}
