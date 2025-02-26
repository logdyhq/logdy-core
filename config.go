package main

import (
	"os"

	"github.com/logdyhq/logdy-core/http"
	"github.com/logdyhq/logdy-core/modes"
	"github.com/spf13/cobra"
)

type ConfVal interface {
	string | int64 | bool
}

func getConfigValue[T ConfVal](arg ...T) T {
	var zero T

	for _, v := range arg {
		// Check if v is not the zero value of its type
		switch any(v).(type) {
		case string:
			if any(v).(string) != "" {
				return v
			}
		case int64:
			if any(v).(int64) != 0 {
				return v
			}
		case bool:
			// For bool, we consider true as non-zero
			// If you want to treat false as valid too, adjust this condition
			if any(v).(bool) {
				return v
			}
		}
	}

	return zero
}

func getFlagString(name string, cmd *cobra.Command, def bool) string {
	v, _ := cmd.Flags().GetString(name)
	if def {
		return v
	}
	if !def && cmd.Flags().Changed(name) {
		return v
	}
	return ""
}

func getFlagBool(name string, cmd *cobra.Command, def bool) bool {
	v, _ := cmd.Flags().GetBool(name)
	if def {
		return v
	}
	if !def && cmd.Flags().Changed(name) {
		return v
	}
	return false
}

func getFlagInt(name string, cmd *cobra.Command, def bool) int64 {
	v, _ := cmd.Flags().GetInt64(name)
	if def {
		return v
	}
	if !def && cmd.Flags().Changed(name) {
		return v
	}
	return 0
}

// this function controls the precedence of arguments for string values passed:
// 1. cli
// 2. env
// 3. default
func getStringCfgVal(cli, env string, cmd *cobra.Command) string {
	return getConfigValue(getFlagString(cli, cmd, false), os.Getenv(env), getFlagString(cli, cmd, true))
}

// this function controls the precedence of arguments for bool values passed:
// 1. cli
// 2. default
func getBoolCfgVal(cli string, cmd *cobra.Command) bool {
	return getConfigValue(getFlagBool(cli, cmd, false), getFlagBool(cli, cmd, true))
}

// this function controls the precedence of arguments for int values passed:
// 1. cli
// 2. default
func getIntCfgVal(cli string, cmd *cobra.Command) int64 {
	return getConfigValue(getFlagInt(cli, cmd, false), getFlagInt(cli, cmd, true))
}

func parseConfig(cmd *cobra.Command) {
	config = &http.Config{
		HttpPathPrefix: "",
	}

	const prefix = "LOGDY_"

	config.ServerPort = getStringCfgVal("port", prefix+"PORT", cmd)
	config.ServerIp = getStringCfgVal("ui-ip", prefix+"UI_IP", cmd)
	config.UiPass = getStringCfgVal("ui-pass", prefix+"UI_PASS", cmd)
	config.ConfigFilePath = getStringCfgVal("config", prefix+"CONFIG", cmd)
	config.AppendToFile = getStringCfgVal("append-to-file", prefix+"APPEND_TO_FILE", cmd)
	config.ApiKey = getStringCfgVal("api-key", prefix+"API_KEY", cmd)

	config.BulkWindowMs = getIntCfgVal("bulk-window", cmd)
	config.MaxMessageCount = getIntCfgVal("max-message-count", cmd)

	config.AppendToFileRaw = getBoolCfgVal("append-to-file-raw", cmd)
	config.AnalyticsDisabled = getBoolCfgVal("no-analytics", cmd)
	modes.FallthroughGlobal = getBoolCfgVal("fallthrough", cmd)
	modes.DisableANSICodeStripping = getBoolCfgVal("disable-ansi-code-stripping", cmd)
}
