package twtichauth

import (
	"agones-minecraft/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

func NewTwitchConfig() *oauth2.Config {
	id, sec, re := config.GetTwichCreds()
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: sec,
		Endpoint:     twitch.Endpoint,
		RedirectURL:  re,
	}
}
