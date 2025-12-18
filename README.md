# Netcat (Go Implementation)

A robust implementation of the `netcat` utility written in Go. This tool provides network debugging and investigation capabilities, supporting TCP/UDP connections, listening, and port scanning with IPv4/IPv6 support.

## Features

- **Client Mode**: Connect to arbitrary TCP/UDP ports.
- **Server Mode**: Listen on arbitrary TCP/UDP ports.
- **Port Scanning**: Fast, concurrent port scanning with customizable workers.
- **IPv4/IPv6**: Full support for forcing IPv4 (`-4`) or IPv6 (`-6`).
- **Concurrency**: Multi-threaded port scanning (`-j`).
- **Access Control**: Source IP filtering in listen mode (`-s`).
- **Persistence**: Keep-alive listener mode (`-k`).
- **Timeouts**: Connection and idle timeouts (`-w`).

## Usage

```bash
nc [host] [port] [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help message |
| `--ipv4` | `-4` | Force IPv4 only |
| `--ipv6` | `-6` | Force IPv6 only |
| `--jobs` | `-j` | Number of concurrent workers for scanning (default 3) |
| `--keep-alive` | `-k` | Keep server open after client disconnects |
| `--listen` | `-l` | Listen mode (server) |
| `--numeric-ip` | `-n` | Disable DNS lookup (numeric IP only) |
| `--port` | `-p` | Source port (client/scan) or Listen port (server) |
| `--scan` | `-z` | Scan mode (e.g., `20:80` or `80 443 22`) |
| `--source` | `-s` | Specify source IP address (for filtering or binding) |
| `--time-outs` | `-w` | Connection/Idle timeout in seconds |
| `--udp` | `-u` | UDP mode |
| `--verbose` | `-v` | Verbose output |

## Examples

### 1. Simple Chat (Client-Server)
**Server (Listen on port 8080):**
```bash
./nc -l -p 8080 -v
```
**Client (Connect to server):**
```bash
./nc localhost 8080 -v
```

### 2. Port Scanning
**Scan ports 60 through 80 on example.com:**
```bash
./nc -v -z 60:80 example.com -j 20 -w 10
```
**Scan specific ports with high concurrency:**
```bash
./nc  example.com -v -j 10 -z 80 443 8080
```

> [!WARNING]  
> File transfer is not working under this version and will be fixed

### 3. File Transfer
**Receiver (Listen and write to file):**
```bash
./nc -l -p 9000 > send.txt
```
**Sender (Connect and send file):**
```bash
./nc localhost 9000 < receive.txt
```

### 4. UDP Connection
**Server:**
```bash
./nc -u -l -p 5000 -v -k
```
**-k for keep listening or program will exit after received first UDP packet**
**Client:**
```bash
./nc -u localhost 5000 -v
```

## Implementation Progress

### Implemented ✅
- [x] **TCP Client/Server**: Basic connection and listening.
- [x] **UDP Client/Server**: UDP packet sending and receiving.
- [x] **Port Scanning**: Range and list scanning with concurrency control.
- [x] **IP Version Control**: Force IPv4 or IPv6.
- [x] **Source Filtering**: Restrict connections to a specific source IP.
- [x] **Persistence**: `-k` flag to keep listener alive.
- [x] **Timeouts**: Idle and connection timeouts.
- [x] **Standard I/O**: Piping stdin/stdout works correctly.

### Missing / Roadmap 🚧
- [ ] **Command Execution**: `-e` / `-c` flags to execute a program after connection (e.g., reverse shell).
- [ ] **Hex Dump**: `-x` flag to dump traffic in hex.
- [ ] **Proxy Support**: `-X` / `-x` for SOCKS/HTTP proxies.
- [ ] **Unix Domain Sockets**: `-U` support.
- [ ] **Telnet Negotiation**: `-t` support.
- [ ] **Daemon Mode**: `-d` to run in background.
