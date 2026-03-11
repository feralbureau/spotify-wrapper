# SpotAPI-Go

A Go port of the SpotAPI Python library.

## Features
- Ported from the original Python library.
- Uses `tls-client` for fingerprinting.
- Supports Private and Public APIs.

## Installation
```bash
go get github.com/spotapi/spotapi-go
```

## Example
```go
package main

import (
    "github.com/spotapi/spotapi-go/pkg/spotapi"
    "github.com/spotapi/spotapi-go/internal/http"
    "github.com/spotapi/spotapi-go/internal/types"
    tls_client "github.com/bogdanfinn/tls-client"
)

func main() {
    client, _ := http.NewClient(tls_client.HelloChrome_120, "", 3)
    cfg := &types.Config{
        Client: client,
    }

    // Use the library
    playlist := spotapi.NewPublicPlaylist("37i9dQZF1DXcBWIGoYBM5M", client, "en")
    info, _ := playlist.GetPlaylistInfo(20, 0)
    // ...
}
```
