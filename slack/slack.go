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

type Slack struct {
        Webhook string
}

func (s *Slack) NewReview(username string, title string, desc string, url string) error {
        log.Print("Sending new review to slack")
        newAttachment := attachment{
                Title: title,
                TitleLink: url,
                Text: desc,
        }
        attachments := make([]attachment, 1)
        newMessage := message{
                Username: "pull-request",
                Text: fmt.Sprintf("New pull request to review from %s:", username),
                Attachments: append(attachments, newAttachment),
        }
        return s.sendMessage(&newMessage)
}

func (s *Slack) CompleteReview(title string) error {
        newMessage := message{
                Username: "pull-request",
                Text: fmt.Sprintf("Pull request: %s finished review", title),
        }
        return s.sendMessage(&newMessage)
}

func (s *Slack) sendMessage(newMessage *message) error {
        log.Printf("Sending message: %v", newMessage)
        messageBytes, err := json.Marshal(*newMessage)
        if err != nil {
                return err
        }
        resp, err := http.Post(s.Webhook, "application/json", bytes.NewBuffer(messageBytes))
        log.Printf("Got status code: %d from slack, err: %v", resp.StatusCode, err)
        if resp.StatusCode != http.StatusOK {
                return errors.New(fmt.Sprintf("Got invalid status code %d", resp.StatusCode))
        }
        return err
}
