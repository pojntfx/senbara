#!/usr/bin/env -S gjs -m

import Gtk from "gi://Gtk?version=4.0";
import Gio from "gi://Gio";
import SenbaraGtk from "gi://SenbaraGtk?version=1.0";

class SenbaraGtkExampleApp {
  constructor() {
    this.app = new Gtk.Application({
      application_id: "com.pojtinger.felicitas.senbaragtk.Example",
      flags: Gio.ApplicationFlags.FLAGS_NONE,
    });

    this.app.connect("activate", this._onActivate.bind(this));
  }

  _onActivate() {
    SenbaraGtk.init_types();

    this.window = SenbaraGtk.MainApplicationWindow.new();

    this.window.connect("button-test-clicked", () => {
      print("Button clicked");
    });

    this.window.set_application(this.app);
    this.window.present();
  }

  run(argv) {
    return this.app.run(argv);
  }
}

const app = new SenbaraGtkExampleApp();
app.run([]);
