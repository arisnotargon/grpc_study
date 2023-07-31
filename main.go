package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
)

var logger *log.Logger

type PubSubMessage struct {
	Message struct {
		// data	如果定义为[]byte的话就会自动解析收到的base64
		Data []byte `json:"data,omitempty"`
		// Data string `json:"data,omitempty"`
		ID string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type MessageBody struct {
	Endpoint string `json:"endpoint,omitempty"`
	Headers  []struct {
		Key   string `json:"key,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"headers,omitempty"`
	Body []byte `json:"body,omitempty"`
}

const (
	webhookTopicName    = "projects/horseman159753/subscriptions/test-topic-cloudrun-sub"
	deadLetterTopicName = "projects/horseman159753/subscriptions/test-dead-mq-sub"
)

const (
	slackWebhook = "https://hooks" + "slack.com" + "/services/" + "T05KF554TPD" + "/B05L5D4H7BJ" + "/JnzpTXg8x1AO7DFpJQoFoItb"
)

func init() {
	writer, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE, 0755)
	writer2 := os.Stdout
	if err != nil {
		log.Fatalf("create file log.txt failed: %v", err)
	}
	logger = log.New(io.MultiWriter(writer, writer2), "", log.Lshortfile|log.LstdFlags)
}

func handler(w http.ResponseWriter, r *http.Request) {
	reqContent := make([]byte, r.ContentLength)
	_, _ = r.Body.Read(reqContent)

	pm := &PubSubMessage{}

	err := json.Unmarshal(reqContent, pm)
	if err != nil {
		logger.Fatalf("unmarshal_pubsubmessage_error,err : [%v], reqContent:%s", err, reqContent)
		w.WriteHeader(500)
		fmt.Fprintf(w, "unmarshal_pubsubmessage_error,err : %s\n", err.Error())
		return
	}

	logger.Printf("pm:[%v]", pm)

	switch pm.Subscription {
	case webhookTopicName:
		// test no ack
		logger.Printf("in normal webhookTopicName, sub name : %s, msg id: %s", pm.Subscription, pm.Message.ID)
		w.WriteHeader(500)
		fmt.Fprintf(w, "test no ack")
	case deadLetterTopicName:
		logger.Printf("in deadLetterTopicName, PubSubMessage:[%v]", pm)
		restyReq := resty.New().R()
		restyReq.SetHeader("Content-type", "application/json").
			SetBody([]byte(
				fmt.Sprintf(
					`{"text":"subname:%s,
			messege body :%s,
			 messege id: %s"}'`,
					pm.Subscription,
					base64.StdEncoding.EncodeToString(pm.Message.Data),
					pm.Message.ID,
				),
			))

		response, err := restyReq.Post(slackWebhook)

		logger.Printf("resty response:[%v], resty error:[%v]", response, err)

		fmt.Fprintf(w, "ack")
	default:
		logger.Fatalf("pm.Subscription error:%s", pm.Subscription)
		w.WriteHeader(500)
		fmt.Fprintf(w, "pm.Subscription error:%s", pm.Subscription)
	}
}

func main() {
	logger.Print("helloworld: starting server...")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Printf("helloworld: listening on port %s", port)
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
