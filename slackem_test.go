package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testSlackWebhookURL = "http://example.com"

func init() {
	slackWebhookURL = testSlackWebhookURL
}

func assertContainsAllRelevantJSON(t *testing.T, body string) {
	assert.Contains(t, body, `"attachments":[`)
	assert.Contains(t, body, `"fallback":"hi there team!"`)
	assert.Contains(t, body, `"color":"#CCCCCC"`)
	assert.Contains(t, body, `"text":"hi there team!"`)
	assert.Contains(t, body, `"mrkdwn_in":["text"]`)
	assert.Contains(t, body, `"channel":"#growth"`)
	assert.Contains(t, body, `"username":"slackem"`)
	assert.Contains(t, body, `"icon_emoji":":rocket:"`)
}

func mockServerWithBodyChannel(code int) (chan []byte, *httptest.Server) {
	done := make(chan []byte, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)

		buf := bytes.NewBuffer(nil)
		io.Copy(buf, r.Body)

		var v interface{}
		err := json.Unmarshal(buf.Bytes(), &v)
		if err != nil {
			panic(err)
		}

		b, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		done <- b
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	slackClient = &http.Client{Transport: transport}

	return done, server
}

func TestPostSlackMessage(t *testing.T) {
	body, server := mockServerWithBodyChannel(200)
	defer server.Close()
	defer close(body)

	postSlackMessage([]string{"growth", "hi", "there", "team!"}, "grey")

	select {
	case json := <-body:
		assertContainsAllRelevantJSON(t, string(json))
	case <-time.After(time.Second * 3):
		t.Fatal("timeout after 3 seconds")
	}
}

func TestSlackPostToChannel(t *testing.T) {
	body, server := mockServerWithBodyChannel(200)
	defer server.Close()
	defer close(body)

	err := slackPostToChannel("#growth", "hi there team!", "grey")
	assert.NoError(t, err)

	select {
	case json := <-body:
		assertContainsAllRelevantJSON(t, string(json))
	case <-time.After(time.Second * 3):
		t.Fatal("timeout after 3 seconds")
	}
}

func TestSlackPostToChannelWithError(t *testing.T) {
	body, server := mockServerWithBodyChannel(404)
	defer server.Close()
	defer close(body)

	err := slackPostToChannel("#growth", "hi there team!", "grey")
	assert.EqualError(t, err, "slack: unexpected response: 404")
}

func TestSlackNewPayload(t *testing.T) {
	json, err := slackNewPayload("#growth", "hi there team!", "grey")
	assert.NoError(t, err)
	assertContainsAllRelevantJSON(t, string(json))
}

func TestSetupConfigs(t *testing.T) {
	// Defaults
	os.Setenv("SLACK_WEBHOOK_URL", "")
	os.Setenv("SLACK_USERNAME", "")
	os.Setenv("SLACK_ICON_EMOJI", "")
	setupConfigs()
	assert.Equal(t, testSlackWebhookURL, slackWebhookURL)
	assert.Equal(t, slackUsername, slackUsername)
	assert.Equal(t, slackIconEmoji, slackIconEmoji)

	// Overrides
	os.Setenv("SLACK_WEBHOOK_URL", "http://fly.akite.com")
	os.Setenv("SLACK_USERNAME", "beefarmer")
	os.Setenv("SLACK_ICON_EMOJI", ":bumblebee:")
	setupConfigs()
	assert.Equal(t, "http://fly.akite.com", slackWebhookURL)
	assert.Equal(t, "beefarmer", slackUsername)
	assert.Equal(t, ":bumblebee:", slackIconEmoji)
}

func TestUsage(t *testing.T) {
	buf := &bytes.Buffer{}
	usage(buf)
	assert.Contains(t, buf.String(), "SLACK_WEBHOOK_URL")
	assert.Contains(t, buf.String(), "SLACK_USERNAME")
	assert.Contains(t, buf.String(), "SLACK_ICON_EMOJI")
}
