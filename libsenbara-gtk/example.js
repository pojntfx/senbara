#!/usr/bin/env -S gjs -m

import Adw from "gi://Adw?version=1";
import GObject from "gi://GObject";
import Gio from "gi://Gio";
import SenbaraGtk from "gi://SenbaraGtk?version=1.0";
import system from "system";

const ExampleApplication = GObject.registerClass(
  {
    GTypeName: "ExampleApplication",
  },
  class ExampleApplication extends Adw.Application {
    constructor() {
      super({
        application_id: "com.pojtinger.felicitas.senbaragtk.Example",
        flags: Gio.ApplicationFlags.DEFAULT_FLAGS,
      });
    }

    #window = null;

    vfunc_activate() {
      this.#window = new SenbaraGtk.MainApplicationWindow({
        application: this,
      });

      this.#window.connect("button-test-clicked", () => {
        console.log("Test button clicked");

        this.#window.show_toast("Button was clicked!");
        this.#window.test_button_sensitive = false;

        setTimeout(() => {
          this.#window.show_toast("Button re-enabled after 3 seconds");
          this.#window.test_button_sensitive = true;
        }, 3000);
      });

      this.#window.present();
    }
  }
);

new ExampleApplication().run([system.programInvocationName, ...ARGV]);
