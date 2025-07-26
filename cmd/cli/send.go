package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	payloadPath string
	token       string
	url         string
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a CLI event to the event-router",
	Run: func(cmd *cobra.Command, args []string) {
		if token == "" {
			token = os.Getenv("VALID_PAT")
		}
		if token == "" {
			fmt.Println("Missing token. Use --token or set VALID_PAT.")
			os.Exit(1)
		}

		if payloadPath == "" {
			fmt.Println("Missing --payload flag.")
			os.Exit(1)
		}

		payload, err := os.ReadFile(payloadPath)
		if err != nil {
			fmt.Printf("Failed to read payload: %v\n", err)
			os.Exit(1)
		}

		if !json.Valid(payload) {
			fmt.Println("Invalid JSON in payload file.")
			os.Exit(1)
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
		if err != nil {
			fmt.Printf("Failed to create request: %v\n", err)
			os.Exit(1)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Request failed: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("✅ Response: %s\n", body)
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.Flags().StringVarP(&payloadPath, "payload", "p", "", "Path to JSON payload file")
	sendCmd.Flags().StringVarP(&token, "token", "t", "", "PAT token (or set VALID_PAT env)")
	sendCmd.Flags().StringVarP(&url, "url", "u", "http://localhost:8080/cli/send-event", "Event Router URL")
}
