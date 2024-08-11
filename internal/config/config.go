package config

import (
	"fmt"
	"os"
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
	Help          bool
}

func Parse() (*Config, error) {
	cfg := &Config{
		LogLevel: "info",
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			cfg.Help = true
			return cfg, nil
		case "-b", "--version-base":
			if i+1 < len(args) {
				cfg.VersionBase = args[i+1]
				i++
			}
		case "-t", "--version-target":
			if i+1 < len(args) {
				cfg.VersionTarget = args[i+1]
				i++
			}
		case "-f", "--values":
			if i+1 < len(args) {
				cfg.ValuesFile = args[i+1]
				i++
			}
		case "-o", "--output-file":
			if i+1 < len(args) {
				cfg.OutputFile = args[i+1]
				i++
			}
		case "-i", "--in-place":
			cfg.InPlace = true
		case "-r", "--repository":
			if i+1 < len(args) {
				cfg.Repository = args[i+1]
				i++
			}
		case "-c", "--chart":
			if i+1 < len(args) {
				chartNameParts := []string{args[i+1]}
				for j := i + 2; j < len(args); j++ {
					if strings.HasPrefix(args[j], "-") {
						break
					}
					chartNameParts = append(chartNameParts, args[j])
					i = j
				}
				cfg.ChartName = strings.Join(chartNameParts, " ")
			}
		case "-k", "--keep":
			if i+1 < len(args) {
				cfg.KeepValues = append(cfg.KeepValues, strings.Split(args[i+1], ",")...)
				i++
			}
		case "-s", "--silent":
			cfg.Silent = true
		case "-l", "--log-level":
			if i+1 < len(args) {
				cfg.LogLevel = args[i+1]
				i++
			}
		case "-d", "--dry-run":
			cfg.DryRun = true
		case "--ignore-missing":
			cfg.IgnoreMissing = true
		}
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) validate() error {
	if cfg.Help {
		return nil
	}

	var errors []string

	if cfg.VersionBase == "" {
		errors = append(errors, "version-base is required (use -b or --version-base)")
	}
	if cfg.VersionTarget == "" {
		errors = append(errors, "version-target is required (use -t or --version-target)")
	}
	if cfg.ValuesFile == "" {
		errors = append(errors, "values file is required (use -f or --values)")
	}
	if !cfg.InPlace && cfg.OutputFile == "" {
		errors = append(errors, "either in-place (-i) or output-file (-o) must be specified")
	}
	if cfg.Repository == "" {
		errors = append(errors, "repository is required (use -r or --repository)")
	}
	if cfg.ChartName == "" {
		errors = append(errors, "chart name is required (use -c or --chart)")
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid configuration:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

func PrintHelp() {
	fmt.Println("Usage: helm valgrade [flags]")
	fmt.Println("\nFlags:")
	fmt.Println("  -b, --version-base string    The version of the chart you are upgrading from")
	fmt.Println("  -t, --version-target string  The version of the chart you are upgrading to")
	fmt.Println("  -f, --values string          The path to the values file you are using")
	fmt.Println("  -o, --output-file string     The path to the output file")
	fmt.Println("  -i, --in-place               Update the values file in place")
	fmt.Println("  -r, --repository string      The repository where the chart is located")
	fmt.Println("  -c, --chart string           The name of the chart")
	fmt.Println("  -k, --keep string            Exclude specific values from the upgrade process (comma-separated)")
	fmt.Println("  -s, --silent                 Suppress all output")
	fmt.Println("  -l, --log-level string       Set the log level (debug, info, warn, error, fatal) (default \"info\")")
	fmt.Println("  -d, --dry-run                Print the result without writing to the output file")
	fmt.Println("      --ignore-missing         Ignore missing values in the old chart version")
	fmt.Println("  -h, --help                   Display this help message")
}
