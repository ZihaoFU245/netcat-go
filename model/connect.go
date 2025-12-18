/*
Copyright © 2025 zihaofu245 <zihaofu12@gmail.com>
*/
// Package model implement interactions with net
package model

import (
	"context"
	"fmt"
	"io"
	"nc/util"
	"net"
	"os"
	"time"
)

// ConnectWithTimer establishes a connection with a timeout (idleSeconds).
// If idleSeconds > 0, the connection will be terminated after the specified duration.
func ConnectWithTimer(host string, portStr string, verbose bool, udp bool, idleSeconds int) error {
	var ctx context.Context
	var cancel context.CancelFunc

	// Create a context with timeout if idleSeconds is specified
	if idleSeconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(idleSeconds)*time.Second)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	return connect(ctx, host, portStr, verbose, udp)
}

// connect orchestrates the connection process: validation, establishment, and I/O handling.
func connect(ctx context.Context, host string, portStr string, verbose bool, udp bool) error {
	// Validate the port number
	port, err := util.PortCheck(portStr)
	if err != nil {
		return err
	}

	// Attempt to establish the connection (with retries)
	conn, err := establishConnection(ctx, host, port, verbose, udp)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Handle data transfer between Stdin/Stdout and the connection
	return handleIO(ctx, conn)
}

// establishConnection attempts to connect to the target host/port.
// It retries every second until successful or until the context is canceled.
func establishConnection(ctx context.Context, host, port string, verbose, udp bool) (net.Conn, error) {
	var d net.Dialer
	network := "tcp"
	if udp {
		network = "udp"
	}
	address := net.JoinHostPort(host, port)

	for {
		// Check if context is canceled before trying
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		conn, err := d.DialContext(ctx, network, address)
		if err == nil {
			if verbose {
				fmt.Println("Connected to", address)
			}
			return conn, nil
		}

		if verbose {
			fmt.Println("Connection failed, retrying...")
		}

		// Wait for 1 second or context cancellation before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
}

// handleIO manages the bidirectional data copy between the connection and Stdin/Stdout.
// It also monitors the context to close the connection on timeout.
func handleIO(ctx context.Context, conn net.Conn) error {
	// Channel to signal when the remote connection is closed
	remoteDone := make(chan struct{})

	// Copy from Connection -> Stdout
	go func() {
		_, _ = io.Copy(os.Stdout, conn)
		// When the remote side closes the connection (or read error), signal completion
		close(remoteDone)
	}()

	// Copy from Stdin -> Connection
	go func() {
		_, _ = io.Copy(conn, os.Stdin)
		// Stdin closed. We continue waiting for response from remote.
		// Note: For TCP, we could optionally CloseWrite() here.
	}()

	// Wait for either context cancellation (timeout) or remote connection close
	select {
	case <-ctx.Done():
		// Context canceled (timeout), close connection to interrupt IO
		_ = conn.Close()
		return ctx.Err()
	case <-remoteDone:
		// Remote closed connection or Read failed
		return nil
	}
}
