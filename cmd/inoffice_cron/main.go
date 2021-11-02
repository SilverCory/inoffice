package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"office"
	"office/inoffice"
	"os"
	"time"
)

func main() {
	fmt.Println("Running cron to publish!")
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

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		panic("status code non 20x")
	}

	_, _ = io.Copy(os.Stdout, resp.Body)
	fmt.Println("Done!")
}
