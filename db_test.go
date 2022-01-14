// arplogger is a tool for Linux systems that listens for arp packets on
// the specified interface(s) to discover new hosts appearing on the
// local IPv4 network.
//
// Copyright (c) 2021-2022 Johannes Heimansberg
// License: MIT
package main

import "testing"

func TestDb(t *testing.T) {
	var db DB
	err := db.Init("/tmp/test.db")
	if err != nil {
		t.Fatalf("Init failed")
	}
	db.Clear()

	testData := []struct {
		mac                            string
		ip                             string
		expectedSearchErrorBeforeDbAdd bool
		expectedDbAddResult            bool
		expectedSearchErrorAfterDbAdd  bool
		expectedSearchResultAfterDbAdd bool
	}{
		{"00:11:22:33:44:55", "1.2.3.4", false, true, false, true},
		{"00-11-22-33-44-55", "1.2.3.4", false, true, false, true},
		{"00F11:22:33:44:55", "1.2.3.4", true, false, true, false},
		{"00:11:22:33:44:55", "inv.alid", false, false, false, false},
		{"in:va:li:dm:ac", "inv.alid", true, false, true, false},
	}

	for _, test := range testData {
		db.Clear()

		res, err := db.CheckMAC(test.mac)
		if test.expectedSearchErrorBeforeDbAdd && err == nil {
			t.Errorf("Expected CheckMAC to fail, but it didn't: MAC=%q res=%t", test.mac, res)
		} else if !test.expectedSearchErrorBeforeDbAdd && err != nil {
			t.Errorf("Expected CheckMAC to succeed, but it didn't: MAC=%q res=%t err=%s", test.mac, res, err.Error())
		}
		if err == nil && res {
			t.Errorf("Expected CheckMAC to return false, but it didn't: MAC=%q", test.mac)
		}

		err = db.Add(test.mac, test.ip)
		if err != nil && test.expectedDbAddResult {
			t.Errorf("Expected 'Add' to succeed: MAC=%q IP=%q", test.mac, test.ip)
		} else if err == nil && !test.expectedDbAddResult {
			t.Errorf("Expected 'Add' to fail: MAC=%q IP=%q", test.mac, test.ip)
		}

		// Expect success for valid data
		res, err = db.CheckMAC(test.mac)
		if test.expectedSearchErrorAfterDbAdd && err == nil {
			t.Errorf("Expected CheckMAC to fail, but it didn't: MAC=%q res=%t", test.mac, res)
		} else if !test.expectedSearchErrorAfterDbAdd && err != nil {
			t.Errorf("Expected CheckMAC to succeed, but it didn't: MAC=%q res=%t err=%s", test.mac, res, err.Error())
		}
		if test.expectedSearchResultAfterDbAdd != res {
			t.Errorf("Unexpected CheckMAC result for MAC=%q: want(%t) != have(%t)", test.mac, test.expectedSearchResultAfterDbAdd, res)
		}
	}
}
