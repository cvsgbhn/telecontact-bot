package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

/* send logs to bot */
func TelegramLogs(numa []string, numb []string, callSessionId []string, startTime string, answer string, err_log string) {
	token := os.Getenv("TELEGRAM_TOKEN")
	chatId := os.Getenv("CHAT_ID")
	var allLogs string
	if err_log != "" {
		allLogs = " numa: " + numa[0] + "\nnumb: " + numb[0] + "\ncall_session_id: " + callSessionId[0] + "\nstart_time: " + startTime + "\nreturned_code: " + answer + "\nerror: " + err_log
	} else {
		allLogs = " numa: " + numa[0] + "\nnumb: " + numb[0] + "\ncall_session_id: " + callSessionId[0] + "\nstart_time: " + startTime + "\nreturned_code: " + answer
	}

	Url, err := url.Parse("https://api.telegram.org")
	if err != nil {
		logrus.Fatal("TelegramLogs: ", err)
		panic(err)
	}

	Url.Path += "/bot" + token + "/sendMessage"
	parameters := url.Values{}
	parameters.Add("chat_id", chatId)
	parameters.Add("text", allLogs)
	Url.RawQuery = parameters.Encode()

	req, err := http.NewRequest("POST", Url.String(), nil)
	if err != nil {
		logrus.Fatal("TelegramLogs: ", err)
	}

	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Fatal("TelegramLogs: ", err)
	}
	defer resp.Body.Close()
}

/* GET phone call */
func getPhoneCall(w http.ResponseWriter, r *http.Request) {
	/* structure for response json */
	type Answer struct {
		ReturnedCode int `json:"returned_code"`
	}

	/* parse query params */
	numa, ok := r.URL.Query()["numa"]
	numb, ok := r.URL.Query()["numb"]
	callSessionId, ok := r.URL.Query()["call_session_id"]
	startTime, ok := r.URL.Query()["start_time"]

	answer := Answer{}

	/* check if time exists */
	if !ok || len(startTime[0]) < 0 {
		answer = Answer{ReturnedCode: 1}
		jsonBytes, err := json.Marshal(answer)
		if err != nil {
			logrus.Fatal("ListenAndServe: ", err)
			TelegramLogs(numa, numb, callSessionId, "", strconv.Itoa(answer.ReturnedCode), err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonBytes)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonBytes)
		return
	}

	/* parse time from unix to utc */
	parsedTime, err := strconv.ParseInt(startTime[0], 10, 64)
	if err != nil {
		answer = Answer{ReturnedCode: 1}
		TelegramLogs(numa, numb, callSessionId, strconv.Itoa(int(parsedTime)), strconv.Itoa(answer.ReturnedCode), err.Error())
		jsonBytes, err := json.Marshal(answer)
		if err != nil {
			logrus.Fatal("ListenAndServe: ", err)
			TelegramLogs(numa, numb, callSessionId, strconv.Itoa(int(parsedTime)), strconv.Itoa(answer.ReturnedCode), err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonBytes)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonBytes)
		panic(err)
	}
	tm := time.Unix(parsedTime, 0)

	/* check time interval (hours) */
	if tm.Hour()+3 >= 10 && tm.Hour()+3 < 22 {
		answer = Answer{ReturnedCode: 1}
	} else {
		answer = Answer{ReturnedCode: 0}
	}
	jsonBytes, err := json.Marshal(answer)

	/* send logs */
	TelegramLogs(numa, numb, callSessionId, tm.String(), strconv.Itoa(answer.ReturnedCode), "")
	logrus.WithFields(logrus.Fields{
		"project":         "detect time interval",
		"package":         "main",
		"function":        "detect-moscow-time.getPhoneCall",
		"numa":            numa,
		"numb":            numb,
		"call_session_id": callSessionId,
		"start_time":      startTime,
		"answer":          answer,
	})

	/* return response */
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

/* get port from environment */
func getPort() string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

func main() {
	http.HandleFunc("/", getPhoneCall)
	err := http.ListenAndServe(getPort(), nil)
	if err != nil {
		logrus.Fatal("ListenAndServe: ", err)
	}
}
