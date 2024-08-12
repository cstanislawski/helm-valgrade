package values

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
		wantErr  bool
	}{
		{
			name:     "Valid YAML",
			content:  "key: value\nnestedKey:\n  subKey: subValue",
			filename: "test.yaml",
			wantErr:  false,
		},
		{
			name:     "Invalid YAML",
			content:  "invalid: : content",
			filename: "test.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-*"+filepath.Ext(tt.filename))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			got, err := Load(tmpfile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Load() returned nil, expected *yaml.Node")
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Write YAML",
			content: "key: value\nnestedKey:\n  subKey: subValue",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			var node yaml.Node
			err = yaml.Unmarshal([]byte(tt.content), &node)
			if err != nil {
				t.Fatal(err)
			}

			if err := Write(tmpfile.Name(), &node); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := Load(tmpfile.Name())
				if err != nil {
					t.Errorf("Failed to load written file: %v", err)
				}
				if got == nil {
					t.Errorf("Load() returned nil after Write()")
				} else {
					var gotContent, wantContent []byte
					gotContent, err = yaml.Marshal(got)
					if err != nil {
						t.Errorf("Failed to marshal loaded content: %v", err)
					}
					wantContent, err = yaml.Marshal(&node)
					if err != nil {
						t.Errorf("Failed to marshal original content: %v", err)
					}
					if string(gotContent) != string(wantContent) {
						t.Errorf("Written content = %s, want %s", string(gotContent), string(wantContent))
					}
				}
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	yamlContent := `
key: value
nestedKey:
  subKey: subValue
`
	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		keys    []string
		want    string
		wantErr bool
	}{
		{
			name:    "Get top-level key",
			keys:    []string{"key"},
			want:    "value",
			wantErr: false,
		},
		{
			name:    "Get nested key",
			keys:    []string{"nestedKey", "subKey"},
			want:    "subValue",
			wantErr: false,
		},
		{
			name:    "Get non-existent key",
			keys:    []string{"nonExistent"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValue(&node, tt.keys...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
	yamlContent := `
key: value
nestedKey:
  subKey: subValue
`
	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		keys     []string
		newValue string
		wantErr  bool
	}{
		{
			name:     "Set top-level key",
			keys:     []string{"key"},
			newValue: "newValue",
			wantErr:  false,
		},
		{
			name:     "Set nested key",
			keys:     []string{"nestedKey", "subKey"},
			newValue: "newSubValue",
			wantErr:  false,
		},
		{
			name:     "Set non-existent key",
			keys:     []string{"nonExistent"},
			newValue: "newValue",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetValue(&node, tt.newValue, tt.keys...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got, err := GetValue(&node, tt.keys...)
				if err != nil {
					t.Errorf("Failed to get value after setting: %v", err)
				}
				if got != tt.newValue {
					t.Errorf("SetValue() = %v, want %v", got, tt.newValue)
				}
			}
		})
	}
}
