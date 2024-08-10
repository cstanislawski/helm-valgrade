package config

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	VersionBase   string
	VersionTarget string
	ValuesFile    string
	OutputFile    string
	InPlace       bool
	Repository    string
	ChartName     string
	KeepValues    []string
	Silent        bool
	LogLevel      string
	DryRun        bool
	IgnoreMissing bool
}

func Parse() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.VersionBase, "version-base", "", "The version of the chart you are upgrading from")
	flag.StringVar(&cfg.VersionBase, "b", "", "The version of the chart you are upgrading from (shorthand)")

	flag.StringVar(&cfg.VersionTarget, "version-target", "", "The version of the chart you are upgrading to")
	flag.StringVar(&cfg.VersionTarget, "t", "", "The version of the chart you are upgrading to (shorthand)")

	flag.StringVar(&cfg.ValuesFile, "values", "", "The path to the values file you are using")
	flag.StringVar(&cfg.ValuesFile, "f", "", "The path to the values file you are using (shorthand)")

	flag.StringVar(&cfg.OutputFile, "output-file", "", "The path to the output file")
	flag.StringVar(&cfg.OutputFile, "o", "", "The path to the output file (shorthand)")

	flag.BoolVar(&cfg.InPlace, "in-place", false, "Update the values file in place")
	flag.BoolVar(&cfg.InPlace, "i", false, "Update the values file in place (shorthand)")

	flag.StringVar(&cfg.Repository, "repository", "", "The repository where the chart is located")
	flag.StringVar(&cfg.Repository, "r", "", "The repository where the chart is located (shorthand)")

	flag.StringVar(&cfg.ChartName, "chart", "", "The name of the chart")
	flag.StringVar(&cfg.ChartName, "c", "", "The name of the chart (shorthand)")

	var keepValues string
	flag.StringVar(&keepValues, "keep", "", "Exclude specific values from the upgrade process (comma-separated)")
	flag.StringVar(&keepValues, "k", "", "Exclude specific values from the upgrade process (comma-separated) (shorthand)")

	flag.BoolVar(&cfg.Silent, "silent", false, "Suppress all output")
	flag.BoolVar(&cfg.Silent, "s", false, "Suppress all output (shorthand)")

	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Set the log level (debug, info, warn, error, fatal)")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Set the log level (debug, info, warn, error, fatal) (shorthand)")

	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Print the result without writing to the output file")
	flag.BoolVar(&cfg.DryRun, "d", false, "Print the result without writing to the output file (shorthand)")

	flag.BoolVar(&cfg.IgnoreMissing, "ignore-missing", false, "Ignore missing values in the old chart version")

	flag.Parse()

	if keepValues != "" {
		cfg.KeepValues = strings.Split(keepValues, ",")
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) validate() error {
	if cfg.VersionBase == "" {
		return fmt.Errorf("version-base is required")
	}
	if cfg.VersionTarget == "" {
		return fmt.Errorf("version-target is required")
	}
	if cfg.ValuesFile == "" {
		return fmt.Errorf("values file is required")
	}
	if !cfg.InPlace && cfg.OutputFile == "" {
		return fmt.Errorf("either in-place or output-file must be specified")
	}
	if cfg.InPlace && cfg.OutputFile != "" {
		return fmt.Errorf("in-place and output-file cannot be used together")
	}
	return nil
}
