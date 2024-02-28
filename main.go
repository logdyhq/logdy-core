package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"logdy/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"logdy/models"
	"logdy/modes"
)

var ch chan models.Message

var Version = "0.6.0"

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
	PersistentPostRun: func(cmd *cobra.Command, args []string) {

		noupdates, _ := cmd.Flags().GetBool("no-updates")
		if !noupdates {
			go utils.CheckUpdatesAndPrintInfo(Version)
		}

		if len(args) == 0 {
			utils.Logger.Info("Listen to stdin (from pipe)")
			go modes.ConsumeStdin(ch)
		}

		httpPort, _ := cmd.Flags().GetString("port")
		uiPass, _ := cmd.Flags().GetString("ui-pass")
		configFile, _ := cmd.Flags().GetString("config")
		noanalytics, _ := cmd.Flags().GetBool("no-analytics")
		modes.FallthroughGlobal, _ = cmd.Flags().GetBool("fallthrough")
		verbose, _ := cmd.Flags().GetBool("verbose")
		bulkWindow, _ := cmd.Flags().GetInt64("bulk-window")
		maxMessageCount, _ := cmd.Flags().GetInt64("max-message-count")

		if !noanalytics {
			utils.Logger.Warn("No opt-out from analytics, we'll be receiving anonymous usage data, which will be used to improve the product. To opt-out use the flag --no-analytics.")
		}

		if verbose {
			utils.Logger.SetLevel(logrus.TraceLevel)
		} else {
			utils.Logger.SetLevel(logrus.InfoLevel)
		}

		handleHttp(ch, httpPort, !noanalytics, uiPass, configFile, bulkWindow, maxMessageCount)
	},
}

var listenStdCmd = &cobra.Command{
	Use:   "stdin [command]",
	Short: "Listens to STDOUT/STDERR of a provided command. Example `logdy stdin \"npm run dev\"`",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			utils.Logger.Info("Listen to stdin (from pipe)")
			go modes.ConsumeStdin(ch)
			return
		}

		utils.Logger.WithFields(logrus.Fields{
			"cmd": args[0],
		}).Info("Listen to command stdout")
		arg := strings.Split(args[0], " ")
		modes.StartCmd(ch, arg[0], arg[1:])
	},
}

var followCmd = &cobra.Command{
	Use:   "follow <file1> [<file2> ... <fileN>]",
	Short: "Follows lines added to files. Example `logdy follow foo.log /var/log/bar.log`",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modes.FollowFiles(ch, args)
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

var listenSocketCmd = &cobra.Command{
	Use:   "socket <port1> [<port2> ... <portN>]",
	Short: "Sets up a port to listen on for incoming log messages. Example `logdy socket 8233`. You can setup multiple ports `logdy socket 8123 8124 8125`",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")
		go modes.StartSocketServers(ch, ip, args)
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

		go modes.GenerateRandomData(produceJson, num, ch, context.Background())
	},
}

func init() {
	ch = make(chan models.Message, 1000)
	rootCmd.PersistentFlags().StringP("port", "p", "8080", "Port on which the Web UI will be served")
	rootCmd.PersistentFlags().StringP("ui-pass", "", "", "Password that will be used to authenticate in the UI")
	rootCmd.PersistentFlags().StringP("config", "", "", "Path to a file where a config (json) for the UI is located")
	rootCmd.PersistentFlags().Int64P("bulk-window", "", 100, "A time window during which log messages are gathered and send in a bulk to a client. Decreasing this window will improve the 'real-time' feeling of messages presented on the screen but could decrease UI performance")
	rootCmd.PersistentFlags().Int64P("max-message-count", "", 100_000, "Max number of messages that will be stored in a buffer for further retrieval. On buffer overflow, oldest messages will be removed.")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose logs")
	rootCmd.PersistentFlags().BoolP("no-analytics", "n", false, "Opt-out from sending anonymous analytical data that helps improve Logdy")
	rootCmd.PersistentFlags().BoolP("no-updates", "u", false, "Opt-out from checking updates on program startup")
	rootCmd.PersistentFlags().BoolP("fallthrough", "t", false, "When used will fallthrough all of the stdin received to the terminal as is")
	demoSocketCmd.PersistentFlags().BoolP("sample-text", "", true, "By default demo data will produce JSON, use this flag to produce raw text")
	listenSocketCmd.PersistentFlags().StringP("ip", "", "", "IP address to listen to, leave empty to listen on all IP addresses")

	utils.InitLogger()

	rootCmd.AddCommand(listenStdCmd)
	rootCmd.AddCommand(listenSocketCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(demoSocketCmd)
	rootCmd.AddCommand(followCmd)
}

func main() {
	utils.Logger.Out = os.Stdout
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	utils.Logger.Debug("Exiting")
}
