package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase2"
)

// appendTimestamp appends a timestamp to the given string if not present, or updates the timestamp if present.
//
// The timestamp is appended to the string in the format: `.YYYYMMDDHHMMSS.uuuuuu`, where `u` stands for microsecond.
// For an example `foo.bar` string, the new string will be `foo.bar.202506142137.569775`, provided updateTimestamp
// is called on June 14th, 2025 at 9:21:37.569775 PM CEST.
//
// Always succeeds. Returns the timestamped input string.
func appendTimestamp(str string) string {
	timestampRegex := regexp.MustCompile(`\.\d{14}.\d{6}$`)
	currentTimestamp := time.Now().Format("20060102150405.000000")

	if timestampRegex.MatchString(str) {
		return timestampRegex.ReplaceAllString(str, "."+currentTimestamp)
	}

	return str + "." + currentTimestamp
}

func Phase2Contribute(_ context.Context, cmd *cli.Command) error {
	phase2FilePath := cmd.String("phase2")
	log.Printf(
		"Contribution to Phase 2:\n"+
			"\tLoad Phase 2 from:    %s\n",
		phase2FilePath,
	)
	if phase2FilePath == "" {
		return fmt.Errorf("input Phase 2 file path is empty")
	}
	newFileName := appendTimestamp(phase2FilePath)
	err := phase2.Contribute(phase2FilePath, newFileName)
	if err != nil {
		return err
	}

	log.Printf("Phase2 file with contributions: %s", newFileName)
	return nil
}
