# gopher

The Golang-Configured Makefile-like tool that sits on your directory. Run `go build` or other tools while you work.

(My primary use case is to keep running go builds as I fix compliler errors.)

## TODO
- [ ] Improve the output of pretty.Printer
- [ ] Support more gotools
    - [ ] Add more options to those supported
- [ ] Better validate functions in gopher files (better errors)
- [ ] Write logs to files since stdout gets cleared
- [ ] Have the .gopher module be initialized and install runtime dependencies
