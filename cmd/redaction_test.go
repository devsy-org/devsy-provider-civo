package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

func TestProviderRedactsTokenAPIKey(t *testing.T) {
	const secret = "secret\"civo-canary"
	token, err := json.Marshal(map[string]string{"apikey": secret, "region": "test-region"})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CIVO_API_KEY", "")
	t.Setenv("CIVO_TOKEN", string(token))
	t.Setenv("DEVSY_LOG_LEVEL", "info")
	t.Setenv("DEVSY_LOG_FORMAT", "auto")

	output := captureProviderLogging(t, func(_ *cobra.Command, _ []string) error {
		log.Error("useful diagnostic " + secret)
		return nil
	})
	var record struct {
		Message string `json:"msg"`
	}
	if err := json.Unmarshal(output, &record); err != nil {
		t.Fatalf("invalid JSON log: %v", err)
	}
	if strings.Contains(record.Message, secret) ||
		!strings.Contains(record.Message, "useful diagnostic") {
		t.Fatal("credential leaked or useful diagnostic was removed")
	}
}
