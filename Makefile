PREFIX?=/usr/local
_INSTDIR=$(PREFIX)
BINDIR?=$(_INSTDIR)/getwtxt
VERSION?=$(shell git tag | grep ^v | sort -V | tail -n 1)
GOFLAGS?=-tags netgo \
				 -ldflags '-X github.com/getwtxt/getwtxt/svc.Vers=${VERSION} -extldflags "-static"'

getwtxt: getwtxt.go go.mod go.sum
	@echo
	@echo Building getwtxt. This may take a minute or two.
	@mkdir logs
	go build $(GOFLAGS) \
		-o $@
	@echo
	@echo ...Done\!

.PHONY: clean
clean:
	@echo
	@echo Cleaning build and module caches...
	go clean -cache -modcache
	@echo
	@echo ...Done\!

.PHONY: update
update:
	@echo
	@echo Updating from upstream repository...
	@echo
	git pull --rebase
	@echo
	@echo ...Done\!

.PHONY: install
install:
	@echo
	@echo Installing getwtxt...
	@echo
	@echo Creating user/group...
	adduser -home $(BINDIR) --system --group getwtxt
	@echo
	@echo
	@echo Creating directories...
	mkdir -p $(BINDIR)/assets/tmpl $(BINDIR)/docs $(BINDIR)/logs
	@echo
	@echo Copying files...
	install -m755 getwtxt $(BINDIR)
	install -m644 getwtxt.yml $(BINDIR)
	install -m644 assets/style.css $(BINDIR)/assets
	install -m644 assets/tmpl/index.html $(BINDIR)/assets/tmpl
	install -m644 README.md $(BINDIR)/docs
	install -m644 LICENSE $(BINDIR)/docs
	install -m644 etc/getwtxt.service /etc/systemd/system
	@echo
	@echo
	@echo Setting ownership...
	chown -R getwtxt:getwtxt $(BINDIR)
	@echo
	@echo ...Done\! Don\'t forget to run
	@echo '         $$ systemctl enable getwtxt'

.PHONY: uninstall
uninstall:
	@echo
	@echo Uninstalling getwtxt...
	@echo
	@echo Stopping service if running...
	@echo systemctl stop getwtxt
	@systemctl stop getwtxt >/dev/null 2>&1 || true
	@echo
	@echo Disabling service autostart...
	@echo systemctl disable getwtxt
	@systemctl disable getwtxt >/dev/null 2>&1 || true
	@echo
	@echo Removing files
	rm -rf $(BINDIR)
	rm -f /etc/systemd/system/getwtxt.service
	@echo
	@echo Removing user
	- userdel getwtxt
	@echo
	@echo ...Done\!
