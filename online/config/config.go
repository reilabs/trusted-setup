package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Config struct {
	CeremonyName string `json:"ceremonyName"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	R1cs         string `json:"r1cs"`
	Phase1       string `json:"phase1"`
}

func (c *Config) Validate() error {
	const msg = "config validation error: "
	if c.CeremonyName == "" {
		return fmt.Errorf(msg + "ceremony name must be provided")
	}
	if c.Host == "" {
		return fmt.Errorf(msg + "host must be provided")
	}
	if c.Port == 0 {
		return fmt.Errorf(msg + "port must be provided and non-zero")
	}
	if c.R1cs == "" {
		return fmt.Errorf(msg + "R1CS input path must be provided")
	}
	if c.Phase1 == "" {
		return fmt.Errorf(msg + "Phase 1 input path must be provided")
	}
	return nil
}

func New(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("error closing config file: %v", err)
		}
	}(file)

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(byteValue, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err = config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}
