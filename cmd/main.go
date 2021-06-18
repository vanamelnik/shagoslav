package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = 3000

func meetingServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<html><head><title>Шагослав meeting</title></head><body>Это не так, <strong>Шагослав!</strong></body></html>")
}

func main() {
	log.Printf("Listening at %d...", port)
	http.HandleFunc("/", meetingServer)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
