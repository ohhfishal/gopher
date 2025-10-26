package watch

// From https://pkg.go.dev/cmd/go#hdr-Build__json_encoding
type BuildEvent struct {
	// TODO: Get the import path using go list -json. Then use that to truncate this one
	ImportPath string
	Action     string
	Output     string

	// The Action field is one of the following:
	// build-output - The toolchain printed output
	// build-fail - The build failed
}
