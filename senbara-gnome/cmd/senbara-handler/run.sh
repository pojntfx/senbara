#!/bin/bash

set -ex

go install -x ./senbara-gnome/cmd/senbara-handler/ && senbara-handler
