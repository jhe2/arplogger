# arplogger Makefile
# Copyright (c) 2021-2022 Johannes Heimansberg
# SPDX-License-Identifier: MIT
PREFIX?=/usr/local
DESTDIR?=
# A typical install would install the binary owned by root with group
# arplogger. The program would then run under the arplogger user which
# is member of the arplogger group.
INSTUSER?=arplogger
INSTGROUP?=arplogger
LOGDIR?=/var/log/arplogger
DATADIR?=/var/cache/arplogger
VERSION=`git describe --always --long --dirty --tags||cat VERSION||echo "unknown"`

arplogger: main.go db.go
	CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)"

clean:
	-rm arplogger

test:
	go test

# User and group must exist before installing: useradd -r arplogger
install: arplogger
	-mkdir -p "$(DESTDIR)$(PREFIX)/bin"
	cp arplogger "$(DESTDIR)$(PREFIX)/bin"
	chown root:$(INSTGROUP) "$(DESTDIR)$(PREFIX)/bin/arplogger"
	chmod 0750 "$(DESTDIR)$(PREFIX)/bin/arplogger"
	setcap CAP_NET_RAW=p "$(DESTDIR)$(PREFIX)/bin/arplogger"
	mkdir "$(LOGDIR)"
	chown $(INSTUSER):$(INSTGROUP) "$(LOGDIR)"
	mkdir "$(DATADIR)"
	chown $(INSTUSER):$(INSTGROUP) "$(DATADIR)"

uninstall:
	rm -f "$(DESTDIR)$(PREFIX)/bin/arplogger"
	-rm -f "$(LOGDIR)/"*.log
	rmdir "$(LOGDIR)"
	-rm -f "$(DATADIR)/arplogger.db"
	rmdir "$(DATADIR)"
