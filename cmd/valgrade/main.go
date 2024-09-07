package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

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

	if cfg.Help {
		config.PrintHelp()
		os.Exit(0)
	}

	setupLogger(cfg.LogLevel, cfg.Silent)

	errors := run(cfg)
	if len(errors) > 0 {
		for _, err := range errors {
			log.Error().Msgf("%v", err)
		}
		log.Error().Msg("Failed to execute valgrade")
		os.Exit(1)
	}

	log.Info().Msg("Valgrade completed successfully")
}

func run(cfg *config.Config) []error {
	var errors []error

	baseChart, err := chart.Fetch(cfg.Repository, cfg.ChartName, cfg.VersionBase, nil)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to fetch base chart: %w", err))
		return errors
	}

	targetChart, err := chart.Fetch(cfg.Repository, cfg.ChartName, cfg.VersionTarget, nil)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to fetch target chart: %w", err))
		return errors
	}

	targetValues := targetChart.GetDefaultValuesYamlNode()

	userValues, err := values.Load(cfg.ValuesFile)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to load user values: %w", err))
		return errors
	}

	userValuesMap := make(map[string]interface{})
	if err := userValues.Decode(&userValuesMap); err != nil {
		errors = append(errors, fmt.Errorf("failed to decode user values: %w", err))
		return errors
	}

	diffResult, err := diff.Compare(baseChart, targetChart, userValuesMap, cfg.KeepValues, cfg.IgnoreMissing)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to compare charts: %w", err))
		return errors
	}

	upgradedValues, upgradeErrors := applyUpgrades(diffResult, targetValues)
	if len(upgradeErrors) > 0 {
		for _, err := range upgradeErrors {
			errors = append(errors, fmt.Errorf("failed to apply upgrades: %w", err))
		}
		return errors
	}

	if cfg.DryRun {
		if err := printUpgradedValues(upgradedValues); err != nil {
			errors = append(errors, fmt.Errorf("failed to print upgraded values: %w", err))
		}
		return errors
	}

	if err := writeOutput(upgradedValues, cfg.OutputFile, cfg.InPlace, cfg.ValuesFile); err != nil {
		errors = append(errors, fmt.Errorf("failed to write output: %w", err))
	}

	return errors
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

func applyUpgrades(diffResult *diff.Diff, targetValues *yaml.Node) (*yaml.Node, []error) {
	var errors []error

	for k, v := range diffResult.Added {
		if err := values.SetValue(targetValues, v, strings.Split(k, ".")...); err != nil {
			errors = append(errors, fmt.Errorf("failed to set added value %s: %w", k, err))
		}
	}

	for k, v := range diffResult.Modified {
		if err := values.SetValue(targetValues, v, strings.Split(k, ".")...); err != nil {
			errors = append(errors, fmt.Errorf("failed to set modified value %s: %w", k, err))
		}
	}

	for k := range diffResult.Removed {
		if err := values.DeleteValue(targetValues, strings.Split(k, ".")...); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete removed value %s: %w", k, err))
		}
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return targetValues, nil
}

func printUpgradedValues(upgradedValues *yaml.Node) error {
	return values.Write(os.Stdout.Name(), upgradedValues)
}

func writeOutput(upgradedValues *yaml.Node, outputFile string, inPlace bool, valuesFile string) error {
	var targetFile string
	if inPlace {
		targetFile = valuesFile
		log.Info().Msg("Updating values file in-place")
	} else {
		targetFile = outputFile
		log.Info().Str("file", outputFile).Msg("Writing updated values to new file")
	}

	err := values.Write(targetFile, upgradedValues)
	if err != nil {
		return fmt.Errorf("failed to write updated values: %w", err)
	}

	log.Info().Str("file", targetFile).Msg("Successfully wrote updated values")
	return nil
}
