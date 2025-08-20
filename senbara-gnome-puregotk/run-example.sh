#!/bin/bash

# Set up the environment for the Senbara GTK library
export GI_TYPELIB_PATH="$HOME/.local/lib/girepository-1.0:$GI_TYPELIB_PATH"
export LD_LIBRARY_PATH="$HOME/.local/lib:$LD_LIBRARY_PATH"

# Run the example GJS app
exec ./example.js "$@"