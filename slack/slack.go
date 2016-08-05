package slack

import (
        "net/http"
        "fmt"
        "encoding/json"
        "errors"
        "log"
        "bytes"
)

type attachment struct {
        Title string            `json:"title"`
        TitleLink string        `json:"title_link"`
        Text string             `json:"text"`
}

type message struct {
        Username string                 `json:"user_name"`
        IconEmoji string                `json:"icon_emoji"`
        Text string                     `json:"text"`
        Attachments []attachment        `json:"attachments"`
}

//TODO (beh) this needs to be configurable
var slackWebhookUrl string = "https://hooks.slack.com/services/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"

func NewReview(title string, desc string, url string) error {
        log.Print("Sending new review to slack")
        newAttachment := attachment{
                Title: title,
                TitleLink: url,
                Text: desc,
        }
        attachments := make([]attachment, 1)
        newMessage := message{
                Username: "pull-request",
                Text: "New pull request to review:",
                Attachments: append(attachments, newAttachment),
        }
        return sendMessage(&newMessage)
}

func CompleteReview(title string) error {
        newMessage := message{
                Username: "pull-request",
                Text: fmt.Sprintf("Pull request: %s finished review", title),
        }
        return sendMessage(&newMessage)
}

func sendMessage(newMessage *message) error {
        log.Printf("Sending message: %v", newMessage)
        messageBytes, err := json.Marshal(*newMessage)
        if err != nil {
                return err
        }
        resp, err := http.Post(slackWebhookUrl, "application/json", bytes.NewBuffer(messageBytes))
        log.Printf("Got status code: %d from slack, err: %v", resp.StatusCode, err)
        if resp.StatusCode != http.StatusOK {
                return errors.New(fmt.Sprintf("Got invalid status code %d", resp.StatusCode))
        }
        return err
}
