package inoffice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"office"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type server struct {
	store  Store
	env    office.Env
	client *http.Client
}

func StartServer(store Store, env office.Env) {
	var s = &server{
		store:  store,
		env:    env,
		client: &http.Client{Timeout: time.Second * 5},
	}

	http.HandleFunc("/command-inoffice", s.handlerInOffice)
	http.HandleFunc("/interaction", s.handlerInteraction)

	fmt.Println("Server listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func (s *server) handlerInOffice(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, s.env.SlackSigningSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Unable to create secret verifier: %s\n", err)
		return
	}

	defer r.Body.Close()

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
	_, err = slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	if err = verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Printf("Unauthorized request: %s\n", err)
		return
	}

	weekStart := StartOfWeek(time.Now())

	allForWeek, err := s.store.GetAllForWeek(weekStart)
	if err != nil {
		msg := BuildErrorMessage(fmt.Errorf("unable to get entries: %w", err))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			fmt.Printf("Unable to encode json: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	msg := BuildInOfficeMessage(weekStart, allForWeek)
	msg.ResponseType = slack.ResponseTypeInChannel
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		fmt.Printf("Unable to encode json: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *server) handlerInteraction(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, s.env.SlackSigningSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Unable to create secret verifier: %s\n", err)
		return
	}

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))

	var payload slack.InteractionCallback
	if err = json.Unmarshal([]byte(r.FormValue("payload")), &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("Unmarshal json payload failed: %s\n", err)
		return
	}

	if err = verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Printf("Unauthorized request: %s\n", err)
		return
	}

	// Ignore not block actions.
	if payload.Type != slack.InteractionTypeBlockActions {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("Payload type is not InteractionTypeBlockActions: %s\n", payload.Type)
		return
	}

	action := payload.ActionCallback.BlockActions[0]
	var inOffice = InOffice{
		UserID:   payload.User.ID,
		Username: payload.User.Name,
		InOn:     Day(strings.ToUpper(action.Value)),
	}

	actionID := action.ActionID
	weekStartStrs := strings.Split(actionID, "-toggle")
	if len(weekStartStrs) != 2 {
		err = fmt.Errorf("Payload ActionID not {timestamp}-toggle: %s\n", actionID)
		if err := s.httpResponse(BuildErrorMessage(err), payload.ResponseURL); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Unable to respond: %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	weekStartInt, err := strconv.ParseInt(weekStartStrs[0], 10, 64)
	if err != nil {
		err = fmt.Errorf("ParseInt week start string failed (%s): %v\n", weekStartStrs[0], err)
		if err := s.httpResponse(BuildErrorMessage(err), payload.ResponseURL); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Unable to respond: %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// Add 30 for shits and giggles if it gets down to ~1 second and we're using 00:00:00.0000001 could be fun.
	inOffice.WeekStart = StartOfWeek(time.Unix(weekStartInt+30, 0))

	// Check we're booking into the future.
	if IsInPast(inOffice.WeekStart, inOffice.InOn) {
		if err := s.httpResponse(BuildErrorMessage(fmt.Errorf("invalid date (in the past or after 4PM today)")), payload.ResponseURL); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Unable to respond: %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := s.store.Save(inOffice); err != nil {
		fmt.Printf("Unable to save entry: %v\n", err)
		if err := s.httpResponse(BuildErrorMessage(fmt.Errorf("unable to save entry: %w", err)), payload.ResponseURL); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Unable to respond: %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	allForWeek, err := s.store.GetAllForWeek(inOffice.WeekStart)
	if err != nil {
		if err := s.httpResponse(BuildErrorMessage(err), payload.ResponseURL); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("Unable to respond: %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	msg := BuildInOfficeMessage(inOffice.WeekStart, allForWeek)
	msg.ReplaceOriginal = true

	if err := s.httpResponse(msg, payload.ResponseURL); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Unable to respond: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) httpResponse(msg slack.Message, responseURL string) error {
	var buf = new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, responseURL, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	_, err = s.client.Do(req)
	return err
}
