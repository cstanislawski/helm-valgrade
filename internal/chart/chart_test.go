package chart

import (
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	testRepositoryName = "test-repo"
	testRepositoryURL  = "https://test-repo.example.com"
	testChartName      = "test-chart"
	testVersion        = "1.0.0"
)

type mockChartFetcher struct{}

func (m *mockChartFetcher) Fetch(repository, name, version string, _ *action.Configuration) (*Chart, error) {
	return &Chart{
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    name,
				Version: version,
			},
			Values: map[string]interface{}{
				"key":          "value",
				"alertmanager": map[string]interface{}{},
				"grafana":      map[string]interface{}{},
			},
			Schema: []byte(`{"type": "object"}`),
		},
	}, nil
}

func TestFetch(t *testing.T) {
	fetcher := &mockChartFetcher{}

	chart, err := fetcher.Fetch(testRepositoryName, testChartName, testVersion, nil)
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
	fetcher := &mockChartFetcher{}

	chart, err := fetcher.Fetch(testRepositoryName, testChartName, testVersion, nil)
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
			t.Error("Schema is empty")
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

type mockRepoFile struct {
	repositories []*repo.Entry
}

func (m *mockRepoFile) Update(e *repo.Entry) bool {
	for i, r := range m.repositories {
		if r.Name == e.Name {
			m.repositories[i] = e
			return true
		}
	}
	m.repositories = append(m.repositories, e)
	return false
}

func (m *mockRepoFile) WriteFile(path string, perm os.FileMode) error {
	return nil
}

func TestUpdateRepository(t *testing.T) {
	mockRepo := &mockRepoFile{
		repositories: []*repo.Entry{
			{Name: testRepositoryName, URL: testRepositoryURL},
		},
	}

	err := updateRepositoryWithMock(testRepositoryName, testRepositoryURL, mockRepo)
	if err != nil {
		t.Fatalf("Failed to update repository: %v", err)
	}

	foundRepo := false
	for _, repo := range mockRepo.repositories {
		if repo.Name == testRepositoryName {
			foundRepo = true
			if repo.URL != testRepositoryURL {
				t.Errorf("Repository URL mismatch. Expected %s, got %s", testRepositoryURL, repo.URL)
			}
			break
		}
	}

	if !foundRepo {
		t.Errorf("Repository was not added to the mock repository file")
	}

	newRepoName := "new-repo"
	newRepoURL := "https://new-repo.example.com"
	err = updateRepositoryWithMock(newRepoName, newRepoURL, mockRepo)
	if err != nil {
		t.Fatalf("Failed to add new repository: %v", err)
	}

	foundNewRepo := false
	for _, repo := range mockRepo.repositories {
		if repo.Name == newRepoName {
			foundNewRepo = true
			if repo.URL != newRepoURL {
				t.Errorf("New repository URL mismatch. Expected %s, got %s", newRepoURL, repo.URL)
			}
			break
		}
	}

	if !foundNewRepo {
		t.Errorf("New repository was not added to the mock repository file")
	}
}

func updateRepositoryWithMock(name, url string, mockRepo *mockRepoFile) error {
	repoEntry := &repo.Entry{
		Name: name,
		URL:  url,
	}

	mockRepo.Update(repoEntry)
	return nil
}
