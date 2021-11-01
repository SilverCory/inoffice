package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"office"
	"office/inoffice"
	"time"
)

func main() {
	var env = office.GetEnv()

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	var weekStart = inoffice.NextWeek(time.Now())
	var msg = inoffice.BuildInOfficeMessage(weekStart, make(map[inoffice.Day][]inoffice.InOffice))

	msg.Channel = env.SlackMainChannel

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(msg); err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://slack.com/api/chat.postMessage", buf)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer "+env.SlackBotToken)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	_, err = httpClient.Do(req)
	if err != nil {
		panic(err)
	}
}
