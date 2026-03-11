# Legal Notice

> **Disclaimer**: This repository and any associated code are provided "as is" without warranty of any kind, either expressed or implied. The author of this repository does not accept any responsibility for the use or misuse of this repository or its contents. The author does not endorse any actions or consequences arising from the use of this repository. Any copies, forks, or re-uploads made by other users are not the responsibility of the author. The repository is solely intended as a Proof Of Concept for educational purposes regarding the use of a service's private API. By using this repository, you acknowledge that the author makes no claims about the accuracy, legality, or safety of the code and accepts no liability for any issues that may arise. More information can be found [HERE](./LEGAL_NOTICE.md).

# spotify-wrapper

Welcome to spotify-wrapper! This library is designed to interact with the private and public Spotify APIs, emulating the requests typically made through a web browser. This wrapper provides a convenient way to access Spotify's rich set of features programmatically in Go. (Originally ported from Aran404/SpotAPI.)

**Note**: This project is intended solely for educational purposes and should be used responsibly. Accessing private endpoints and scraping data without proper authorization may violate Spotify's terms of service.

## Table of Contents

1. [Introduction](#spotify-wrapper)
2. [Features](#features)
3. [Installation](#installation)
4. [Quick Examples](#quick-examples)
5. [Contributing](#contributing)
7. [License](#license)

## Features
- **No Premium Required**: Unlike the Web API which requires Spotify Premium, **spotify-wrapper** requires no Spotify Premium at all!
- **Public API Access**: Retrieve and manipulate public Spotify data such as playlists, albums, and tracks with ease.
- **Private API Access**: Explore private Spotify endpoints to tailor your application to your needs.
- **Ready to Use**: **spotify-wrapper** is designed for immediate integration, allowing you to accomplish tasks with just a few lines of code.
- **No API Key Required**: Seamlessly use **spotify-wrapper** without needing a Spotify API key.
- **Browser-like Requests**: Accurately replicate the HTTP requests Spotify makes in the browser.
- **Multi-Language Support**: Set your preferred language for API responses using ISO 639-1 language codes (e.g., 'ko', 'ja', 'zh', 'en').

## Installation
```bash
go get github.com/feralbureau/spotify-wrapper
```

## Quick Examples

### With User Authentication
```go
package main

import (
    "log"
    "github.com/feralbureau/spotify-wrapper"
)

func main() {
    client := spotapi.NewLogin("YOUR_EMAIL", "YOUR_PASSWORD")
    err := client.Login()
    if err != nil {
        log.Fatal(err)
    }

    playlist := spotapi.NewPrivatePlaylist(client)
    err = playlist.CreatePlaylist("spotify-wrapper showcase!")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Without User Authentication
```go
package main

import (
    "fmt"
    "log"
    "github.com/feralbureau/spotify-wrapper"
)

func main() {
    song := spotapi.NewSong()
    results, err := song.QuerySongs("weezer", 20)
    if err != nil {
        log.Fatal(err)
    }

    for idx, item := range results {
        fmt.Printf("%d %s\n", idx, item.Name)
    }
}
```

### With Language Support
```go
package main

import (
    "github.com/feralbureau/spotify-wrapper"
)

func main() {
    artist := spotapi.NewArtist("ko") // Korean
    playlist := spotapi.NewPublicPlaylist("37i9dQZF1DXcBWIGoYBM5M", "ko")
    song := spotapi.NewSong("en")
    song.SetLanguage("ja") // Switch to Japanese
}
```

## Contributing
Contributions are welcome! If you find any issues or have suggestions, please open an issue or submit a pull request.

## License
This project is licensed under the **GPL 3.0** License. See [LICENSE](https://choosealicense.com/licenses/gpl-3.0/) for details.

