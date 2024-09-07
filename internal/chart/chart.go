package chart

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type Chart struct {
	*chart.Chart
	NormalizedValues map[string]interface{}
}

func Fetch(repository, name, version string, actionConfig *action.Configuration) (*Chart, error) {
	settings := cli.New()

	if actionConfig == nil {
		actionConfig = new(action.Configuration)
		if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), logf); err != nil {
			return nil, fmt.Errorf("failed to initialize action configuration: %w", err)
		}
	}

	client := action.NewInstall(actionConfig)
	client.ChartPathOptions.RepoURL = repository
	client.ChartPathOptions.Version = version

	chartDownloader := downloader.ChartDownloader{
		Out:              os.Stdout,
		Verify:           downloader.VerifyNever,
		Getters:          getter.All(settings),
		Options:          []getter.Option{},
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
	}

	chartURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(repository, "/"), strings.ReplaceAll(name, " ", "-"))
	filename, _, err := chartDownloader.DownloadTo(chartURL, version, settings.RepositoryCache)
	if err != nil {
		return nil, fmt.Errorf("failed to download chart: %w", err)
	}

	loadedChart, err := loader.Load(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	normalizedValues := normalizeValues(loadedChart.Values)

	return &Chart{
		Chart:            loadedChart,
		NormalizedValues: normalizedValues,
	}, nil
}

func (c *Chart) GetDefaultValues() map[string]interface{} {
	return c.Values
}

func (c *Chart) GetDefaultValuesYamlNode() *yaml.Node {
	data, err := yaml.Marshal(c.Values)
	if err != nil {
		return nil
	}

	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		return nil
	}

	return &node
}

func normalizeValues(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = normalizeValue(v)
	}
	return result
}

func normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case float64:
		if val == float64(int(val)) {
			return int(val)
		}
	case []interface{}:
		normalized := make([]interface{}, len(val))
		for i, item := range val {
			normalized[i] = normalizeValue(item)
		}
		return normalized
	case map[string]interface{}:
		return normalizeValues(val)
	}
	return v
}

func (c *Chart) GetSchema() []byte {
	return c.Schema
}

func (c *Chart) GetVersion() string {
	return c.Metadata.Version
}

func (c *Chart) GetName() string {
	return c.Metadata.Name
}

func UpdateRepository(name, url string, settings *cli.EnvSettings) error {
	repoFile, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		if os.IsNotExist(err) {
			repoFile = repo.NewFile()
		} else {
			return fmt.Errorf("failed to load repository file: %w", err)
		}
	}

	repoEntry := &repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("failed to download repository index: %w", err)
	}

	repoFile.Update(repoEntry)

	if err := repoFile.WriteFile(settings.RepositoryConfig, 0644); err != nil {
		return fmt.Errorf("failed to write updated repository file: %w", err)
	}

	return nil
}

func logf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
