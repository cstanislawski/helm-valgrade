package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
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

	if cfg.Help {
		config.PrintHelp()
		os.Exit(0)
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

	diffResult, err := diff.Compare(baseChart, targetChart, nodeToMap(userValues), cfg.KeepValues, cfg.IgnoreMissing)
	if err != nil {
		return fmt.Errorf("failed to compare charts: %w", err)
	}

	upgradedValues, err := applyUpgrades(diffResult, userValues)
	if err != nil {
		return fmt.Errorf("failed to apply upgrades: %w", err)
	}

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

func applyUpgrades(diffResult *diff.Result, userValues *yaml.Node) (*yaml.Node, error) {
	for k, v := range diffResult.Added {
		if err := values.SetValue(userValues, fmt.Sprintf("%v", v), strings.Split(k, ".")...); err != nil {
			return nil, fmt.Errorf("failed to set added value %s: %w", k, err)
		}
	}

	for k, v := range diffResult.Modified {
		if err := values.SetValue(userValues, fmt.Sprintf("%v", v), strings.Split(k, ".")...); err != nil {
			return nil, fmt.Errorf("failed to set modified value %s: %w", k, err)
		}
	}

	for k := range diffResult.Removed {
		if err := values.DeleteValue(userValues, strings.Split(k, ".")...); err != nil {
			return nil, fmt.Errorf("failed to delete removed value %s: %w", k, err)
		}
	}

	return userValues, nil
}

func printUpgradedValues(upgradedValues *yaml.Node) error {
	return values.Write(os.Stdout.Name(), upgradedValues)
}

func writeOutput(upgradedValues *yaml.Node, outputFile string, inPlace bool, valuesFile string) error {
	if inPlace {
		outputFile = valuesFile
	}

	return values.Write(outputFile, upgradedValues)
}

func nodeToMap(node *yaml.Node) map[string]interface{} {
	var m map[string]interface{}
	err := node.Decode(&m)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert yaml.Node to map")
		return nil
	}
	return m
}
