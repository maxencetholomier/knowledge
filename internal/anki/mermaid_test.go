package anki

import "testing"

func TestIsBrowserLaunchFailure(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name: "chromium sandbox failure",
			output: `Error: Failed to launch the browser process:  Code: null
[FATAL:sandbox/linux/suid/client/setuid_sandbox_host.cc:166] The SUID sandbox helper ` +
				`binary was found, but is not configured correctly.`,
			want: true,
		},
		{
			name:   "mermaid syntax error",
			output: "Parse error on line 2:\n  A --< B\n     ^\nExpecting 'SEMI', got 'LT'",
			want:   false,
		},
		{
			name:   "successful render",
			output: "Generating single mermaid chart",
			want:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isBrowserLaunchFailure([]byte(test.output)); got != test.want {
				t.Errorf("isBrowserLaunchFailure() = %v, want %v", got, test.want)
			}
		})
	}
}
