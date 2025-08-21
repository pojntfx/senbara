.PHONY: all libsenbara-gtk.so SenbaraGtk-1.0.gir SenbaraGtk-1.0.typelib docs example.js

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
		--library=senbara-gtk \
		--library-path=. \
		--output=$@ \
		--pkg=gtk4 \
		--warn-all \
		--identifier-prefix=SenbaraGtk \
		--symbol-prefix=senbara_gtk \
		main.h

SenbaraGtk-1.0.typelib: SenbaraGtk-1.0.gir
	g-ir-compiler $< -o $@

docs:
	gi-docgen generate --config docs.toml --output-dir=docs --no-namespace-dir SenbaraGtk-1.0.gir

example.js:
	LD_LIBRARY_PATH="${PWD}:${LD_LIBRARY_PATH}" GI_TYPELIB_PATH="${PWD}:${GI_TYPELIB_PATH}" ./example.js
