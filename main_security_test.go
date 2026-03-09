package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/logdyhq/logdy-core/http"
	"github.com/spf13/cobra"
)

func testCmdForConfig() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("port", "8080", "")
	cmd.Flags().String("ui-ip", "127.0.0.1", "")
	cmd.Flags().String("ui-pass", "", "")
	cmd.Flags().String("config", "", "")
	cmd.Flags().String("append-to-file", "", "")
	cmd.Flags().String("rotate-file-size", "", "")
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().Int64("bulk-window", 100, "")
	cmd.Flags().Int64("max-message-count", 100_000, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("disable-ansi-code-stripping", false, "")
	cmd.Flags().Bool("append-to-file-raw", false, "")
	cmd.Flags().Bool("analytics", false, "")
	cmd.Flags().Bool("no-analytics", false, "")
	cmd.Flags().Bool("allow-insecure-bind", false, "")
	cmd.Flags().Bool("fallthrough", false, "")
	return cmd
}

func TestAnalyticsDefaultIsDisabled(t *testing.T) {
	cmd := testCmdForConfig()
	parseConfig(cmd)
	if config == nil {
		t.Fatal("config should be initialized")
	}
	if !config.AnalyticsDisabled {
		t.Fatal("analytics must be disabled by default")
	}
}

func TestAnalyticsCanBeEnabledExplicitly(t *testing.T) {
	cmd := testCmdForConfig()
	if err := cmd.Flags().Set("analytics", "true"); err != nil {
		t.Fatalf("set analytics flag: %v", err)
	}
	parseConfig(cmd)
	if config == nil {
		t.Fatal("config should be initialized")
	}
	if config.AnalyticsDisabled {
		t.Fatal("analytics should be enabled when --analytics=true")
	}
}

func TestNoAnalyticsOverridesAnalyticsFlag(t *testing.T) {
	cmd := testCmdForConfig()
	if err := cmd.Flags().Set("analytics", "true"); err != nil {
		t.Fatalf("set analytics flag: %v", err)
	}
	if err := cmd.Flags().Set("no-analytics", "true"); err != nil {
		t.Fatalf("set no-analytics flag: %v", err)
	}
	parseConfig(cmd)
	if config == nil {
		t.Fatal("config should be initialized")
	}
	if !config.AnalyticsDisabled {
		t.Fatal("analytics should be disabled when --no-analytics=true")
	}
}

func TestIsPublicBind(t *testing.T) {
	cases := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", false},
		{"localhost", false},
		{"localhost.internal", false},
		{"myhost.local", true},
		{"::1", false},
		{"0.0.0.0", true},
		{"192.168.1.10", true},
		{"", true},
	}
	for _, tc := range cases {
		got := isPublicBind(tc.ip)
		if got != tc.want {
			t.Fatalf("isPublicBind(%q)=%v want %v", tc.ip, got, tc.want)
		}
	}
}

func TestValidateServerExposure(t *testing.T) {
	if err := validateServerExposure(&http.Config{
		ServerIp: "127.0.0.1",
	}); err != nil {
		t.Fatalf("loopback should not require auth: %v", err)
	}

	if err := validateServerExposure(&http.Config{
		ServerIp: "0.0.0.0",
		UiPass:   "secret",
		ApiKey:   "token",
	}); err != nil {
		t.Fatalf("public bind with auth should pass: %v", err)
	}

	if err := validateServerExposure(&http.Config{
		ServerIp: "0.0.0.0",
	}); err == nil {
		t.Fatal("public bind without auth must fail")
	}

	if err := validateServerExposure(&http.Config{
		ServerIp:          "0.0.0.0",
		AllowInsecureBind: true,
	}); err != nil {
		t.Fatalf("allow-insecure-bind should bypass validation: %v", err)
	}
}

func TestStartWebServerRejectsInsecureBindAtStartup(t *testing.T) {
	if os.Getenv("LOGDY_TEST_HELPER_START_WEBSERVER") == "1" {
		config = &http.Config{
			ServerIp:          "0.0.0.0",
			AnalyticsDisabled: true,
			AllowInsecureBind: false,
			ServerPort:        "0",
		}
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("no-updates", true, "")
		_ = cmd.Flags().Set("no-updates", "true")
		startWebServer(cmd)
		return
	}

	c := exec.Command(os.Args[0], "-test.run=TestStartWebServerRejectsInsecureBindAtStartup")
	c.Env = append(os.Environ(), "LOGDY_TEST_HELPER_START_WEBSERVER=1")
	_, err := c.CombinedOutput()
	if err == nil {
		t.Fatal("expected startup to fail for insecure non-local bind without auth")
	}
}
