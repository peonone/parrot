#!/usr/bin/env bash

BINDIR=app
MODLIST="auth chat user"
BASEPKG=github.com/peonone/parrot

mkdir -p $BINDIR

for mod in $MODLIST; do
    if [ ! -d $mod/apps ]; then 
        continue
    fi

    for comp in $mod/apps/*; do
        comp=`basename $comp`
        echo "building $mod.$comp"
        go build -o $BINDIR/$mod.$comp $BASEPKG/$mod/apps/$comp 
    done
done