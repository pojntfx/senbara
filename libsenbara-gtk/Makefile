.PHONY: all libsenbara-gtk.so SenbaraGtk-1.0.gir SenbaraGtk-1.0.typelib docs example.js

PREFIX ?= /usr
DESTDIR ?=

all: libsenbara-gtk.so SenbaraGtk-1.0.typelib docs

libsenbara-gtk.so:
	go generate ./...
	go build -buildmode=c-shared -o libsenbara-gtk.so main.go

SenbaraGtk-1.0.gir: libsenbara-gtk.so
	g-ir-scanner \
		--namespace=SenbaraGtk \
		--nsversion=1.0 \
		--include=GObject-2.0 \
		--include=Gtk-4.0 \
		--include=Adw-1 \
		--library=senbara-gtk \
		--library-path=. \
		--output=$@ \
		--pkg=gtk4 \
		--pkg=libadwaita-1 \
		--warn-all \
		--identifier-prefix=SenbaraGtk \
		--symbol-prefix=senbara_gtk \
		main.h

SenbaraGtk-1.0.typelib: SenbaraGtk-1.0.gir
	g-ir-compiler $< -o $@

docs: SenbaraGtk-1.0.gir
	gi-docgen generate --config docs.toml --output-dir=docs --no-namespace-dir SenbaraGtk-1.0.gir

example.js:
	LD_LIBRARY_PATH="${PWD}:${LD_LIBRARY_PATH}" GI_TYPELIB_PATH="${PWD}:${GI_TYPELIB_PATH}" ./example.js

install:
	install -D -m 755 libsenbara-gtk.so $(DESTDIR)$(PREFIX)/lib64/libsenbara-gtk.so
	install -D -m 644 SenbaraGtk-1.0.typelib $(DESTDIR)$(PREFIX)/lib64/girepository-1.0/SenbaraGtk-1.0.typelib
	install -D -m 644 SenbaraGtk-1.0.gir $(DESTDIR)$(PREFIX)/share/gir-1.0/SenbaraGtk-1.0.gir

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/lib64/libsenbara-gtk.so
	rm -f $(DESTDIR)$(PREFIX)/lib64/girepository-1.0/SenbaraGtk-1.0.typelib
	rm -f $(DESTDIR)$(PREFIX)/share/gir-1.0/SenbaraGtk-1.0.gir
