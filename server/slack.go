package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"os"
	"strconv"
	"strings"
	"time"
)

func checkSlackBotEnabled() bool {
	return os.Getenv("SLACK_TOKEN") == ""
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
	jrs := getByDate(date)
	text := ""
	for _, jr := range jrs {
		text += fmt.Sprintf("%s %s %s\n", time.Unix(0, jr.ID).UTC().Format("2006-01-02 15:04"), jr.Name, jr.Content)
	}
	return slack.Attachment{Text: text}
}

func slackGetToday() slack.Attachment {
	dt := time.Now().UTC().Format("2006-01-02")
	jrs := getByDate(dt)
	text := ""
	for _, jr := range jrs {
		text += fmt.Sprintf("%s %s %s\n", time.Unix(0, jr.ID).UTC().Format("2006-01-02 15:04"), jr.Name, jr.Content)
	}
	return slack.Attachment{Pretext: dt, Text: text}
}

func slackGetYesterday() slack.Attachment {
	dt := Day(time.Now().UTC(), -1).UTC().Format("2006-01-02")
	jrs := getByDate(dt)
	text := ""
	for _, jr := range jrs {
		text += fmt.Sprintf("%s %s %s\n", time.Unix(0, jr.ID).UTC().Format("2006-01-02 15:04"), jr.Name, jr.Content)
	}
	return slack.Attachment{Pretext: dt, Text: text}
}
