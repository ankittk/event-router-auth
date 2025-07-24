package dispatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"event-router-auth/internal/model"
)

type Forwarder interface {
	Forward(event model.Event) error
}

type httpForwarder struct {
	logger   *log.Logger
	client   *http.Client
	endpoint string
}

func NewForwarder(logger *log.Logger) Forwarder {
	return &httpForwarder{
		logger:   logger,
		client:   &http.Client{Timeout: 5 * time.Second},
		endpoint: "http://consumer:9090/internal-event", // Mock internal service
	}
}

func (f *httpForwarder) Forward(event model.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, f.endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create forward request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("send forward request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downstream returned status: %s", resp.Status)
	}

	f.logger.Printf("Forwarded event to %s", f.endpoint)
	return nil
}
