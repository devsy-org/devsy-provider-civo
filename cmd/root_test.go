package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

func TestProviderLogLevels(t *testing.T) {
	for _, tc := range []struct {
		level string
		want  string
	}{
		{"", "info,warn,error"},
		{"error", "error"},
		{"warn", "warn,error"},
		{"info", "info,warn,error"},
		{"debug", "debug,info,warn,error"},
		{"trace", "trace,debug,info,warn,error"},
		{"WaRn", "warn,error"},
	} {
		t.Run("level="+tc.level, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", tc.level)
			t.Setenv("DEVSY_LOG_FORMAT", "")
			output := captureProviderLogging(t, func(_ *cobra.Command, _ []string) error {
				log.TraceContext(context.Background(), "trace message")
				log.Debug("debug message")
				log.Info("info message")
				log.Warn("warn message")
				log.Error("error message")
				return nil
			})
			var levels []string
			for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
				var record struct {
					Time    string `json:"ts"`
					Level   string `json:"level"`
					Message string `json:"msg"`
				}
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					t.Fatalf("invalid JSON log %q: %v", line, err)
				}
				if record.Time == "" || record.Message != record.Level+" message" {
					t.Fatalf("invalid log record: %+v", record)
				}
				levels = append(levels, record.Level)
			}
			want := tc.want
			if got := strings.Join(levels, ","); got != want {
				t.Fatalf("levels = %q, want %q", got, want)
			}
		})
	}
}

func TestProviderLogFormats(t *testing.T) {
	for _, format := range []string{"json", "auto", "text", "plain", "logfmt", "JsOn", "PLAIN"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", "info")
			t.Setenv("DEVSY_LOG_FORMAT", format)
			output := captureProviderLogging(t, func(_ *cobra.Command, _ []string) error {
				log.Info("format canary")
				return nil
			})
			if !strings.Contains(string(output), "format canary") {
				t.Fatal("diagnostic missing")
			}
			isJSON := json.Valid(output)
			if isJSON != (strings.EqualFold(format, "json") || strings.EqualFold(format, "auto")) {
				t.Fatalf("unexpected encoding for %s", format)
			}
		})
	}
}

func TestInvalidLoggingEnvironmentStopsCommand(t *testing.T) {
	for _, name := range []string{"DEVSY_LOG_LEVEL", "DEVSY_LOG_FORMAT"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", "")
			t.Setenv("DEVSY_LOG_FORMAT", "")
			const secret = "invalid-secret-canary"
			t.Setenv(name, secret)
			ran := false
			root := NewRootCmd()
			root.RunE = func(_ *cobra.Command, _ []string) error {
				ran = true
				return nil
			}
			root.SetArgs(nil)
			err := root.Execute()
			if err == nil || !strings.Contains(err.Error(), name) ||
				strings.Contains(err.Error(), secret) {
				t.Fatalf("unexpected configuration error: %v", err)
			}
			if ran {
				t.Fatal("command ran with invalid logging configuration")
			}
		})
	}
}

func captureProviderLogging(t *testing.T, run func(*cobra.Command, []string) error) []byte {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stderr
	os.Stderr = writer
	defer func() { os.Stderr = previous }()

	root := NewRootCmd()
	root.RunE = run
	root.SetArgs(nil)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	_ = log.Sync()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return output
}
