PREFIX?=/usr/local
_INSTDIR=$(PREFIX)
BINDIR?=$(_INSTDIR)/getwtxt
VERSION?=$(shell git describe --tags --abbrev=0)
GOFLAGS?=-ldflags '-X git.sr.ht/~gbmor/getwtxt/svc.Vers=${VERSION}'

getwtxt: getwtxt.go go.mod go.sum
	@printf "\n%s\n" "Building getwtxt. This may take a minute or two."
	@mkdir -p logs
	go build $(GOFLAGS) -o $@
	@printf "\n%s\n" "...Done!"

.PHONY: clean
clean:
	@printf "\n%s\n" "Cleaning build ..."
	go clean
	@printf "\n%s\n" "...Done!"

.PHONY: install
install:
	@printf "\n%s\n" "Installing getwtxt..."

	@printf "\n%s\n" "Creating user/group..."
	adduser -home $(BINDIR) --system --group getwtxt

	@printf "\n%s\n" "Creating directories..."
	mkdir -p $(BINDIR)/assets/tmpl $(BINDIR)/docs $(BINDIR)/logs $(BINDIR)/static

	@printf "\n%s\n" "Copying files..."
	install -m755 getwtxt $(BINDIR)
	@if [ -f "$(BINDIR)/getwtxt.yml" ]; then printf "%s\n" "getwtxt.yml exists. Skipping ..."; else printf "%s\n" "getwtxt.yml ..." && install -m644 getwtxt.yml "$(BINDIR)"; fi
	@if [ -f "$(BINDIR)/assets/style.css" ]; then printf "%s\n" "style.css exists. Skipping ..."; else printf "%s\n" "style.css ..." && install -m644 assets/style.css "$(BINDIR)/assets/style.css"; fi
	@if [ -f "$(BINDIR)/assets/tmpl/index.html" ]; then printf "%s\n" "tmpl/index.html exists. Skipping ..."; else printf "%s\n" "tmpl/index.html ..." && install -m644 assets/tmpl/index.html "$(BINDIR)/assets/tmpl/index.html"; fi
	install -m644 static/kognise.water.css.dark.min.css $(BINDIR)/static
	install -m644 README.md $(BINDIR)/docs
	install -m644 LICENSE $(BINDIR)/docs
	install -m644 etc/getwtxt.service /etc/systemd/system

	@printf "\n%s\n" "Setting ownership..."
	chown -R getwtxt:getwtxt $(BINDIR)

	@printf "\n%s\n" "If any files were skipped and there were changes upstream, you may need to merge them manually."

	@printf "\n%s\n\t%s\n\n" "...Done! Don't forget to run:" "systemctl enable getwtxt"

.PHONY: uninstall
uninstall:
	@printf "\n%s\n" "Uninstalling getwtxt..."

	@printf "\n%s\n%s\n" "Stopping service if running..." "systemctl stop getwtxt"
	@systemctl stop getwtxt >/dev/null 2>&1 || true

	@printf "\n%s\n%s\n" "Disabling service autostart..." "systemctl disable getwtxt"
	@systemctl disable getwtxt >/dev/null 2>&1 || true

	@printf "\n%s\n" "Removing files"
	rm -rf $(BINDIR)/assets
	rm -rf $(BINDIR)/logs
	rm -f $(BINDIR)/getwtxt
	rm -f $(BINDIR)/getwtxt.yml
	rm -f /etc/systemd/system/getwtxt.service

	@printf "\n%s\n" "Removing user"
	userdel getwtxt

	@printf "\n%s\n" "The database is still intact in /usr/local/getwtxt/"
	@printf "\n%s\n" "...Done!"
