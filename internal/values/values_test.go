package values

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "Valid YAML",
			content:  "key: value\nnestedKey:\n  subKey: subValue",
			filename: "test.yaml",
			want:     map[string]interface{}{"key": "value", "nestedKey": map[string]interface{}{"subKey": "subValue"}},
			wantErr:  false,
		},
		{
			name:     "Valid JSON",
			content:  `{"key": "value", "nestedKey": {"subKey": "subValue"}}`,
			filename: "test.json",
			want:     map[string]interface{}{"key": "value", "nestedKey": map[string]interface{}{"subKey": "subValue"}},
			wantErr:  false,
		},
		{
			name:     "Invalid format",
			content:  "Invalid content",
			filename: "test.txt",
			want:     nil,
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]interface{}
		ext     string
		wantErr bool
	}{
		{
			name:    "Write YAML",
			values:  map[string]interface{}{"key": "value", "nestedKey": map[string]interface{}{"subKey": "subValue"}},
			ext:     ".yaml",
			wantErr: false,
		},
		{
			name:    "Write JSON",
			values:  map[string]interface{}{"key": "value", "nestedKey": map[string]interface{}{"subKey": "subValue"}},
			ext:     ".json",
			wantErr: false,
		},
		{
			name:    "Invalid format",
			values:  map[string]interface{}{"key": "value"},
			ext:     ".txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-*"+tt.ext)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if err := Write(tmpfile.Name(), tt.values); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := Load(tmpfile.Name())
				if err != nil {
					t.Errorf("Failed to load written file: %v", err)
				}
				if !reflect.DeepEqual(got, tt.values) {
					t.Errorf("Written values = %v, want %v", got, tt.values)
				}
			}
		})
	}
}
