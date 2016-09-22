# Gobbler

A simple [Last.fm](http://www.lastfm.com)] scrobbler in Go

## Functionality

This is a very simple library that only allows you to authenticate
and scrobble songs.

## Import

```
import "github.com/honzab/gobbler"
```

## Usage

```
gobbler := Gobbler{
    ApiKey: "<your_api_key>",
    Secret: "<your_api_secret",
}
// Error handling removed for simplicity
success, _ := gobbler.Login("<your_username>", "<your_password>")
r, _ := gobbler.Scrobble("Röyksopp", "What else is there?", "The Understanding")
```

