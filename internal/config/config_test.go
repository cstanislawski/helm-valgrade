package config

import (
	"flag"
	"os"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestParse_ValidInput(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--version-base=1.0.0",
		"--version-target=2.0.0",
		"--values=test.yaml",
		"--output-file=result.yaml",
		"--repository=https://charts.example.com",
		"--chart=mychart",
	}

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.VersionBase != "1.0.0" {
		t.Errorf("Expected version-base '1.0.0', got '%s'", cfg.VersionBase)
	}

	if cfg.VersionTarget != "2.0.0" {
		t.Errorf("Expected version-target '2.0.0', got '%s'", cfg.VersionTarget)
	}

	if cfg.ValuesFile != "test.yaml" {
		t.Errorf("Expected values file 'test.yaml', got '%s'", cfg.ValuesFile)
	}

	if cfg.OutputFile != "result.yaml" {
		t.Errorf("Expected output file 'result.yaml', got '%s'", cfg.OutputFile)
	}

	if cfg.Repository != "https://charts.example.com" {
		t.Errorf("Expected repository 'https://charts.example.com', got '%s'", cfg.Repository)
	}

	if cfg.ChartName != "mychart" {
		t.Errorf("Expected chart name 'mychart', got '%s'", cfg.ChartName)
	}
}

func TestParse_InPlaceFlag(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--version-base=1.0.0",
		"--version-target=2.0.0",
		"--values=test.yaml",
		"--in-place",
		"--repository=https://charts.example.com",
		"--chart=mychart",
	}

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !cfg.InPlace {
		t.Errorf("Expected in-place to be true, got false")
	}

	if cfg.OutputFile != "" {
		t.Errorf("Expected output file to be empty, got '%s'", cfg.OutputFile)
	}
}

func TestParse_MissingRequiredFlags(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--version-base=1.0.0",
		"--values=test.yaml",
		"--repository=https://charts.example.com",
		"--chart=mychart",
	}

	_, err := Parse()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "invalid configuration: version-target is required (use -t or --version-target)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestParse_InPlaceAndOutputFileTogether(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--version-base=1.0.0",
		"--version-target=2.0.0",
		"--values=test.yaml",
		"--in-place",
		"--output-file=result.yaml",
		"--repository=https://charts.example.com",
		"--chart=mychart",
	}

	_, err := Parse()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "in-place and output-file cannot be used together"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestParse_KeepValues(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--version-base=1.0.0",
		"--version-target=2.0.0",
		"--values=test.yaml",
		"--output-file=result.yaml",
		"--keep=foo,bar",
		"--repository=https://charts.example.com",
		"--chart=mychart",
	}

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedKeepValues := []string{"foo", "bar"}
	for i, value := range expectedKeepValues {
		if cfg.KeepValues[i] != value {
			t.Errorf("Expected KeepValues[%d] to be '%s', got '%s'", i, value, cfg.KeepValues[i])
		}
	}
}

func TestParse_DefaultLogLevel(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--version-base=1.0.0",
		"--version-target=2.0.0",
		"--values=test.yaml",
		"--output-file=result.yaml",
		"--repository=https://charts.example.com",
		"--chart=mychart",
	}

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got '%s'", cfg.LogLevel)
	}
}

func TestParse_CustomLogLevel(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--version-base=1.0.0",
		"--version-target=2.0.0",
		"--values=test.yaml",
		"--output-file=result.yaml",
		"--log-level=debug",
		"--repository=https://charts.example.com",
		"--chart=mychart",
	}

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.LogLevel)
	}
}
