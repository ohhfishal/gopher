#!/bin/bash

mkdir -p testdata/buildOutputs

for gofile in testdata/go/*.go.txt; do
    [ -e "$gofile" ] || continue
    
    basename=$(basename "$gofile" .go)
    go build -json "$gofile" 2>&1 > "testdata/buildOutputs/${basename}.json"
    echo "$gofile -> testdata/buildOutputs/${basename}.json"
done

for gofile in testdata/go/*.go.txt; do
    [ -e "$gofile" ] || continue
    
    basename=$(basename "$gofile" .go)
    tmpdir=$(mktemp -d)
    
    cp "$gofile" "$tmpdir/main.go"
    
    pushd "$tmpdir" > /dev/null
    go mod init testmodule > /dev/null 2>&1
    go build -json . 2>&1 > "$OLDPWD/testdata/buildOutputs/module_${basename}.json"
    popd > /dev/null
    
    rm -rf "$tmpdir"
    
    echo "$gofile -> testdata/buildOutputs/module_${basename}.json"
done
