#!/bin/bash

mkdir -p testdata/buildOutputs

tmpdir=$(mktemp -d)
trap 'rm -rf $tmpdir' EXIT

for tmplfile in testdata/go/*.tmpl; do
    [ -e "$tmplfile" ] || continue
    
    basename=$(basename "$tmplfile" .tmpl)
    gofile="$tmpdir/${basename}.go"
    
    # Copy template to temp dir with .go extension
    cp "$tmplfile" "$gofile"
    
    # Build and capture output
    { go build -json "$gofile" > "testdata/buildOutputs/${basename}.json"; } 2>&1
    echo "$tmplfile -> testdata/buildOutputs/${basename}.json"
done

# TODO: Fix this...
# for gofile in testdata/go/*.go.txt; do
#     [ -e "$gofile" ] || continue
#
#     basename=$(basename "$gofile" .go)
#     tmpdir=$(mktemp -d)
#
#     cp "$gofile" "$tmpdir/main.go"
#
#     pushd "$tmpdir" > /dev/null
#     go mod init testmodule > /dev/null 2>&1
#     go build -json . 2>&1 > "$OLDPWD/testdata/buildOutputs/module_${basename}.json"
#     popd > /dev/null
#
#     rm -rf "$tmpdir"
#
#     echo "$gofile -> testdata/buildOutputs/module_${basename}.json"
# done
