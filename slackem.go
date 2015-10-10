package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/nlopes/slack"
)

// Configuration vars.
var (
	slackWebhookURL = ""
	slackUsername   = "slackem"
	slackIconEmoji  = ":rocket:"
)

var slackClient *http.Client

// Possible colors for the notification.
var hexColors = map[string]string{
	"grey":  "#CCCCCC",
	"red":   "#BB0000",
	"green": "#7CD197",
	"blue":  "#103FFB",
}

type slackPayload struct {
	Channel     string             `json:"channel"`
	Username    string             `json:"username"`
	IconEmoji   string             `json:"icon_emoji"`
	Attachments []slack.Attachment `json:"attachments"`
}

func slackNewAttachment(message, color string) []slack.Attachment {
	return []slack.Attachment{{
		Fallback:   message,
		Text:       message,
		Color:      hexColors[color],
		MarkdownIn: []string{"text"},
	}}
}

func slackNewPayload(channel, message, color string) ([]byte, error) {
	return json.Marshal(slackPayload{
		Channel:     channel,
		Username:    slackUsername,
		IconEmoji:   slackIconEmoji,
		Attachments: slackNewAttachment(message, color),
	})
}

func slackPostToChannel(channel, message, color string) error {
	payload, _ := slackNewPayload(channel, message, color)
	if slackClient == nil {
		slackClient = http.DefaultClient
	}
	resp, err := slackClient.Post(slackWebhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Println("[slack] failed to notify:", err)
		return err
	}
	if resp.StatusCode >= 201 {
		log.Println("[slack] unexpected response:", resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			log.Println("[slack] couldn't read response body:", err)
			return err
		}
		log.Println("[slack] response:", string(body))
		return fmt.Errorf("slack: unexpected response: %d", resp.StatusCode)
	}
	return nil
}

func postSlackMessage(args []string, color string) {
	err := slackPostToChannel("#"+args[0], strings.Join(args[1:], " "), color)
	if err != nil {
		log.Println("[slack] error!", err)
	}
}

func setupConfigs() {
	if url := os.Getenv("SLACK_WEBHOOK_URL"); url != "" {
		slackWebhookURL = url
	}
	if username := os.Getenv("SLACK_USERNAME"); username != "" {
		slackUsername = username
	}
	if emoji := os.Getenv("SLACK_ICON_EMOJI"); emoji != "" {
		slackIconEmoji = emoji
	}
}

func usage(writer io.Writer) {
	fmt.Fprintf(writer, "usage: %s channel Type your message...\n", path.Base(os.Args[0]))
	fmt.Fprint(writer, "configuration: all through environment vars.\n\n")
	fmt.Fprint(writer, "    SLACK_WEBHOOK_URL - your incoming webhook url, required\n")
	fmt.Fprint(writer, "    SLACK_USERNAME    - the username for who sent the message, defaults to slackem\n")
	fmt.Fprint(writer, "    SLACK_ICON_EMOJI  - the emoji icon to use, defaults to :rocket:\n")
}

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "fatal: %s\n", msg)
	usage(os.Stderr)
	os.Exit(1)
}

func main() {
	var color string
	flag.StringVar(&color, "color", "grey", "The color on the left of the message.")
	flag.Parse()

	setupConfigs()
	if slackWebhookURL == "" {
		fatal("you must provide an incoming webhook url")
	}

	log.Println(os.Args)
	log.Println(color)
	if len(os.Args) < 3 {
		fatal("not enough args")
	}

	args := []string{}
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			args = append(args, arg)
		}
	}
	postSlackMessage(args, color)
}
