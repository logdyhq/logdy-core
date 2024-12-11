package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/logdyhq/logdy-core/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/logdyhq/logdy-core/http"
	"github.com/logdyhq/logdy-core/modes"
)

var Version = "0.0.0"

var config *http.Config

var rootCmd = &cobra.Command{
	Use:     "logdy [command]",
	Short:   "Logdy",
	Version: Version,
	Long: `Visit https://logdy.dev for more info!
Logdy is a hackable web UI for all kinds of logs produced locally. 
Break free from the terminal and stream your logs in any format to a web UI 
where you can filter and browse well formatted application output.
	`,
	Run: func(cmd *cobra.Command, args []string) {
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		parseConfig(cmd)

		verbose, _ := cmd.Flags().GetBool("verbose")
		utils.SetLoggerLevel(verbose)
	},
}

var listenStdCmd = &cobra.Command{
	Use:   "stdin [command]",
	Short: "Listens to STDOUT/STDERR of a provided command. Example `logdy stdin \"npm run dev\"`",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			utils.Logger.Info("Listen to stdin (from pipe)")
			go modes.ConsumeStdin(http.Ch)
			return
		}

		utils.Logger.WithFields(logrus.Fields{
			"cmd": args[0],
		}).Info("Listen to command stdout")
		arg := strings.Split(args[0], " ")
		modes.StartCmd(http.Ch, arg[0], arg[1:])
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		startWebServer(cmd, args)
	},
}

var followCmd = &cobra.Command{
	Use:   "follow <file1> [<file2> ... <fileN>]",
	Short: "Follows lines added to files. Example `logdy follow foo.log /var/log/bar.log`",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		http.InitializeClients(*config)
		fullRead, _ := cmd.Flags().GetBool("full-read")

		if fullRead {
			modes.ReadFiles(http.Ch, args)
		}

		modes.FollowFiles(http.Ch, args)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		startWebServer(cmd, args)
	},
}

var forwardCmd = &cobra.Command{
	Use:   "forward <port>",
	Short: "Forwards the STDIN to a specified port, example `tail -f file.log | logdy forward 8123`",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")

		modes.ConsumeStdinAndForwardToPort(ip, args[0])
	},
}

var UtilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "A set of utility commands that help working with large files",
}

var utilsCutByStringCmd = &cobra.Command{
	Use:   "cut-by-string <file> <start> <end> {case-insensitive = true} {out-file = ''}",
	Short: "A utility that cuts a file by a start and end string into a new file or standard output.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetLoggerDiscard(true)
		modes.UtilsCutByString(utils.AString(args, 0, ""), utils.AString(args, 1, ""), utils.AString(args, 2, ""),
			utils.ABool(args, 3, true), utils.AString(args, 4, ""), "", 0)
	},
}
var utilsCutByLineNumberCmd = &cobra.Command{
	Use:   "cut-by-line-number <file> <count> <offset> {out-file = ''}",
	Short: "A utility that cuts a file by a line number count and offset into a new file or standard output.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetLoggerDiscard(true)

		modes.UtilsCutByLineNumber(utils.AString(args, 0, ""), utils.AInt(args, 1, 0), utils.AInt(args, 2, 0), utils.AString(args, 3, ""))
	},
}

var utilsCutByDateCmd = &cobra.Command{
	Use:   "cut-by-date <file> <start> <end> <date-format> <search-offset> {out-file = ''}",
	Short: "A utility that cuts a file by a start and end date into a new file or standard output.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetLoggerDiscard(true)

		modes.UtilsCutByString(utils.AString(args, 0, ""), utils.AString(args, 1, ""), utils.AString(args, 2, ""), false,
			utils.AString(args, 5, ""), utils.AString(args, 3, ""), utils.AInt(args, 4, 0))
	},
}

var listenSocketCmd = &cobra.Command{
	Use:   "socket <port1> [<port2> ... <portN>]",
	Short: "Sets up a port to listen on for incoming log messages. Example `logdy socket 8233`. You can setup multiple ports `logdy socket 8123 8124 8125`",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")
		go modes.StartSocketServers(http.Ch, ip, args)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		startWebServer(cmd, args)
	},
}

var demoSocketCmd = &cobra.Command{
	Use:   "demo [number]",
	Short: "Starts a demo mode, random logs will be produced, the [number] defines a number of messages produced per second",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		produceJson := !cmd.Flag("sample-text").Changed
		num := 1
		if len(args) == 1 {
			var err error
			num, err = strconv.Atoi(args[0])
			if err != nil {
				panic(err)
			}
		}

		go modes.GenerateRandomData(produceJson, num, http.Ch, context.Background())
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		startWebServer(cmd, args)
	},
}

func startWebServer(cmd *cobra.Command, args []string) {
	noupdates, _ := cmd.Flags().GetBool("no-updates")
	if !noupdates && Version != "0.0.0" {
		go utils.CheckUpdatesAndPrintInfo(Version)
	}

	if len(args) == 0 {
		utils.Logger.Info("Listen to stdin (from pipe)")
		go modes.ConsumeStdin(http.Ch)
	}

	if !config.AnalyticsEnabled {
		utils.Logger.Warn("No opt-out from analytics, we'll be receiving anonymous usage data, which will be used to improve the product. To opt-out use the flag --no-analytics.")
	}

	http.HandleHttp(config, http.InitializeClients(*config), nil)
	http.StartWebserver(config)
}

func parseConfig(cmd *cobra.Command) {
	config = &http.Config{
		HttpPathPrefix: "",
	}

	config.ServerPort, _ = cmd.Flags().GetString("port")
	config.ServerIp, _ = cmd.Flags().GetString("ui-ip")
	config.UiPass, _ = cmd.Flags().GetString("ui-pass")
	config.ConfigFilePath, _ = cmd.Flags().GetString("config")
	config.BulkWindowMs, _ = cmd.Flags().GetInt64("bulk-window")
	config.AppendToFile, _ = cmd.Flags().GetString("append-to-file")
	config.ApiKey, _ = cmd.Flags().GetString("api-key")
	config.AppendToFileRaw, _ = cmd.Flags().GetBool("append-to-file-raw")
	config.MaxMessageCount, _ = cmd.Flags().GetInt64("max-message-count")
	config.AnalyticsEnabled, _ = cmd.Flags().GetBool("no-analytics")

	modes.FallthroughGlobal, _ = cmd.Flags().GetBool("fallthrough")
	modes.DisableANSICodeStripping, _ = cmd.Flags().GetBool("disable-ansi-code-stripping")
}

func init() {
	utils.InitLogger()
	http.InitChannel()

	rootCmd.AddCommand(UtilsCmd)
	UtilsCmd.AddCommand(utilsCutByStringCmd)
	UtilsCmd.AddCommand(utilsCutByDateCmd)
	UtilsCmd.AddCommand(utilsCutByLineNumberCmd)

	rootCmd.PersistentFlags().StringP("port", "p", "8080", "Port on which the Web UI will be served")
	rootCmd.PersistentFlags().StringP("ui-ip", "", "127.0.0.1", "Bind Web UI server to a specific IP address")
	rootCmd.PersistentFlags().StringP("ui-pass", "", "", "Password that will be used to authenticate in the UI")
	rootCmd.PersistentFlags().StringP("config", "", "", "Path to a file where a config (json) for the UI is located")
	rootCmd.PersistentFlags().StringP("append-to-file", "", "", "Path to a file where message logs will be appended, the file will be created if it doesn't exist")
	rootCmd.PersistentFlags().StringP("api-key", "", "", "API key (send as a header "+http.API_KEY_HEADER_NAME+")")
	rootCmd.PersistentFlags().Int64P("bulk-window", "", 100, "A time window during which log messages are gathered and send in a bulk to a client. Decreasing this window will improve the 'real-time' feeling of messages presented on the screen but could decrease UI performance")
	rootCmd.PersistentFlags().Int64P("max-message-count", "", 100_000, "Max number of messages that will be stored in a buffer for further retrieval. On buffer overflow, oldest messages will be removed.")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose logs")
	rootCmd.PersistentFlags().BoolP("disable-ansi-code-stripping", "", false, "Use this flag to disable Logdy from stripping ANSI sequence codes")
	rootCmd.PersistentFlags().BoolP("append-to-file-raw", "", false, "When 'append-to-file' is set, raw lines without metadata will be saved to a file")
	rootCmd.PersistentFlags().BoolP("no-analytics", "n", false, "Opt-out from sending anonymous analytical data that helps improve Logdy")
	rootCmd.PersistentFlags().BoolP("no-updates", "u", false, "Opt-out from checking updates on program startup")
	rootCmd.PersistentFlags().BoolP("fallthrough", "t", false, "Will fallthrough all of the stdin received to the terminal as is (will display incoming messages)")

	rootCmd.AddCommand(listenStdCmd)

	listenSocketCmd.PersistentFlags().StringP("ip", "", "", "IP address to listen to, leave empty to listen on all IP addresses")
	rootCmd.AddCommand(listenSocketCmd)

	rootCmd.AddCommand(forwardCmd)

	demoSocketCmd.PersistentFlags().BoolP("sample-text", "", true, "By default demo data will produce JSON, use this flag to produce raw text")
	rootCmd.AddCommand(demoSocketCmd)

	followCmd.Flags().BoolP("full-read", "", false, "Whether the the file(s) should be read entirely")
	rootCmd.AddCommand(followCmd)

}

func main() {
	utils.SetLoggerDiscard(false)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	utils.Logger.Debug("Exiting")
}
