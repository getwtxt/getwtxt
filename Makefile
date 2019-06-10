PREFIX?=/usr/local
_INSTDIR=$(PREFIX)
BINDIR?=$(_INSTDIR)/getwtxt
VERSION?=$(shell git tag | grep ^v | sort -V | tail -n 1)
GOFLAGS?=-tags netgo \
				 -ldflags '-X github.com/getwtxt/getwtxt/svc.Vers=${VERSION} -extldflags "-static"' \
				 -buildmode pie

getwtxt: getwtxt.go go.mod go.sum
	go build $(GOFLAGS) \
		-o $@

RM?=rm -f

clean:
	$(RM) getwtxt

update:
	git pull --rebase

install:
	adduser -home $(BINDIR) --system --group getwtxt
	mkdir -p $(BINDIR)/assets/tmpl $(BINDIR)/docs
	install -m755 getwtxt $(BINDIR)
	install -m644 getwtxt.yml $(BINDIR)
	install -m644 assets/style.css $(BINDIR)/assets
	install -m644 assets/tmpl/index.html $(BINDIR)/assets/tmpl
	install -m644 README.md $(BINDIR)/docs
	install -m644 LICENSE $(BINDIR)/docs
	install -m644 etc/getwtxt.service /etc/systemd/system
	chown -R getwtxt:getwtxt $(BINDIR)

uninstall:
	systemctl stop getwtxt >/dev/null 2>&1
	systemctl disable getwtxt >/dev/null 2>&1
	rm -rf $(BINDIR)
	rm -f /etc/systemd/system/getwtxt.service
	userdel getwtxt
