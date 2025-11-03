#!/bin/bash

interactive=false
if [ "$1" = "-i" ]; then
    interactive=true
fi

for jsonfile in testdata/buildOutputs/*.json; do
    [ -e "$jsonfile" ] || continue
    
    echo ""
    echo "================================================================================"
    echo "File: $jsonfile"
    echo "================================================================================"
    echo ""
    echo "Input JSON:"
    echo "--------------------------------------------------------------------------------"
    cat "$jsonfile"
    echo ""
    echo "--------------------------------------------------------------------------------"
    echo "Report Output:"
    echo "--------------------------------------------------------------------------------"
    cat "$jsonfile" | go run . report
    echo ""
    echo "================================================================================"
    echo ""
    
    if [ "$interactive" = true ]; then
        read -r -p "Press Enter to continue to next file..."
        echo ""
    fi
done
