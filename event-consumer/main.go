package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"event-router-auth/internal/model"
)

func main() {
	http.HandleFunc("/internal-event", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event model.Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid event", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Printf("[event-consumer] ✅ Received event: %+v", event)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "event processed")
	})

	log.Println("[event-consumer] 🚀 Listening on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
