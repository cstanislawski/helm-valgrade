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
	Help          bool
}

func Parse() (*Config, error) {
	cfg := &Config{
		LogLevel: "info",
	}

	flag.Usage = PrintHelp

	flag.StringVar(&cfg.VersionBase, "version-base", "", "")
	flag.StringVar(&cfg.VersionBase, "b", "", "")
	flag.StringVar(&cfg.VersionTarget, "version-target", "", "")
	flag.StringVar(&cfg.VersionTarget, "t", "", "")
	flag.StringVar(&cfg.ValuesFile, "values", "", "")
	flag.StringVar(&cfg.ValuesFile, "f", "", "")
	flag.StringVar(&cfg.OutputFile, "output-file", "", "")
	flag.StringVar(&cfg.OutputFile, "o", "", "")
	flag.BoolVar(&cfg.InPlace, "in-place", false, "")
	flag.BoolVar(&cfg.InPlace, "i", false, "")
	flag.StringVar(&cfg.Repository, "repository", "", "")
	flag.StringVar(&cfg.Repository, "r", "", "")
	flag.StringVar(&cfg.ChartName, "chart", "", "")
	flag.StringVar(&cfg.ChartName, "c", "", "")
	flag.Var((*stringSliceFlag)(&cfg.KeepValues), "keep", "")
	flag.Var((*stringSliceFlag)(&cfg.KeepValues), "k", "")
	flag.BoolVar(&cfg.Silent, "silent", false, "")
	flag.BoolVar(&cfg.Silent, "s", false, "")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "")
	flag.StringVar(&cfg.LogLevel, "l", "info", "")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "")
	flag.BoolVar(&cfg.DryRun, "d", false, "")
	flag.BoolVar(&cfg.IgnoreMissing, "ignore-missing", false, "")
	flag.BoolVar(&cfg.Help, "help", false, "")
	flag.BoolVar(&cfg.Help, "h", false, "")

	flag.Parse()

	return cfg, cfg.validate()
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
	if cfg.InPlace && cfg.OutputFile != "" {
		return fmt.Errorf("in-place and output-file cannot be used together")
	}
	if cfg.Repository == "" {
		errors = append(errors, "repository is required (use -r or --repository)")
	}
	if cfg.ChartName == "" {
		errors = append(errors, "chart name is required (use -c or --chart)")
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid configuration: %s", errors[0])
	}

	return nil
}

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, strings.Split(value, ",")...)
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
