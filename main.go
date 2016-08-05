package main

import (
        "net/http"
        "log"
        "encoding/json"
        "strings"
        "github.com/beh9540/slackpull/slack"
)

type PullRequest struct {
        IssueUrl string `json:"issue_url"`
        Title string    `json:"title"`
        Body string     `json:"body"`
        Url string      `json:"url"`
        HtmlUrl string  `json:"html_url"`
}

type WebHook struct {
        Action string                   `json:"action"`
        PullRequest *PullRequest        `json:"pull_request"`
}

type Label struct {
        Url string
        Name string
        Color string
}

type Issue struct {
         Labels []Label `json:"labels"`
 }

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
        event := r.Header.Get("X-GitHub-Event")
        log.Printf("Got event: %s", event)
        var webHook WebHook
        err := json.NewDecoder(r.Body).Decode(&webHook)
        if err != nil {
                log.Printf("Got error with process pullrequest json: %v", err)
                w.WriteHeader(http.StatusInternalServerError)
                return
        }
        switch event {
        case "pull_request":
                err = processPullRequest(&webHook)
                if err != nil {
                        log.Printf("Got error with getting issue: %v", err)
                        w.WriteHeader(http.StatusInternalServerError)
                        return
                }
        }
        w.WriteHeader(http.StatusOK)
}

func processPullRequest(webhook *WebHook) error {
        var action string
        action = webhook.Action
        log.Printf("Got action: %s", action)
        log.Printf("Got webhook: %s", webhook)
        switch action {
        case "labeled":
                log.Printf("New Pull Request: %v", webhook.PullRequest)
                pullRequest := webhook.PullRequest
                resp, err := http.Get(pullRequest.IssueUrl)
                if err != nil {
                        return err
                }
                defer resp.Body.Close()
                var issue Issue
                if err = json.NewDecoder(resp.Body).Decode(&issue); err != nil {
                        return err
                }
                labels := issue.Labels
                log.Printf("Got labels: %v", labels)
                for _, label := range labels {
                        switch {
                        case strings.Contains(label.Name, "ready for review"):
                                return slack.NewReview(pullRequest.Title, pullRequest.Body, pullRequest.HtmlUrl)
                        case strings.Contains(label.Name, "has been reviewed"):
                                return slack.CompleteReview(pullRequest.Title)
                        }
                }
        }
        return nil
}

func main() {
        http.HandleFunc("/process", WebhookHandler)
        log.Fatal(http.ListenAndServe(":8080", nil))
}