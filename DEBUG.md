# debug

## debug binary size

* check symbol size
```bash
# filter out undefined symbols
go tool nm -size mythosd | grep -v ' U ' | sort -nr | head -20
```

* size profiling https://github.com/Zxilly/go-size-analyzer
```bash
brew install go-size-analyzer
# or
go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest

cd ./build
gsa --web mythosd
# or
gsa --tui mythosd
gsa mythosd
```

* included packages
```bash
go list -deps -f '{{.ImportPath}}' ./...

go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./...

```

* dependency chain
```bash
go mod graph
go mod why <module>
```
