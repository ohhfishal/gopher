# gopher

Better `go build` output.

## Usuage

```bash
go build -json | ./gopher
```


### Output

```
package: github.com/ohhfishal/gopher/watch
report.go
  undefined:
    errors (2:22)
    fmt (5:9), (5:10), ...
    io (2:58), (4:48), ...
  Did you forget to import ("errors", "fmt", "io")?
	 ...
FAILED
```
