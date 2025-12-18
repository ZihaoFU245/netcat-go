/*
Copyright © 2025 zihaofu245 <zihaofu12@gmail.com>
*/
package model

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

// Scan performs a netcat-style "-z" port scan against the given host and port list.
// Ports are attempted with a worker pool (jobs controls concurrency, default 3); open
// ports are printed, and closed ports are only reported when verbose mode is on. If
// idleSeconds is greater than zero, the scan is bounded by that timeout. The optional
// localPort argument sets a local source port when provided (mirrors nc -p behavior).
func Scan(host string, ports []int, verbose bool, udp bool, idleSeconds int, localPort int, jobs int) error {
	if len(ports) == 0 {
		return fmt.Errorf("no ports to scan")
	}
	if jobs < 1 {
		return fmt.Errorf("jobs must be at least 1")
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if idleSeconds > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(idleSeconds)*time.Second)
		defer cancel()
	}

	dialer := net.Dialer{}
	if localPort > 0 {
		if udp {
			dialer.LocalAddr = &net.UDPAddr{Port: localPort}
		} else {
			dialer.LocalAddr = &net.TCPAddr{Port: localPort}
		}
	}

	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	errCh := make(chan error, len(ports))

	for _, port := range ports {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		sem <- struct{}{}
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := scanPort(ctx, &dialer, host, p, verbose, udp, idleSeconds); err != nil {
				if ctx.Err() != nil {
					return
				}
				errCh <- fmt.Errorf("scan error on %s:%d: %w", host, p, err)
			}
		}(port)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if verbose {
			fmt.Println(err.Error())
		}
	}

	return ctx.Err()
}

func scanPort(ctx context.Context, dialer *net.Dialer, host string, port int, verbose bool, udp bool, idleSeconds int) error {
	network := "tcp"
	if udp {
		network = "udp"
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))

	if udp {
		return scanUDP(ctx, dialer, address, host, port, verbose, idleSeconds)
	}

	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		if verbose {
			fmt.Printf("%s:%d closed (%v)\n", host, port, err)
		}
		return nil
	}
	defer conn.Close()

	fmt.Printf("%s:%d open\n", host, port)
	return nil
}

func scanUDP(ctx context.Context, dialer *net.Dialer, address, host string, port int, verbose bool, idleSeconds int) error {
	timeout := 1 * time.Second
	if idleSeconds > 0 {
		timeout = time.Duration(idleSeconds) * time.Second
	}

	conn, err := dialer.DialContext(ctx, "udp", address)
	if err != nil {
		if verbose {
			fmt.Printf("%s:%d closed (%v)\n", host, port, err)
		}
		return nil
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	if _, err := conn.Write([]byte{0}); err != nil {
		if verbose {
			fmt.Printf("%s:%d closed (%v)\n", host, port, err)
		}
		return nil
	}

	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err != nil {
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			// UDP targets often stay silent; treat as open|filtered when no ICMP response arrives.
			fmt.Printf("%s:%d open|filtered\n", host, port)
			return nil
		}
		if verbose {
			fmt.Printf("%s:%d closed (%v)\n", host, port, err)
		}
		return nil
	}

	fmt.Printf("%s:%d open\n", host, port)
	return nil
}
