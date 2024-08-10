package chart

import (
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	testRepositoryName = "prometheus-community"
	testRepositoryURL  = "https://prometheus-community.github.io/helm-charts"
	testChartName      = "kube-prometheus-stack"
	testVersion        = "45.7.1"
)

func setupTestEnvironment(t *testing.T) (*action.Configuration, *cli.EnvSettings) {
	t.Helper()

	settings := cli.New()
	actionConfig := new(action.Configuration)

	tempDir, err := os.MkdirTemp("", "helm-valgrade-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	settings.RepositoryConfig = tempDir + "/repositories.yaml"
	settings.RepositoryCache = tempDir + "/repository-cache"

	if err := os.WriteFile(settings.RepositoryConfig, []byte("apiVersion: v1\nrepositories: []"), 0644); err != nil {
		t.Fatalf("Failed to create empty repository file: %v", err)
	}

	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), t.Logf); err != nil {
		t.Fatalf("Failed to initialize action configuration: %v", err)
	}

	return actionConfig, settings
}

func TestFetch(t *testing.T) {
	actionConfig, settings := setupTestEnvironment(t)

	if err := UpdateRepository(testRepositoryName, testRepositoryURL, settings); err != nil {
		t.Fatalf("Failed to update repository: %v", err)
	}

	chart, err := Fetch(testRepositoryName, testChartName, testVersion, actionConfig)
	if err != nil {
		t.Fatalf("Failed to fetch chart: %v", err)
	}

	if chart == nil {
		t.Fatal("Fetched chart is nil")
	}

	if chart.GetName() != testChartName {
		t.Errorf("Expected chart name %s, got %s", testChartName, chart.GetName())
	}

	if chart.GetVersion() != testVersion {
		t.Errorf("Expected chart version %s, got %s", testVersion, chart.GetVersion())
	}
}

func TestChartMethods(t *testing.T) {
	actionConfig, settings := setupTestEnvironment(t)

	if err := UpdateRepository(testRepositoryName, testRepositoryURL, settings); err != nil {
		t.Fatalf("Failed to update repository: %v", err)
	}

	chart, err := Fetch(testRepositoryName, testChartName, testVersion, actionConfig)
	if err != nil {
		t.Fatalf("Failed to fetch chart: %v", err)
	}

	t.Run("GetDefaultValues", func(t *testing.T) {
		values := chart.GetDefaultValues()
		if values == nil {
			t.Error("Default values are nil")
		}
		if _, ok := values["alertmanager"]; !ok {
			t.Error("Expected 'alertmanager' in default values")
		}
		if _, ok := values["grafana"]; !ok {
			t.Error("Expected 'grafana' in default values")
		}
	})

	t.Run("GetSchema", func(t *testing.T) {
		schema := chart.GetSchema()
		if len(schema) == 0 {
			t.Log("Schema is empty for kube-prometheus-stack")
		} else {
			t.Logf("Schema length: %d bytes", len(schema))
		}
	})

	t.Run("GetVersion", func(t *testing.T) {
		version := chart.GetVersion()
		if version != testVersion {
			t.Errorf("Expected version %s, got %s", testVersion, version)
		}
	})

	t.Run("GetName", func(t *testing.T) {
		name := chart.GetName()
		if name != testChartName {
			t.Errorf("Expected name %s, got %s", testChartName, name)
		}
	})
}

func TestUpdateRepository(t *testing.T) {
	_, settings := setupTestEnvironment(t)

	err := UpdateRepository(testRepositoryName, testRepositoryURL, settings)
	if err != nil {
		t.Fatalf("Failed to update repository: %v", err)
	}

	if _, err := os.Stat(settings.RepositoryConfig); os.IsNotExist(err) {
		t.Errorf("Repository file was not created")
	}

	repoFile, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		t.Fatalf("Failed to load repository file: %v", err)
	}

	found := false
	for _, repo := range repoFile.Repositories {
		if repo.Name == testRepositoryName && repo.URL == testRepositoryURL {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Test repository was not added to the repository file")
	}
}
