package main

import (
	"encoding/json"
	"github.com/beh9540/slackpull/config"
	"github.com/beh9540/slackpull/slack"
	"log"
	"net/http"
	"os"
	"strings"
)

type PullRequest struct {
	IssueUrl string `json:"issue_url"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Url      string `json:"url"`
	HtmlUrl  string `json:"html_url"`
	User     *User  `json:"user"`
}

type User struct {
	Login string `json:"login"`
}

type Repo struct {
	FullName string `json:"full_name"`
}

type WebHook struct {
	Action      string       `json:"action"`
	PullRequest *PullRequest `json:"pull_request"`
	Repo        *Repo        `json:"repository"`
}

type Label struct {
	Url   string
	Name  string
	Color string
}

type Issue struct {
	Labels []Label `json:"labels"`
}

var pullconfig *config.Config

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

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var repoConfig config.RepoConfig
		if err := json.NewDecoder(r.Body).Decode(&repoConfig); err != nil {
			log.Printf("Got error unmarshelling json: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err := pullconfig.Upsert(repoConfig)
		if err != nil {
			log.Printf("Got error upserting config: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		break
	case "GET":
		config := pullconfig.Get()
		out, err := json.Marshal(config)
		if err != nil {
			log.Printf("Got error marshelling json: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(out)
		break
	default:
		w.WriteHeader(http.StatusNotFound)
		break
	}
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
		webhookConfigs := pullconfig.Get()
		slackInst := slack.Slack{
			Webhook: webhookConfigs[webhook.Repo.FullName],
		}
		for _, label := range labels {
			switch {
			case strings.Contains(label.Name, "ready for review"):
				return slackInst.NewReview(pullRequest.User.Login, pullRequest.Title, pullRequest.Body,
					pullRequest.HtmlUrl)
			case strings.Contains(label.Name, "has been reviewed"):
				return slackInst.CompleteReview(pullRequest.Title)
			}
		}
	}
	return nil
}

func main() {
	pullconfig = &config.Config{
		SqlString: os.Args[1],
	}
	pullconfig.Load()
	http.HandleFunc("/process", WebhookHandler)
	http.HandleFunc("/config", ConfigHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
