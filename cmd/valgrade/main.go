package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"

	"github.com/cstanislawski/helm-valgrade/internal/chart"
	"github.com/cstanislawski/helm-valgrade/internal/config"
	"github.com/cstanislawski/helm-valgrade/internal/diff"
	"github.com/cstanislawski/helm-valgrade/internal/values"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel, cfg.Silent)

	if err := run(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute valgrade")
	}
}

func run(cfg *config.Config) error {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return fmt.Errorf("failed to initialize action configuration: %w", err)
	}

	baseChart, err := chart.Fetch(cfg.Repository, cfg.ChartName, cfg.VersionBase, actionConfig)
	if err != nil {
		return fmt.Errorf("failed to fetch base chart: %w", err)
	}

	targetChart, err := chart.Fetch(cfg.Repository, cfg.ChartName, cfg.VersionTarget, actionConfig)
	if err != nil {
		return fmt.Errorf("failed to fetch target chart: %w", err)
	}

	userValues, err := values.Load(cfg.ValuesFile)
	if err != nil {
		return fmt.Errorf("failed to load user values: %w", err)
	}

	diffResult, err := diff.Compare(baseChart, targetChart, userValues, cfg.KeepValues, cfg.IgnoreMissing)
	if err != nil {
		return fmt.Errorf("failed to compare charts: %w", err)
	}

	upgradedValues := applyUpgrades(diffResult, userValues)

	if cfg.DryRun {
		return printUpgradedValues(upgradedValues)
	}

	return writeOutput(upgradedValues, cfg.OutputFile, cfg.InPlace, cfg.ValuesFile)
}

func setupLogger(level string, silent bool) {
	if silent {
		zerolog.SetGlobalLevel(zerolog.Disabled)
		return
	}

	zerolog.SetGlobalLevel(getLogLevel(level))
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func getLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func applyUpgrades(diffResult *diff.Result, userValues map[string]interface{}) map[string]interface{} {
	upgradedValues := make(map[string]interface{})
	for k, v := range userValues {
		upgradedValues[k] = v
	}

	for k, v := range diffResult.Added {
		upgradedValues[k] = v
	}

	for k, v := range diffResult.Modified {
		upgradedValues[k] = v
	}

	for k := range diffResult.Removed {
		delete(upgradedValues, k)
	}

	return upgradedValues
}

func printUpgradedValues(upgradedValues map[string]interface{}) error {
	return values.Write(os.Stdout.Name(), upgradedValues)
}

func writeOutput(upgradedValues map[string]interface{}, outputFile string, inPlace bool, valuesFile string) error {
	if inPlace {
		outputFile = valuesFile
	}

	return values.Write(outputFile, upgradedValues)
}
