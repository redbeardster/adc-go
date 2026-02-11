package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Global variables
var (
	cfgFile string
	debug   bool
)

// ADCConfig represents the ADC configuration
type ADCConfig struct {
	APISIX  AdminAPI `yaml:"apisix"`
	Debug   bool     `yaml:"debug"`
	Version string   `yaml:"version"`
}

// AdminAPI represents APISIX admin API configuration
type AdminAPI struct {
	BaseURL      string `yaml:"base_url"`
	AdminKey     string `yaml:"admin_key"`
	AdminKeyName string `yaml:"admin_key_name,omitempty"`
}

// DeclarativeConfig represents the declarative configuration
type DeclarativeConfig struct {
	Version       string                 `yaml:"version"`
	Routes        []Route                `yaml:"routes,omitempty"`
	Services      []Service              `yaml:"services,omitempty"`
	Upstreams     []Upstream             `yaml:"upstreams,omitempty"`
	Consumers     []Consumer             `yaml:"consumers,omitempty"`
	SSLs          []SSL                  `yaml:"ssls,omitempty"`
	GlobalRules   []GlobalRule           `yaml:"global_rules,omitempty"`
	PluginConfigs []PluginConfig         `yaml:"plugin_configs,omitempty"`
	Metadata      map[string]interface{} `yaml:"metadata,omitempty"`
}

// Route represents an APISIX route
type Route struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name,omitempty"`
	Desc        string                 `yaml:"desc,omitempty"`
	URI         string                 `yaml:"uri,omitempty"`
	URIs        []string               `yaml:"uris,omitempty"`
	Upstream    *Upstream              `yaml:"upstream,omitempty"`
	UpstreamID  string                 `yaml:"upstream_id,omitempty"`
	Plugins     map[string]interface{} `yaml:"plugins,omitempty"`
}

// Service represents an APISIX service
type Service struct {
	ID         string                 `yaml:"id"`
	Name       string                 `yaml:"name,omitempty"`
	Desc       string                 `yaml:"desc,omitempty"`
	UpstreamID string                 `yaml:"upstream_id,omitempty"`
	Plugins    map[string]interface{} `yaml:"plugins,omitempty"`
}

// Consumer represents an APISIX consumer
type Consumer struct {
	Username string                 `yaml:"username"`
	Desc     string                 `yaml:"desc,omitempty"`
	Plugins  map[string]interface{} `yaml:"plugins,omitempty"`
}

// SSL represents an SSL certificate
type SSL struct {
	ID     string            `yaml:"id"`
	Desc   string            `yaml:"desc,omitempty"`
	Cert   string            `yaml:"cert"`
	Key    string            `yaml:"key"`
	Sni    []string          `yaml:"snis,omitempty"`
	Labels map[string]string `yaml:"labels,omitempty"`
}

// Upstream represents an upstream
type Upstream struct {
	ID      string                 `yaml:"id"`
	Name    string                 `yaml:"name,omitempty"`
	Type    string                 `yaml:"type,omitempty"`
	HashOn  string                 `yaml:"hash_on,omitempty"`
	Key     string                 `yaml:"key,omitempty"`
	Nodes   map[string]int         `yaml:"nodes"`
	Retries *int                   `yaml:"retries,omitempty"`
	Timeout map[string]interface{} `yaml:"timeout,omitempty"`
}

// GlobalRule represents a global rule
type GlobalRule struct {
	ID      string                 `yaml:"id"`
	Plugins map[string]interface{} `yaml:"plugins"`
}

// PluginConfig represents a plugin configuration
type PluginConfig struct {
	ID      string                 `yaml:"id"`
	Desc    string                 `yaml:"desc,omitempty"`
	Plugins map[string]interface{} `yaml:"plugins"`
}

// Execute is the main entry point for commands
func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "adc",
		Short: "API Declarative CLI for APISIX",
		Long: `ADC (API Declarative CLI) is a command-line tool for managing
APISIX configuration declaratively using YAML/JSON files.`,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/adc/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")

	// Add all commands
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newApplyCommand())
	rootCmd.AddCommand(newDiffCommand())
	rootCmd.AddCommand(newSyncCommand())

	return rootCmd.Execute()
}

// loadConfig loads configuration from file or environment
// skipValidation can be set to true for commands that don't need APISIX connection
func loadConfig(skipValidation bool) (*ADCConfig, error) {
	viper.SetEnvPrefix("ADC")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("apisix.base_url", "http://127.0.0.1:9180")
	viper.SetDefault("apisix.admin_key_name", "X-API-Key")
	viper.SetDefault("debug", false)
	viper.SetDefault("version", "1.0.0")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config in standard locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".config", "adc"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Try to read config, but don't fail if not found
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only fail on actual read errors, not on missing file
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config not found is OK for commands that don't need it
	}

	var config ADCConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override debug flag if set in command line
	if debug {
		config.Debug = true
	}

	// Skip validation for commands that don't need APISIX connection
	if !skipValidation {
		// Check required fields for APISIX connection
		if config.APISIX.AdminKey == "" {
			adminKey := os.Getenv("ADC_APISIX_ADMIN_KEY")
			if adminKey == "" {
				return nil, fmt.Errorf("admin key is required. Set it in config file or ADC_APISIX_ADMIN_KEY environment variable")
			}
			config.APISIX.AdminKey = adminKey
		}
	}

	return &config, nil
}

// versionCommand prints version information
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ADC Version: dev")
			fmt.Println("Build Time: 2026-02-11_08:52:39")
			fmt.Println("Git Commit: unknown")

			// Try to load config, but don't fail if not found
			if config, err := loadConfig(true); err == nil {
				fmt.Printf("Config Version: %s\n", config.Version)
				fmt.Printf("APISIX URL: %s\n", config.APISIX.BaseURL)
			}
		},
	}
}

// initCommand initializes ADC configuration
func newInitCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize ADC configuration",
		Long: `Initialize ADC configuration by creating a sample config file
and setting up the configuration directory.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing configuration")

	return cmd
}

func initializeConfig(force bool) error {
	// Create config directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "adc")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("config file already exists at %s. Use --force to overwrite", configPath)
	}

	// Create example configuration
	exampleConfig := ADCConfig{
		Version: "1.0.0",
		Debug:   false,
		APISIX: AdminAPI{
			BaseURL:      "http://127.0.0.1:9180",
			AdminKey:     "edd1c9f034335f136f87ad84b625c8f1",
			AdminKeyName: "X-API-Key",
		},
	}

	data, err := yaml.Marshal(&exampleConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Create example declarative config
	exampleDeclarative := `version: "1.0"
routes:
  - id: "route-1"
    name: "Example Route"
    uri: "/api/v1/*"
    upstream:
      id: "upstream-1"
      nodes:
        "httpbin.org:80": 1
    plugins:
      cors: {}

services:
  - id: "service-1"
    name: "Example Service"
    upstream_id: "upstream-1"

consumers:
  - username: "user-1"
    plugins:
      key-auth:
        key: "auth-key-123"

ssls:
  - id: "ssl-1"
    cert: "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
    key: "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
    snis:
      - "example.com"
`

	examplePath := filepath.Join(".", "example-config.yaml")
	if err := os.WriteFile(examplePath, []byte(exampleDeclarative), 0644); err != nil {
		return fmt.Errorf("failed to write example config: %w", err)
	}

	fmt.Println("✓ ADC initialized successfully!")
	fmt.Printf("  Config file created: %s\n", configPath)
	fmt.Printf("  Example declarative config created: %s\n", examplePath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the config file and set your APISIX admin key")
	fmt.Println("  2. Review the example configuration")
	fmt.Println("  3. Run 'adc validate -f example-config.yaml' to test")
	fmt.Println("  4. Run 'adc apply -f example-config.yaml' to apply")

	return nil
}

// validateCommand validates declarative configuration
func newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate -f FILE",
		Short: "Validate declarative configuration file",
		Long:  `Validate syntax and semantics of declarative configuration file.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			if file == "" {
				return fmt.Errorf("file is required. Use -f flag to specify configuration file")
			}

			return validateConfigFile(file)
		},
	}

	cmd.Flags().StringP("file", "f", "", "Declarative configuration file (YAML/JSON)")
	cmd.MarkFlagRequired("file")

	return cmd
}

func validateConfigFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config DeclarativeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Check required fields
	if config.Version == "" {
		return fmt.Errorf("version is required in declarative config")
	}

	// Basic resource validation
	validationErrors := []string{}

	for i, route := range config.Routes {
		if route.ID == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("route at index %d: id is required", i))
		}
		if route.URI == "" && len(route.URIs) == 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("route %s: uri or uris is required", route.ID))
		}
		if route.Upstream == nil && route.UpstreamID == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("route %s: upstream or upstream_id is required", route.ID))
		}
	}

	for i, service := range config.Services {
		if service.ID == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("service at index %d: id is required", i))
		}
	}

	for i, consumer := range config.Consumers {
		if consumer.Username == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("consumer at index %d: username is required", i))
		}
	}

	for i, ssl := range config.SSLs {
		if ssl.ID == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("ssl at index %d: id is required", i))
		}
		if ssl.Cert == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("ssl %s: cert is required", ssl.ID))
		}
		if ssl.Key == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("ssl %s: key is required", ssl.ID))
		}
	}

	if len(validationErrors) > 0 {
		fmt.Println("Validation errors found:")
		for _, err := range validationErrors {
			fmt.Printf("  ✗ %s\n", err)
		}
		return fmt.Errorf("configuration validation failed")
	}

	// Count resources
	upstreamCount := countUpstreams(&config)

	fmt.Printf("✓ Configuration file %s is valid\n", file)
	fmt.Printf("  Version: %s\n", config.Version)
	fmt.Printf("  Routes: %d\n", len(config.Routes))
	fmt.Printf("  Services: %d\n", len(config.Services))
	fmt.Printf("  Upstreams: %d\n", upstreamCount)
	fmt.Printf("  Consumers: %d\n", len(config.Consumers))
	fmt.Printf("  SSLs: %d\n", len(config.SSLs))

	// Show route details
	if len(config.Routes) > 0 {
		fmt.Println("\n  Routes details:")
		for _, route := range config.Routes {
			upstreamInfo := "no upstream"
			if route.Upstream != nil {
				upstreamInfo = fmt.Sprintf("upstream: %s", route.Upstream.ID)
			} else if route.UpstreamID != "" {
				upstreamInfo = fmt.Sprintf("upstream_id: %s", route.UpstreamID)
			}
			fmt.Printf("    - %s: %s (%s)\n", route.ID, route.Name, upstreamInfo)
		}
	}

	return nil
}

func countUpstreams(config *DeclarativeConfig) int {
	seen := make(map[string]bool)
	count := 0

	// Count from separate upstreams list
	for _, upstream := range config.Upstreams {
		if !seen[upstream.ID] {
			count++
			seen[upstream.ID] = true
		}
	}

	// Count from inline upstreams in routes
	for _, route := range config.Routes {
		if route.Upstream != nil && !seen[route.Upstream.ID] {
			count++
			seen[route.Upstream.ID] = true
		}
	}

	return count
}

// applyCommand applies declarative configuration
func newApplyCommand() *cobra.Command {
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "apply -f FILE",
		Short: "Apply declarative configuration to APISIX",
		Long: `Apply declarative configuration from a YAML/JSON file to APISIX.
The command compares the current state with the desired state and makes necessary changes.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			if file == "" {
				return fmt.Errorf("file is required. Use -f flag to specify configuration file")
			}

			// Load declarative config
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var config DeclarativeConfig
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("invalid YAML: %w", err)
			}

			// Load ADC config (skip validation for dry-run)
			adcConfig, err := loadConfig(dryRun)
			if err != nil && !dryRun {
				return err
			}

			if dryRun {
				fmt.Println("Dry run mode. No changes will be applied.")
				fmt.Printf("Configuration file: %s\n", file)
				fmt.Printf("  Version: %s\n", config.Version)
				fmt.Printf("  Routes: %d\n", len(config.Routes))

				// If config exists, show it
				if adcConfig != nil {
					fmt.Printf("  Would apply to: %s\n", adcConfig.APISIX.BaseURL)
				} else {
					fmt.Println("  Note: No valid APISIX configuration found")
					fmt.Println("        Run 'adc init' to create config file")
				}

				for _, route := range config.Routes {
					fmt.Printf("    - %s: %s\n", route.ID, route.Name)
				}
				return nil
			}

			// Real apply requires valid config
			if adcConfig == nil {
				return fmt.Errorf("APISIX configuration is required for apply command")
			}

			fmt.Printf("Applying configuration to: %s\n", adcConfig.APISIX.BaseURL)
			fmt.Println("Note: Full APISIX integration not implemented yet")

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Declarative configuration file (YAML/JSON)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&force, "force", false, "Force apply without confirmation")
	cmd.MarkFlagRequired("file")

	return cmd
}

// diffCommand shows differences between local and remote config
func newDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff -f FILE",
		Short: "Show differences between local config and APISIX state",
		Long:  `Compare declarative configuration with current APISIX state and show differences.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			if file == "" {
				return fmt.Errorf("file is required. Use -f flag to specify configuration file")
			}

			// Load ADC config
			adcConfig, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Diff functionality not implemented yet\n")
			fmt.Printf("Would compare with APISIX at: %s\n", adcConfig.APISIX.BaseURL)
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Declarative configuration file (YAML/JSON)")
	cmd.MarkFlagRequired("file")

	return cmd
}

// syncCommand synchronizes configuration
func newSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync -f FILE",
		Short: "Synchronize configuration with APISIX",
		Long:  `Synchronize declarative configuration with APISIX, ensuring they match exactly.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			if file == "" {
				return fmt.Errorf("file is required. Use -f flag to specify configuration file")
			}

			// Load ADC config
			adcConfig, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Sync functionality not implemented yet\n")
			fmt.Printf("Would sync with APISIX at: %s\n", adcConfig.APISIX.BaseURL)
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Declarative configuration file (YAML/JSON)")
	cmd.MarkFlagRequired("file")

	return cmd
}
