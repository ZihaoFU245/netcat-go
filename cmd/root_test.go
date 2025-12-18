package cmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseScanPort(t *testing.T) {
	makeRange := func(start, end int) []int {
		if start > end {
			return nil
		}
		r := make([]int, 0, end-start+1)
		for p := start; p <= end; p++ {
			r = append(r, p)
		}
		return r
	}

	tests := []struct {
		name      string
		args      []string
		flagRange string
		wantHost  string
		want      []int
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "flagRange single port",
			args:      []string{"example.com"},
			flagRange: "80",
			wantHost:  "example.com",
			want:      []int{80},
		},
		{
			name:      "flagRange two ports (split by :)",
			args:      []string{"example.com"},
			flagRange: "80:443",
			wantHost:  "example.com",
			want:      makeRange(80, 443),
		},
		{
			name:      "args only multiple ports",
			args:      []string{"example.com", "22", "80", "443"},
			flagRange: "",
			wantHost:  "example.com",
			want:      []int{22, 80, 443},
		},
		{
			name:      "deduplicate ports (flagRange and args overlap)",
			args:      []string{"example.com", "80", "22", "80"},
			flagRange: "80:443",
			wantHost:  "example.com",
			want:      append(makeRange(80, 443), 22), // preserves first-seen order
		},
		{
			name:      "whitespace trimming",
			args:      []string{" example.com ", "  22  ", " 80"},
			flagRange: " 443 :  8080 ",
			wantHost:  "example.com",
			want:      append(makeRange(443, 8080), []int{22, 80}...),
		},
		{
			name:      "missing host -> error",
			args:      nil,
			flagRange: "",
			wantErr:   true,
			errSubstr: "-z missing host",
		},
		{
			name:      "missing ports -> error",
			args:      []string{"example.com"},
			flagRange: "",
			wantErr:   true,
			errSubstr: "-z missing ports",
		},
		{
			name:      "invalid port in args -> error",
			args:      []string{"example.com", "99999"},
			flagRange: "",
			wantErr:   true,
			errSubstr: "port parsing failed",
		},
		{
			name:      "invalid flagRange format -> error (too many colons)",
			args:      []string{"example.com"},
			flagRange: "1:2:3",
			wantErr:   true,
			errSubstr: "port parsing failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, got, err := parseScanPort(tt.args, tt.flagRange)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil; host=%q ports=%v", host, got)
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("error=%q, expected to contain %q", err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if host != tt.wantHost {
				t.Fatalf("host=%q, want %q", host, tt.wantHost)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
