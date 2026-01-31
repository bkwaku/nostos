package main

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal string
		envValue   string
		setEnv     bool
		want       string
	}{
		{
			name:       "returns env value when set",
			key:        "TEST_KEY",
			defaultVal: "default",
			envValue:   "custom",
			setEnv:     true,
			want:       "custom",
		},
		{
			name:       "returns default when env not set",
			key:        "TEST_KEY_UNSET",
			defaultVal: "default",
			setEnv:     false,
			want:       "default",
		},
		{
			name:       "returns default when env is empty string",
			key:        "TEST_KEY_EMPTY",
			defaultVal: "default",
			envValue:   "",
			setEnv:     true,
			want:       "default",
		},
		{
			name:       "returns empty default when both empty",
			key:        "TEST_KEY_BOTH_EMPTY",
			defaultVal: "",
			envValue:   "",
			setEnv:     true,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}
			got := getEnv(tt.key, tt.defaultVal)

			if got != tt.want {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defaultVal, got, tt.want)
			}
		})
	}
}
