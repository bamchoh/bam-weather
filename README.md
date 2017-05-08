# bam-weather
This is Tweet Bot that tweets weather forecast in Osaka.

# Usage
```
$ git clone https://github.com/bamchoh/bam-weather
```

set 2 key/secret pare in main package as const variables
```
package main

const (
	ConsumerKey    = "XXXX"
	ConsumerSecret = "YYYY"
	APIKey         = "ZZZZ"
	APISecret      = "AAAA"
	ClientID       = "BBBB"
	ClientSecret   = "CCCC"
	MastodonServer = "https://mstdn.jp"
	MastodonUser   = "DDDD"
	MastodonPass   = "EEEE"
)
```

```
$ go build
$ bam-weather
```
