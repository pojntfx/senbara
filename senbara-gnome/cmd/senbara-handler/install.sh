#!/bin/bash

set -ex

cp senbara-gnome/cmd/senbara-handler/com.pojtinger.felicitas.SenbaraHandler.desktop ~/.local/share/applications
update-desktop-database ~/.local/share/applications
