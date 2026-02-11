package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/api7/adc-go/internal/declarative"
	"gopkg.in/yaml.v3"
)

// LoadDeclarativeConfig loads configuration from YAML or JSON file
func LoadDeclarativeConfig(filename string) (*declarative.DeclarativeConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config declarative.DeclarativeConfig
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &config); err != nil {
			if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
				return nil, fmt.Errorf("failed to parse as YAML or JSON: YAML error: %v, JSON error: %v", err, jsonErr)
			}
		}
	}

	return &config, nil
}

// SaveDeclarativeConfig saves configuration to YAML or JSON file
func SaveDeclarativeConfig(config *declarative.DeclarativeConfig, filename string, format string) error {
	var data []byte
	var err error

	if format == "" {
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".json" {
			format = "json"
		} else {
			format = "yaml"
		}
	}

	switch format {
	case "json":
		data, err = json.MarshalIndent(config, "", "  ")
	case "yaml", "yml":
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported format: %s (use 'json' or 'yaml')", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadMultipleConfigs loads and merges multiple configuration files
func LoadMultipleConfigs(filenames []string) (*declarative.DeclarativeConfig, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("no configuration files specified")
	}

	merged := &declarative.DeclarativeConfig{
		Version: "1.0",
	}

	for _, filename := range filenames {
		config, err := LoadDeclarativeConfig(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", filename, err)
		}

		// Merge configurations
		merged.Routes = append(merged.Routes, config.Routes...)
		merged.Services = append(merged.Services, config.Services...)
		merged.Upstreams = append(merged.Upstreams, config.Upstreams...)
		merged.Consumers = append(merged.Consumers, config.Consumers...)
		merged.SSLs = append(merged.SSLs, config.SSLs...)
		merged.GlobalRules = append(merged.GlobalRules, config.GlobalRules...)
		merged.PluginConfigs = append(merged.PluginConfigs, config.PluginConfigs...)
		merged.StreamRoutes = append(merged.StreamRoutes, config.StreamRoutes...)
	}

	return merged, nil
}

// ExpandEnvVars expands environment variables in configuration
func ExpandEnvVars(data []byte) []byte {
	return []byte(os.ExpandEnv(string(data)))
}
