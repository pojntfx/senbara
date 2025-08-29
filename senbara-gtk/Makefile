.PHONY: all libsenbara-gtk-0.1.so SenbaraGtk-0.1.gir SenbaraGtk-0.1.typelib docs example.js

PREFIX ?= /usr
DESTDIR ?=

all: libsenbara-gtk-0.1.so SenbaraGtk-0.1.typelib docs

libsenbara-gtk-0.1.so:
	go generate ./...
	go build -buildmode=c-shared -o /tmp/libsenbara-gtk-0.1.so .
	mv /tmp/libsenbara-gtk-0.1.so libsenbara-gtk-0.1.so

SenbaraGtk-0.1.gir: libsenbara-gtk-0.1.so
	g-ir-scanner --warn-all --library=senbara-gtk-0.1 --library-path=. --pkg=gtk4 --pkg=libadwaita-1 --namespace=SenbaraGtk --nsversion=0.1 --include=Gtk-4.0 --include=Adw-1 --output=$@ --identifier-prefix=SenbaraGtk --symbol-prefix=senbara_gtk senbara-gtk.h

SenbaraGtk-0.1.typelib: SenbaraGtk-0.1.gir
	g-ir-compiler $< -o $@

docs: SenbaraGtk-0.1.gir
	gi-docgen generate --config docs.toml --output-dir=docs --no-namespace-dir SenbaraGtk-0.1.gir

example.js:
	LD_LIBRARY_PATH="${PWD}:${LD_LIBRARY_PATH}" GI_TYPELIB_PATH="${PWD}:${GI_TYPELIB_PATH}" ./example.js

install:
	install -D -m 755 libsenbara-gtk-0.1.so $(DESTDIR)$(PREFIX)/lib64/libsenbara-gtk-0.1.so
	install -D -m 644 senbara-gtk.h $(DESTDIR)$(PREFIX)/include/senbara-gtk-0.1/senbara-gtk-0.1.h
	install -D -m 644 SenbaraGtk-0.1.typelib $(DESTDIR)$(PREFIX)/lib64/girepository-1.0/SenbaraGtk-0.1.typelib
	install -D -m 644 SenbaraGtk-0.1.gir $(DESTDIR)$(PREFIX)/share/gir-1.0/SenbaraGtk-0.1.gir
	install -D -m 644 senbara-gtk.pc $(DESTDIR)$(PREFIX)/lib64/pkgconfig/senbara-gtk-0.1.pc
	mkdir -p $(DESTDIR)$(PREFIX)/share/doc/libsenbara-gtk-0.1/
	cp docs/* $(DESTDIR)$(PREFIX)/share/doc/libsenbara-gtk-0.1/
	mv $(DESTDIR)$(PREFIX)/share/doc/libsenbara-gtk-0.1/docs.devhelp2 $(DESTDIR)$(PREFIX)/share/doc/libsenbara-gtk-0.1/libsenbara-gtk-0.1.devhelp2

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/lib64/libsenbara-gtk-0.1.so
	rm -f $(DESTDIR)$(PREFIX)/include/senbara-gtk-0.1/senbara-gtk-0.1.h
	rm -f $(DESTDIR)$(PREFIX)/lib64/girepository-1.0/SenbaraGtk-0.1.typelib
	rm -f $(DESTDIR)$(PREFIX)/share/gir-1.0/SenbaraGtk-0.1.gir
	rm -f $(DESTDIR)$(PREFIX)/lib64/pkgconfig/senbara-gtk-0.1.pc
	rm -rf $(DESTDIR)$(PREFIX)/share/doc/libsenbara-gtk-0.1
