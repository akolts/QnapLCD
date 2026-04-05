package sysinfo

import "testing"

func TestParseUptime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"minutes only", "300.50 150.25\n", "5m", false},
		{"hours and minutes", "7261.00 3000.00\n", "2h 1m", false},
		{"days hours minutes", "90061.00 45000.00\n", "1d 1h 1m", false},
		{"zero", "0.00 0.00\n", "0m", false},
		{"large uptime", "8640000.00 1000.00\n", "100d 0h 0m", false},
		{"invalid", "garbage", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUptime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{0, "0m"},
		{59, "0m"},
		{60, "1m"},
		{3599, "59m"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{86399, "23h 59m"},
		{86400, "1d 0h 0m"},
		{90061, "1d 1h 1m"},
		{604800, "7d 0h 0m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.secs)
			if got != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.secs, got, tt.want)
			}
		})
	}
}

func TestParseLoadAvg(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"normal", "0.05 0.10 0.15 1/234 5678\n", "0.05 0.10 0.15", false},
		{"high load", "12.50 10.25 8.75 5/500 12345\n", "12.50 10.25 8.75", false},
		{"too few fields", "0.05\n", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLoadAvg(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTrueNASVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		line1   string
		line2   string
		wantErr bool
	}{
		{
			"standard format",
			"TrueNAS-SCALE-25.10.1",
			"TrueNAS-SCALE", "25.10.1", false,
		},
		{
			"with codename",
			"TrueNAS-SCALE-25.10.1-Goldeye",
			"TrueNAS-SCALE-25.10.1", "Goldeye", false,
		},
		{
			"simple two parts",
			"TrueNAS-25.10",
			"TrueNAS", "25.10", false,
		},
		{
			"no dashes",
			"TrueNAS",
			"TrueNAS", "", false,
		},
		{
			"empty",
			"",
			"", "", true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l1, l2, err := parseTrueNASVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if l1 != tt.line1 || l2 != tt.line2 {
				t.Errorf("got (%q, %q), want (%q, %q)", l1, l2, tt.line1, tt.line2)
			}
		})
	}
}

func TestParseZPoolList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []PoolInfo
		wantErr bool
	}{
		{
			"single pool",
			"tank\t3.38T\t1.15T\tONLINE\n",
			[]PoolInfo{{Name: "tank", Size: "3.38T", Alloc: "1.15T", Health: "ONLINE"}},
			false,
		},
		{
			"multiple pools",
			"tank\t3.38T\t1.15T\tONLINE\nbackup\t1.00T\t500G\tDEGRADED\n",
			[]PoolInfo{
				{Name: "tank", Size: "3.38T", Alloc: "1.15T", Health: "ONLINE"},
				{Name: "backup", Size: "1.00T", Alloc: "500G", Health: "DEGRADED"},
			},
			false,
		},
		{
			"empty output",
			"",
			nil,
			false,
		},
		{
			"malformed line",
			"tank\t3.38T\n",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseZPoolList(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d pools, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("pool %d: got %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestPoolInfoDisplay(t *testing.T) {
	p := PoolInfo{Name: "tank", Size: "3.38T", Alloc: "1.15T", Health: "ONLINE"}

	if got := p.Line1(); got != "tank (ONLINE)" {
		t.Errorf("Line1() = %q, want %q", got, "tank (ONLINE)")
	}
	if got := p.Line2(); got != "1.15T of 3.38T" {
		t.Errorf("Line2() = %q, want %q", got, "1.15T of 3.38T")
	}
}

func TestHostname(t *testing.T) {
	// Just verify it returns a non-empty string.
	h := Hostname()
	if h == "" {
		t.Error("hostname is empty")
	}
}

func TestOSInfo(t *testing.T) {
	info := OSInfo()
	if info == "" {
		t.Error("os info is empty")
	}
	// Should contain the OS and arch.
	if !contains(info, "linux") && !contains(info, "darwin") && !contains(info, "windows") {
		t.Errorf("unexpected os info: %q", info)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
