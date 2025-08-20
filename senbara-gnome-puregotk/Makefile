# Makefile for Senbara PureGoTK widget

PACKAGE_NAME = Senbara
VERSION = 1.0
LIBRARY_NAME = libsenbara
LIBRARY_BASE_NAME = senbara

# GTK version (adjust as needed)
GTK_VERSION = 4.0

# Directories
PREFIX ?= /usr/local
LIBDIR = $(PREFIX)/lib
GIRDIR = $(LIBDIR)/girepository-1.0
INCLUDEDIR = $(PREFIX)/include

# Build flags
GTK_CFLAGS = $(shell pkg-config --cflags gtk4)
GTK_LIBS = $(shell pkg-config --libs gtk4)

.PHONY: all clean install install-user test docs

all: $(LIBRARY_NAME).so $(PACKAGE_NAME)-$(VERSION).typelib

# Build the Go shared library
$(LIBRARY_NAME).so: cmd/senbara-gnome-puregotk/main.go
	@echo "Building Go shared library..."
	go build -buildmode=c-shared -o $(LIBRARY_NAME).so cmd/senbara-gnome-puregotk/main.go
	@echo "Library built successfully"

# Generate GIR file using g-ir-scanner
$(PACKAGE_NAME)-$(VERSION).gir: cmd/senbara-gnome-puregotk/senbara.h $(LIBRARY_NAME).so
	@echo "Generating GIR file..."
	g-ir-scanner \
		--namespace=$(PACKAGE_NAME) \
		--nsversion=$(VERSION) \
		--include=GObject-2.0 \
		--include=Gtk-$(GTK_VERSION) \
		--library=senbara \
		--library-path=. \
		--output=$@ \
		--pkg=gtk4 \
		--warn-all \
		--identifier-prefix=Senbara \
		--symbol-prefix=senbara \
		cmd/senbara-gnome-puregotk/senbara.h
	@echo "GIR file generated successfully"

# Compile GIR to typelib
$(PACKAGE_NAME)-$(VERSION).typelib: $(PACKAGE_NAME)-$(VERSION).gir
	@echo "Compiling typelib..."
	g-ir-compiler $< -o $@
	@echo "Typelib compiled successfully"

# System-wide installation (requires sudo)
install: all
	@echo "Installing system-wide..."
	sudo mkdir -p $(LIBDIR) $(GIRDIR) $(INCLUDEDIR)
	sudo cp $(LIBRARY_NAME).so $(LIBDIR)/
	sudo cp $(PACKAGE_NAME)-$(VERSION).typelib $(GIRDIR)/
	sudo cp cmd/senbara-gnome-puregotk/senbara.h $(INCLUDEDIR)/
	sudo ldconfig
	@echo "Installation complete"

# User installation (no sudo required)
install-user: all
	@echo "Installing for user..."
	mkdir -p ~/.local/lib ~/.local/lib/girepository-1.0 ~/.local/include
	cp $(LIBRARY_NAME).so ~/.local/lib/
	cp $(PACKAGE_NAME)-$(VERSION).typelib ~/.local/lib/girepository-1.0/
	cp cmd/senbara-gnome-puregotk/senbara.h ~/.local/include/
	@echo "User installation complete"
	@echo "Add these to your shell profile:"
	@echo "export LD_LIBRARY_PATH=\$$HOME/.local/lib:\$$LD_LIBRARY_PATH"
	@echo "export GI_TYPELIB_PATH=\$$HOME/.local/lib/girepository-1.0:\$$GI_TYPELIB_PATH"

# Test the installation with a simple GJS script
test: install-user
	@echo "Testing with GJS..."
	@echo "#!/usr/bin/env gjs" > test.js
	@echo "" >> test.js
	@echo "imports.gi.versions.Gtk = '4.0';" >> test.js
	@echo "imports.gi.versions.Senbara = '1.0';" >> test.js
	@echo "" >> test.js
	@echo "const { Gtk, Senbara } = imports.gi;" >> test.js
	@echo "" >> test.js
	@echo "// Initialize" >> test.js
	@echo "Gtk.init();" >> test.js
	@echo "Senbara.init_types();" >> test.js
	@echo "" >> test.js
	@echo "// Create window" >> test.js
	@echo "let window = new Senbara.PureGoTKMainWindow();" >> test.js
	@echo "print('Window created successfully!');" >> test.js
	@echo "" >> test.js
	@echo "// Connect signal" >> test.js
	@echo "window.connect('button-test-clicked', () => {" >> test.js
	@echo "    print('Button test clicked!');" >> test.js
	@echo "});" >> test.js
	@echo "" >> test.js
	@chmod +x test.js
	GI_TYPELIB_PATH=~/.local/lib/girepository-1.0:$$GI_TYPELIB_PATH \
	LD_LIBRARY_PATH=~/.local/lib:$$LD_LIBRARY_PATH \
	./test.js
	rm -f test.js

# Generate documentation using gi-docgen
docs: $(PACKAGE_NAME)-$(VERSION).gir
	@echo "Generating documentation..."
	@mkdir -p docs
	gi-docgen generate \
		--quiet \
		--add-include-path=. \
		--config=docs.toml \
		--output-dir=docs \
		--no-namespace-dir \
		--content-dir=. \
		$(PACKAGE_NAME)-$(VERSION).gir
	@echo "Documentation generated in docs/ directory"

# Check exports in the built library
check-exports: $(LIBRARY_NAME).so
	@echo "Exported symbols:"
	nm -gD $(LIBRARY_NAME).so | grep " T " | grep senbara

# Clean build artifacts
clean:
	rm -f $(LIBRARY_NAME).so $(LIBRARY_NAME).h
	rm -f $(PACKAGE_NAME)-$(VERSION).gir $(PACKAGE_NAME)-$(VERSION).typelib
	rm -f test.js
	rm -rf docs

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Build library and typelib"
	@echo "  install       - Install system-wide (requires sudo)"
	@echo "  install-user  - Install for current user only"
	@echo "  test          - Build, install for user, and test"
	@echo "  docs          - Generate HTML documentation"
	@echo "  check-exports - Show exported symbols"
	@echo "  clean         - Remove build artifacts"
	@echo "  help          - Show this help"