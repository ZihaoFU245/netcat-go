# Netcat implemented in Golang

## Usgae
```bash
netcat implemented in go.

Usage:
  nc [host] [port] [flags]

Flags:
  -h, --help             help for nc
  -j, --jobs int         Number of concurrency for -z port scan (default 3)
  -k, --keep-listening   accept loop
  -l, --listen           listen mode
  -n, --numeric-ip       Disable DNS lookup, only accept ip address
  -p, --port int         port number
  -z, --scan string      Scan a range of ports, [start]:[end], or 80 443 22 ...
  -s, --source string    specify source ip address
  -w, --time-outs int    Timeouts
  -u, --udp              UDP mode
  -v, --verbose          verbose mode
```