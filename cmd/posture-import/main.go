// Command posture-import converts reviewed furniture posture evidence into a
// reversible Liquibase seed.
package main

import (
	"flag"
	"fmt"
	"os"
)

// config contains the posture import input and output paths.
type config struct {
	reviewsPath string
	outputPath  string
}

// main parses command-line flags and runs the posture import.
func main() {
	settings := config{}
	flag.StringVar(&settings.reviewsPath, "reviews", "", "path to the posture review manifest")
	flag.StringVar(&settings.outputPath, "output", "", "path to the generated Liquibase SQL")
	flag.Parse()
	if err := run(settings); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run loads, validates, renders, and writes one posture import.
func run(settings config) error {
	if settings.reviewsPath == "" {
		return fmt.Errorf("-reviews is required")
	}
	if settings.outputPath == "" {
		return fmt.Errorf("-output is required")
	}
	manifest, err := loadManifest(settings.reviewsPath)
	if err != nil {
		return err
	}
	if err = os.WriteFile(settings.outputPath, renderSQL(manifest), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
