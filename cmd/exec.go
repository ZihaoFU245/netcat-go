/*
Copyright © 2025 zihaofu245 <zihaofu12@gmail.com>
*/
package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"nc/model"
	"nc/util"
)

var (
	unsafeMode bool
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a program with stdio",
	Long: `Execute a program using stdio. For example:
	
	nc exec [binary path] -p [port]`,
	Run: func(cmd *cobra.Command, args []string) {
		if !unsafeMode {
			fmt.Fprintln(os.Stderr, "exec is disabled; re-run with --unsafe to confirm you understand the risk")
			os.Exit(1)
		}

		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "missing program to execute")
			os.Exit(1)
		}

		if port <= 0 {
			fmt.Fprintln(os.Stderr, "missing port, supply it with -p")
			os.Exit(1)
		}

		portStr := strconv.Itoa(port)
		if _, err := util.PortCheck(portStr); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		ipMode, err := model.NewIPMode(ipv4Only, ipv6Only)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		listener, err := net.Listen(ipMode.Network(false), ":"+portStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer listener.Close()

		if verbose {
			fmt.Fprintf(os.Stderr, "exec listening on port %d (%s)\n", port, ipMode.Network(false))
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "accept error: %v\n", err)
				}
				if acceptLoop {
					continue
				}
				os.Exit(1)
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "connection from %s\n", conn.RemoteAddr().String())
			}

			if err := execWithConn(conn, args[0], args[1:], verbose); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "exec error: %v\n", err)
			}

			if !acceptLoop {
				break
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	execCmd.Flags().BoolVar(&unsafeMode, "unsafe", false, "ALLOW executing a program for each incoming connection (dangerous)")
	execCmd.Flags().IntVarP(&port, "port", "p", 0, "port to listen on for exec")
	execCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	execCmd.Flags().BoolVarP(&acceptLoop, "keep-alive", "k", false, "keep accepting and spawning programs")
	execCmd.Flags().BoolVarP(&ipv4Only, "ipv4", "4", false, "IPv4 only")
	execCmd.Flags().BoolVarP(&ipv6Only, "ipv6", "6", false, "IPv6 only")
}

// execWithConn wires a single network connection to a child process using stdio.
func execWithConn(conn net.Conn, program string, programArgs []string, verbose bool) error {
	defer conn.Close()

	cmd := exec.Command(program, programArgs...)
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn

	if verbose {
		fmt.Fprintf(os.Stderr, "running: %s\n", strings.Join(cmd.Args, " "))
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}
