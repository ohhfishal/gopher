# gopher

The Golang-Configured Makefile-like tool that sits on your directory. Run `go build` or other tools while you work.

(My primary use case is to keep running go builds as I fix compliler errors.)

## Getting Started

```
# Alternatively use go install
go get -tool github.com/ohhfishal/gopher

# Confirm installation
go tool gopher version

# Get a sample gopher.go file (Akin to makefile)
wget https://raw.githubusercontent.com/ohhfishal/gopher/refs/heads/main/example/default.go -O gopher.go

# Run the hello target to confirm everything is configured
go tool gopher hello
```

After that point, open `gopher.go` and add/edit targets as desired.


## TODO
- [ ] gopher `bootstrap` command
- [ ] Ctrl + R reset??
- [ ] Improve the output of pretty.Printer
- [ ] Support more gotools
    - [ ] Add more options to those supported
- [ ] Better validate functions in gopher files (better errors)
- [ ] Have the .gopher module be initialized and install runtime dependencies
- [ ] More runner hooks (Init, Run, Close)?
