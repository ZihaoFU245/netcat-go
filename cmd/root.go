/*
Copyright © 2025 zihaofu245 <zihaofu12@gmail.com>
*/
// Package cmd implements the netcat command line tool.
package cmd

import (
	"errors"
	"fmt"
	"nc/model"
	"nc/util"
	"os"
	"strconv"
	"strings"

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
	ipv4Only    bool
	ipv6Only    bool
	scan        string
	jobs        int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nc [host] [port]",
	Short: "netcat go",
	Long:  `netcat implemented in go.`,

	Run: func(cmd *cobra.Command, args []string) {
		ipMode, err := model.NewIPMode(ipv4Only, ipv6Only)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Scan ports
		if scan != "" {
			host, ports, err := parseScanPort(args, scan)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if err := model.Scan(host, ports, verbose, udp, idleSeconds, port, jobs, ipMode); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			return
		}

		// -l flag for listen mode
		if listen {
			listenPort, err := parseListenPort(args, port)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if err := model.Listen(listenPort, verbose, udp, acceptLoop, source, ipMode); err != nil {
				fmt.Println(err.Error())
			}
			return
		}

		// reach out mode
		if len(args) == 2 {
			host := args[0]
			portStr := args[1]
			err := model.ConnectWithTimer(host, portStr, verbose, udp, idleSeconds, numeric_ip, ipMode)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			return
		}

		if len(args) > 2 {
			println("too many arguments, use -h flag for help")
			return
		}
		if len(args) < 2 {
			println("missing arguments, use -h flag for help")
			return
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
	rootCmd.Flags().BoolVarP(&numeric_ip, "numeric-ip", "n", false, "Disable DNS lookup, only accept ip address")
	rootCmd.Flags().BoolVarP(&ipv4Only, "ipv4", "4", false, "IPv4 only")
	rootCmd.Flags().BoolVarP(&ipv6Only, "ipv6", "6", false, "IPv6 only")
	rootCmd.Flags().StringVarP(&scan, "scan", "z", "", "Scan a range of ports, [start]:[end], or 80 443 22 ...")
	rootCmd.Flags().IntVarP(&jobs, "jobs", "j", 3, "Number of concurrency for -z port scan")
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

func parseScanPort(args []string, flagRange string) (string, []int, error) {
	var portRange []int
	seen := make(map[int]struct{})

	if len(args) == 0 {
		return "", nil, errors.New("-z missing host, use -h for help")
	}

	host := strings.TrimSpace(args[0])
	if host == "" {
		return "", nil, errors.New("-z missing host, use -h for help")
	}

	addPort := func(portVal int) {
		if _, exists := seen[portVal]; exists {
			return
		}
		seen[portVal] = struct{}{}
		portRange = append(portRange, portVal)
	}

	addPortStr := func(portStr string) error {
		validated, err := util.PortCheck(strings.TrimSpace(portStr))
		if err != nil {
			return errors.New("port parsing failed, use -h for help")
		}
		portVal, _ := strconv.Atoi(validated)
		addPort(portVal)
		return nil
	}

	// -z called with range or single value
	if flagRange != "" {
		parts := strings.Split(strings.TrimSpace(flagRange), ":")
		switch len(parts) {
		case 1:
			if err := addPortStr(parts[0]); err != nil {
				return "", nil, err
			}
		case 2:
			startStr, err1 := util.PortCheck(strings.TrimSpace(parts[0]))
			endStr, err2 := util.PortCheck(strings.TrimSpace(parts[1]))
			if err1 != nil || err2 != nil {
				return "", nil, errors.New("port parsing failed, use -h for help")
			}
			start, _ := strconv.Atoi(startStr)
			end, _ := strconv.Atoi(endStr)
			if start > end {
				return "", nil, errors.New("port parsing failed, use -h for help")
			}
			for p := start; p <= end; p++ {
				addPort(p)
			}
		default:
			return "", nil, errors.New("port parsing failed, use -h for help")
		}
	}

	if len(args) == 1 && len(portRange) == 0 {
		return "", nil, errors.New("-z missing ports, use -h for help")
	}

	for _, strPort := range args[1:] {
		if err := addPortStr(strPort); err != nil {
			return "", nil, err
		}
	}

	return host, portRange, nil
}
