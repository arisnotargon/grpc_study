package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var logger *log.Logger

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
	logger.Printf("helloworld: received a request [%+v], content: [%s]\n", r.Body, reqContent)
	target := os.Getenv("TARGET")
	if target == "" {
		target = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)
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
