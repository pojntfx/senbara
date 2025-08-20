#!/usr/bin/env gjs

imports.gi.versions.Gtk = '4.0';
imports.gi.versions.Senbara = '1.0';

const { Gtk, Gio, Senbara } = imports.gi;

class SenbaraExampleApp {
    constructor() {
        this.app = new Gtk.Application({
            application_id: 'com.pojtinger.felicitas.SenbaraPureGoTK.Example',
            flags: Gio.ApplicationFlags.FLAGS_NONE
        });
        
        this.app.connect('activate', this._onActivate.bind(this));
    }
    
    _onActivate() {
        // Initialize the Senbara type system
        Senbara.init_types();
        
        // Create the main window using the proper constructor
        this.window = Senbara.PureGoTKMainWindow.new();
        
        // Connect the button signal
        this.window.connect('button-test-clicked', () => {
            print('Button clicked in GJS app!');
        });
        
        // Set the application
        this.window.set_application(this.app);
        
        // Show the window
        this.window.present();
        
        print('Senbara GTK window opened successfully!');
    }
    
    run(argv) {
        return this.app.run(argv);
    }
}

// Create and run the application
let app = new SenbaraExampleApp();
app.run(ARGV);