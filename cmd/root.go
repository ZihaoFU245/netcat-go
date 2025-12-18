/*
Copyright © 2025 zihaofu245 <zihaofu12@gmail.com>
*/
// Package cmd implements the netcat command line tool.
package cmd

import (
	"fmt"
	"nc/model"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	listen      bool
	port        int
	udp         bool
	verbose     bool
	acceptLoop  bool
	idleSeconds int
	source      string
	numeric_ip  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nc [host] [port]",
	Short: "netcat go",
	Long:  `netcat implemented in go.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// -l flag for listen mode
		if listen {
			listenPort, err := parseListenPort(args, port)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if err := model.Listen(listenPort, verbose, udp, acceptLoop, source); err != nil {
				fmt.Println(err.Error())
			}
			return
		}

		// reach out mode
		if len(args) == 2 {
			host := args[0]
			portStr := args[1]
			err := model.ConnectWithTimer(host, portStr, verbose, udp, idleSeconds, numeric_ip)
			if err != nil {
				println(err.Error())
			}
		}

		if len(args) > 2 {
			println("too many arguments, use -h flag for help")
		}
		if len(args) < 2 {
			println("missing arguments, use -h flag for help")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nc.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&listen, "listen", "l", false, "listen mode")
	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "port number")
	rootCmd.Flags().BoolVarP(&udp, "udp", "u", false, "UDP mode")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.Flags().BoolVarP(&acceptLoop, "keep-listening", "k", false, "accept loop")
	rootCmd.Flags().IntVarP(&idleSeconds, "time-outs", "w", 0, "Timeouts")
	rootCmd.Flags().StringVarP(&source, "source", "s", "", "specify source ip address")
	rootCmd.Flags().BoolVarP(&numeric_ip, "numeric-ip", "n", false, "Diable DNS lookup, only accept ip address")
}

func parseListenPort(args []string, flagPort int) (int, error) {
	if flagPort > 0 {
		return flagPort, nil
	}

	if len(args) == 0 {
		return 0, fmt.Errorf("missing port number for listen mode")
	}

	if len(args) > 1 {
		return 0, fmt.Errorf("too many positional arguments for listen mode")
	}

	portCandidate, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid port syntax")
	}

	if portCandidate <= 0 || portCandidate > 65535 {
		return 0, fmt.Errorf("invalid port range")
	}

	return portCandidate, nil
}
