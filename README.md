## mapcomments-go

A CLI tool that extracts comments from Go files and generates a Go file with them as a map.

### Installation:

```
go install github.com/antoniszymanski/mapcomments-go@latest
```

### Usage:

```
Usage: mapcomments <entries> ... [flags]

Arguments:
  <entries> ...

Flags:
  -h, --help                          Show context-sensitive help.
  -P, --package="main"
  -W, --output="commentmap_gen.go"
      --with-full-comment

mapcomments: error: expected "<entries> ..."
```
