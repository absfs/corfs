# CorFS - Cache on Read FileSystem

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/absfs/corfs/blob/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/corfs.svg)](https://pkg.go.dev/github.com/absfs/corfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/absfs/corfs)](https://goreportcard.com/report/github.com/absfs/corfs)
[![Test](https://github.com/absfs/corfs/actions/workflows/test.yml/badge.svg)](https://github.com/absfs/corfs/actions/workflows/test.yml)

The `corfs` package implements a Cache-on-Read FileSystem that wraps two `absfs.Filer` implementations. It reads from a primary filesystem and caches content to a secondary filesystem on successful reads, providing a two-tier caching system.

## Features

- **Two-tier caching**: Reads from primary, caches to secondary
- **Transparent operation**: Acts as a standard `absfs.Filer`
- **Best-effort caching**: Cache failures don't affect primary operations

## Install

```bash
go get github.com/absfs/corfs
```

## Example Usage

```go
package main

import (
    "os"

    "github.com/absfs/corfs"
    "github.com/absfs/memfs"
    "github.com/absfs/osfs"
)

func main() {
    // Create primary filesystem (e.g., slow remote storage)
    primary, _ := osfs.NewFS()

    // Create cache filesystem (e.g., fast local memory)
    cache, _ := memfs.NewFS()

    // Create cache-on-read filesystem
    fs := corfs.New(primary, cache)

    // First read comes from primary and caches to secondary
    f, _ := fs.OpenFile("/data/file.txt", os.O_RDONLY, 0)
    defer f.Close()

    // Subsequent reads can be served from cache
    // (if using a separate instance of the same cache)
}
```

## absfs

Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract filesystem interface and features like filesystem composition.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/corfs/blob/master/LICENSE)
