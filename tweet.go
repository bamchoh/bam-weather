package main

import "github.com/ChimeraCoder/anaconda"

func tweet(text string) error {
	anaconda.SetConsumerKey(ConsumerKey)
	anaconda.SetConsumerSecret(ConsumerSecret)
	api := anaconda.NewTwitterApi(APIKey, APISecret)
	_, err := api.PostTweet(text, nil)
	if err != nil {
		return err
	}
	return nil
}
