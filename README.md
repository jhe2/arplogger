# arplogger

Copyright (c) 2021-2022 Johannes Heimansberg

SPDX-License-Identifier: MIT

``arplogger`` is a tool for Linux systems that listens for arp packets on
the specified interface(s) to discover new hosts appearing on the
local IPv4 network.

It is similar to the classic ``arpwatch`` utility and comes with the
following features:

## Features

- Passively listens for ARP packets on the local network to discover hosts
- Supports listening on multiple network interfaces
- Is meant to run with as few privileges as possible
  - It temporarily needs the Linux kernel's ``CAP_NET_RAW`` capability
- Logs discovered hosts to a logfile

Unlike ``arpwatch`` ``arplogger`` does not send e-mails by itself, but
instead just writes log entries to a logfile. That logfile can be used
by other tools to send alerts, if desired.

It keeps track of the hosts in a database file, where it adds a new line
for each newly discovered host, with the host's MAC address and its IPv4
address at that time.
