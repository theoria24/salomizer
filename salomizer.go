package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"

	"github.com/jiro4989/ojosama"
	"github.com/joho/godotenv"
	"github.com/mattn/go-mastodon"
)

func normalizeText(str string) string {
	re := regexp.MustCompile("<br( //)?>")
	str = re.ReplaceAllString(str, "\n")
	re = regexp.MustCompile("</p>\n*<p>")
	str = re.ReplaceAllString(str, "\n\n")
	re = regexp.MustCompile("<a.*?</a>")
	str = re.ReplaceAllString(str, "")
	re = regexp.MustCompile(`<("[^"]*"|'[^']*'|[^'">])*>`)
	str = re.ReplaceAllString(str, "")
	return str
}

func main() {
	err := godotenv.Load()

	c := mastodon.NewClient(&mastodon.Config{
		Server:       os.Getenv("MSTDN_SERVER"),
		ClientID:     os.Getenv("MSTDN_CLIENT_ID"),
		ClientSecret: os.Getenv("MSTDN_CLIENT_SECRET"),
		AccessToken:  os.Getenv("MSTDN_ACCESS_TOKEN"),
	})

	wsc := c.NewWSClient()
	q, err := wsc.StreamingWSUser(context.Background())
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Start Watching")
	}
	for e := range q {
		if t, ok := e.(*mastodon.NotificationEvent); ok {
			if t.Notification.Type == "mention" {
				// fmt.Printf("%+v\n", t.Notification.Status.Content)
				cont := html.UnescapeString(normalizeText(t.Notification.Status.Content))
				text, err := ojosama.Convert(cont, nil)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("input: " + cont + "\noutput:" + text)
				_, err = c.PostStatus(context.Background(), &mastodon.Toot{
					Status:      "@" + t.Notification.Status.Account.Acct + " " + text,
					InReplyToID: t.Notification.Status.ID,
					Visibility:  t.Notification.Status.Visibility,
				})
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
