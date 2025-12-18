/*
Copyright © 2025 zihaofu245 <zihaofu12@gmail.com>
*/
package model

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
)

// Listen starts a TCP/UDP listener with optional source filtering and keep-alive behavior.
func Listen(port int, verbose bool, udp bool, keepOpen bool, source string) error {
	if err := validatePort(port); err != nil {
		return err
	}

	allowedIP, err := resolveSource(source)
	if err != nil {
		return err
	}

	cfg := listenConfig{
		port:      port,
		verbose:   verbose,
		keepOpen:  keepOpen,
		allowedIP: allowedIP,
	}

	announceMode(cfg.port, udp)

	if udp {
		return listenUDP(cfg)
	}

	return listenTCP(cfg)
}

type listenConfig struct {
	port      int
	verbose   bool
	keepOpen  bool
	allowedIP net.IP
}

func validatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("missing valid port number")
	}

	return nil
}

func resolveSource(source string) (net.IP, error) {
	if source == "" {
		return nil, nil
	}

	addr, err := net.ResolveIPAddr("ip", source)
	if err != nil {
		return nil, fmt.Errorf("invalid source address: %w", err)
	}

	return addr.IP, nil
}

func announceMode(port int, udp bool) {
	protocol := "TCP"
	if udp {
		protocol = "UDP"
	}

	fmt.Printf("Listening on port %d (%s)\n", port, protocol)
}

func (cfg listenConfig) allowed(addr net.IP) bool {
	if cfg.allowedIP == nil {
		return true
	}

	return cfg.allowedIP.Equal(addr)
}

func listenUDP(cfg listenConfig) error {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(cfg.port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, 4096)

	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			if cfg.verbose {
				fmt.Fprintf(os.Stderr, "udp read error: %v\n", err)
			}
			if !cfg.keepOpen {
				return err
			}
			continue
		}

		if !cfg.allowed(remote.IP) {
			if cfg.verbose {
				fmt.Fprintf(os.Stderr, "ignored packet from %s\n", remote.IP.String())
			}
			if !cfg.keepOpen {
				return nil
			}
			continue
		}

		if _, err := os.Stdout.Write(buf[:n]); err != nil {
			return err
		}

		if !cfg.keepOpen {
			return nil
		}
	}
}

func listenTCP(cfg listenConfig) error {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.port))
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			if cfg.verbose {
				fmt.Fprintf(os.Stderr, "accept error: %v\n", err)
			}
			if !cfg.keepOpen {
				return err
			}
			continue
		}

		remoteIP := extractIP(conn.RemoteAddr())
		if !cfg.allowed(remoteIP) {
			if cfg.verbose {
				fmt.Fprintf(os.Stderr, "rejected connection from %s\n", remoteIP.String())
			}
			_ = conn.Close()
			if !cfg.keepOpen {
				return nil
			}
			continue
		}

		printConnectionInfo(conn)

		if cfg.keepOpen {
			go handleTCPConnection(conn, cfg.verbose)
			continue
		}

		return handleTCPConnection(conn, cfg.verbose)
	}
}

func handleTCPConnection(conn net.Conn, verbose bool) error {
	var once sync.Once
	closeConn := func() { _ = conn.Close() }

	done := make(chan struct{})

	go func() {
		if _, err := io.Copy(conn, os.Stdin); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "error sending data: %v\n", err)
		}
		once.Do(closeConn)
	}()

	go func() {
		if _, err := io.Copy(os.Stdout, conn); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "error receiving data: %v\n", err)
		}
		once.Do(closeConn)
		close(done)
	}()

	<-done
	return nil
}

func extractIP(addr net.Addr) net.IP {
	switch a := addr.(type) {
	case *net.TCPAddr:
		return a.IP
	case *net.UDPAddr:
		return a.IP
	default:
		return nil
	}
}

func printConnectionInfo(conn net.Conn) {
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		fmt.Printf("Connection from %s %d\n", addr.IP.String(), addr.Port)
		return
	}
	fmt.Printf("Connection from %s\n", conn.RemoteAddr().String())
}
