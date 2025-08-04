package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reilabs/trusted-setup/online/config"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    config.Config
		wantError bool
	}{
		{
			name: "valid config",
			config: config.Config{
				CeremonyName: "Test ceremony",
				Host:         "localhost",
				Port:         8080,
				R1cs:         "input.r1cs",
				Phase1:       "phase1.dat",
			},
			wantError: false,
		},
		{
			name: "missing name",
			config: config.Config{
				Port:   8080,
				R1cs:   "input.r1cs",
				Phase1: "phase1.dat",
			},
			wantError: true,
		},
		{
			name: "missing host",
			config: config.Config{
				CeremonyName: "Test ceremony",
				Port:         8080,
				R1cs:         "input.r1cs",
				Phase1:       "phase1.dat",
			},
			wantError: true,
		},
		{
			name: "port zero",
			config: config.Config{
				CeremonyName: "Test ceremony",
				Host:         "localhost",
				Port:         0,
				R1cs:         "input.r1cs",
				Phase1:       "phase1.dat",
			},
			wantError: true,
		},
		{
			name: "missing R1cs",
			config: config.Config{
				CeremonyName: "Test ceremony",
				Host:         "localhost",
				Port:         8080,
				Phase1:       "phase1.dat",
			},
			wantError: true,
		},
		{
			name: "missing Phase1",
			config: config.Config{
				CeremonyName: "Test ceremony",
				Host:         "localhost",
				Port:         8080,
				R1cs:         "input.r1cs",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.config.Validate()
				if (err != nil) != tt.wantError {
					t.Errorf("validate() error = %v, wantError %v", err, tt.wantError)
				}
			},
		)
	}
}

func TestNewConfig(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir := t.TempDir()

	validConfigContent := `{
		"ceremonyName": "Test ceremony",
		"host": "localhost",
		"port": 8080,
		"r1cs": "input.r1cs",
		"phase1": "phase1.dat"
	}`

	invalidConfigContent := `{
		"host": "",
		"port": 0,
		"r1cs": "",
		"phase1": ""
	}`

	invalidJSONContent := `{
		"host": "localhost",
		"port": 8080,
		"r1cs": "input.r1cs",
		"phase1": "phase1.dat",
	` // malformed JSON

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{name: "valid config file", content: validConfigContent, expectError: false},
		{name: "invalid config file (validation fail)", content: invalidConfigContent, expectError: true},
		{name: "invalid JSON file", content: invalidJSONContent, expectError: true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				filePath := filepath.Join(tempDir, tt.name+".json")
				err := os.WriteFile(filePath, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("failed to write temp config file: %v", err)
				}

				cfg, err := config.NewConfig(filePath)
				if (err != nil) != tt.expectError {
					t.Errorf("NewConfig() error = %v, expectError %v", err, tt.expectError)
				}
				if err == nil && cfg == nil {
					t.Errorf("NewConfig() returned nil config without error")
				}
			},
		)
	}

	t.Run(
		"non-existent file", func(t *testing.T) {
			_, err := config.NewConfig(filepath.Join(tempDir, "nonexistent.json"))
			if err == nil {
				t.Errorf("expected error when loading non-existent config file, got nil")
			}
		},
	)
}
