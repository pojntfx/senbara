#!/usr/bin/env -S gjs -m

import GObject from "gi://GObject";
import Gio from "gi://Gio";
import Gtk from "gi://Gtk?version=4.0";
import SenbaraGtk from "gi://SenbaraGtk?version=1.0";
import system from "system";

SenbaraGtk.init_types();

const ExampleApplication = GObject.registerClass(
  {
    GTypeName: "ExampleApplication",
  },
  class ExampleApplication extends Gtk.Application {
    constructor() {
      super({
        application_id: "com.pojtinger.felicitas.senbaragtk.Example",
        flags: Gio.ApplicationFlags.DEFAULT_FLAGS,
      });
    }

    #window = null;

    vfunc_activate() {
      this.#window = SenbaraGtk.MainApplicationWindow.new();
      this.#window.set_application(this);

      this.#window.connect("button-test-clicked", () => {
        console.log("Test button clicked");

        this.#window.set_test_button_sensitive(false);
        console.log("Test button disabled");

        setTimeout(() => {
          this.#window.set_test_button_sensitive(true);
          console.log("Test button re-enabled");
        }, 3000);
      });

      this.#window.present();
    }
  }
);

new ExampleApplication().run([system.programInvocationName, ...ARGV]);
