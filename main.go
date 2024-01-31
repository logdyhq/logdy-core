package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ch chan Message

var rootCmd = &cobra.Command{
	Use:   "logdy [command]",
	Short: "Logdy",
	Args:  cobra.MinimumNArgs(1),
	//https://patorjk.com/software/taag/#p=display&f=Colossal&t=Logdy%20v0.1
	Long: `	


888                            888                         .d8888b.       d888   
888                            888                        d88P  Y88b     d8888   
888                            888                        888    888       888   
888      .d88b.   .d88b.   .d88888 888  888      888  888 888    888       888   
888     d88""88b d88P"88b d88" 888 888  888      888  888 888    888       888   
888     888  888 888  888 888  888 888  888      Y88  88P 888    888       888   
888     Y88..88P Y88b 888 Y88b 888 Y88b 888       Y8bd8P  Y88b  d88P d8b   888   
88888888 "Y88P"   "Y88888  "Y88888  "Y88888        Y88P    "Y8888P"  Y8P 8888888 
                      888               888                                      
                 Y8b d88P          Y8b d88P                                      
                  "Y88P"            "Y88P"                                       

	
Visit https://logdy.dev for more info!
Logdy is a hackable web UI for all kinds of logs produced locally. 
Break free from the terminal and stream your logs in any format to a web UI 
where you can filter and browse well formatted application output.
	`,
	Run: func(cmd *cobra.Command, args []string) {
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		httpPort, _ := cmd.Flags().GetString("port")
		noanalytics, _ := cmd.Flags().GetBool("no-analytics")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if !noanalytics {
			logger.Warn("No opt-out from analytics, we'll be receiving anonymous usage data, which will be used to improve the product. To opt-out use the flag --no-analytics.")
		}

		if verbose {
			logger.SetLevel(logrus.TraceLevel)
		} else {
			logger.SetLevel(logrus.InfoLevel)
		}

		handleHttp(ch, httpPort, !noanalytics)
	},
}

var listenStdCmd = &cobra.Command{
	Use:   "stdin [command]",
	Short: "Listens to STDOUT/STDERR of a provided command. Example ./logdy stdin \"npm run dev\"",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger.WithFields(logrus.Fields{
			"cmd": args[0],
		}).Info("Command")

		arg := strings.Split(args[0], " ")
		startCmd(ch, arg[0], arg[1:])
	},
}

var listenSocketCmd = &cobra.Command{
	Use:   "socket [port]",
	Short: "Sets up a port to listen on for incoming log messages. Example ./logdy socket 8233",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip, _ := cmd.Flags().GetString("ip")

		go startSocketServer(ch, ip, args[0])
	},
}

var demoSocketCmd = &cobra.Command{
	Use:   "demo [number]",
	Short: "Starts a demo mode, random logs will be produced, the [number] defines a number of messages produced per second",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		produceJson := !cmd.Flag("sample-text").Changed
		num, err := strconv.Atoi(args[0])

		if err != nil {
			panic(err)
		}

		go generateRandomData(produceJson, num, ch)
	},
}

func init() {
	ch = make(chan Message, 1000)
	rootCmd.PersistentFlags().StringP("port", "p", "8080", "Port on which the Web UI will be served")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose logs")
	rootCmd.PersistentFlags().BoolP("no-analytics", "n", false, "Opt-out from sending anonymous analytical data that help improve this product")
	demoSocketCmd.PersistentFlags().BoolP("sample-text", "", true, "By default demo data will produce JSON, use this flag to produce raw text")
	listenSocketCmd.PersistentFlags().StringP("ip", "", "", "IP address to listen to, leave empty to listen on all IP addresses")

	initLogger()

	rootCmd.AddCommand(listenStdCmd)
	rootCmd.AddCommand(listenSocketCmd)
	rootCmd.AddCommand(demoSocketCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
