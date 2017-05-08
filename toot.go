package main

import (
	"context"
	"log"

	mastodon "github.com/mattn/go-mastodon"
)

func toot(text string) {
	c := mastodon.NewClient(&mastodon.Config{
		Server:       MastodonServer,
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
	})
	err := c.Authenticate(context.Background(), MastodonUser, MastodonPass)
	if err != nil {
		log.Fatal(err)
	}
	toot := mastodon.Toot{Status: text, Visibility: "unlisted"}
	_, err = c.PostStatus(context.Background(), &toot)
	if err != nil {
		log.Fatal(err)
	}
}
