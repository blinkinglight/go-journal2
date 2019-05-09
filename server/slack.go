package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

func checkSlackBotEnabled() bool {
	return os.Getenv("SLACK_TOKEN") == ""
}

var (
	autoPostCh = make(chan string)
)

type slackWebhookMessage struct {
	Text string `json:"text"`
}

func sendToSlack(s string) {
	select {
	case autoPostCh <- s:
	default:
	}
	if *flagWebhook != "" {
		postData := slackWebhookMessage{s}
		b, err := json.Marshal(postData)
		if err != nil {
			log.Printf("json marshal error: %v", err)
		}
		response, err := http.Post(*flagWebhook, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Printf("http post error: %v", err)
		}
		if response.Body != nil {
			response.Body.Close()
		}
	}
}

func startSlackBot() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return
	}

	api := slack.New(token)
	auth, err := api.AuthTest()
	if err != nil {
		panic(err)
	}

	if *flagAutoPost != "" {
		go func() {
			for msg := range autoPostCh {
				log.Printf("sending: %v", msg)
				api.SendMessage(*flagAutoPost, slack.MsgOptionAttachments(slackStringToSlack(msg)))
			}
		}()
	}

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if ev.User != auth.UserID {
				parts := strings.SplitN(ev.Text, " ", 3)
				if len(parts) >= 2 {
					if parts[0] == fmt.Sprintf("<@%s>", auth.UserID) {
						if parts[1] == "today" {
							api.PostMessage(ev.Channel, slack.MsgOptionText("Yesterday notes", false), slack.MsgOptionAttachments(slackGetToday()))
						}
						if parts[1] == "yesterday" {
							api.PostMessage(ev.Channel, slack.MsgOptionText("Yesterday notes", false), slack.MsgOptionAttachments(slackGetYesterday()))
						}
						if len(parts) >= 3 {
							if parts[1] == "last" {
								api.PostMessage(ev.Channel, slack.MsgOptionText("Lastest notes", false), slack.MsgOptionAttachments(slackGetLast(parts[2])))
							}
							if parts[1] == "date" {
								api.PostMessage(ev.Channel, slack.MsgOptionText(parts[2], false), slack.MsgOptionAttachments(slackGetByDate(parts[2])))
							}
						}
					}
				}
			}
		}
	}
}

func slackStringToSlack(s string) slack.Attachment {
	return slack.Attachment{Text: s}
}

func slackGetLast(_limit string) slack.Attachment {

	limit, err := strconv.Atoi(_limit)
	if err != nil {
		limit = 10
	}

	jrs := getRecords(limit)
	text := ""
	for _, jr := range jrs {
		text += fmt.Sprintf("%s %s %s\n", time.Unix(0, jr.ID).UTC().Format("2006-01-02 15:04"), jr.Name, jr.Content)
	}
	return slack.Attachment{Text: text}
}

func slackGetByDate(date string) slack.Attachment {
	jrs := getByDate(date, 1)
	text := ""
	for _, jr := range jrs {
		text += fmt.Sprintf("%s %s %s\n", time.Unix(0, jr.ID).UTC().Format("2006-01-02 15:04"), jr.Name, jr.Content)
	}
	return slack.Attachment{Text: text}
}

func slackGetToday() slack.Attachment {
	dt := time.Now().UTC().Format("2006-01-02")
	jrs := getByDate(dt, 1)
	text := ""
	for _, jr := range jrs {
		text += fmt.Sprintf("%s %s %s\n", time.Unix(0, jr.ID).UTC().Format("2006-01-02 15:04"), jr.Name, jr.Content)
	}
	return slack.Attachment{Pretext: dt, Text: text}
}

func slackGetYesterday() slack.Attachment {
	dt := Day(time.Now().UTC(), -1).UTC().Format("2006-01-02")
	jrs := getByDate(dt, 1)
	text := ""
	for _, jr := range jrs {
		text += fmt.Sprintf("%s %s %s\n", time.Unix(0, jr.ID).UTC().Format("2006-01-02 15:04"), jr.Name, jr.Content)
	}
	return slack.Attachment{Pretext: dt, Text: text}
}
