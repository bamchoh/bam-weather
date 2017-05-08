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
	ConsumerKey    = "XXXXXXXXXXXXXXXXXXXX"
	ConsumerSecret = "YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY"
	APIKey         = "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
	APISecret      = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
)
```

```
$ go build
$ bam-weather
```
