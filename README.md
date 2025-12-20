# gopher

The Golang-Configured Makefile-like tool that sits on your directory. Run `go build` or other tools while you work.

(My primary use case is to keep running go builds as I fix compliler errors.)

## TODO
- [ ] gopher `bootstrap` command
- [ ] Ctrl + R reset??
- [ ] Improve the output of pretty.Printer
- [ ] Support more gotools
    - [ ] Add more options to those supported
- [ ] Better validate functions in gopher files (better errors)
- [ ] Have the .gopher module be initialized and install runtime dependencies
- [ ] More runner hooks (Init, Run, Close)?
