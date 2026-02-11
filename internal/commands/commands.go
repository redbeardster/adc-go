package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/api7/adc-go/internal/apisix"
	"github.com/api7/adc-go/internal/backup"
	"github.com/api7/adc-go/internal/config"
	"github.com/api7/adc-go/internal/declarative"
	"github.com/api7/adc-go/internal/diff"
	"github.com/api7/adc-go/internal/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Global variables
var (
	cfgFile string
	debug   bool
)

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
	rootCmd.AddCommand(newPingCommand())
	rootCmd.AddCommand(newDumpCommand())
	rootCmd.AddCommand(newBackupCommand())
	rootCmd.AddCommand(newRestoreCommand())
	rootCmd.AddCommand(newBackupListCommand())
	rootCmd.AddCommand(newBackupDeleteCommand())

	return rootCmd.Execute()
}

// loadConfig loads configuration from file or environment
func loadConfig(skipValidation bool) (*config.ADCConfig, error) {
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
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".config", "adc"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg config.ADCConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if debug {
		cfg.Debug = true
	}

	if !skipValidation {
		if cfg.APISIX.AdminKey == "" {
			adminKey := os.Getenv("ADC_APISIX_ADMIN_KEY")
			if adminKey == "" {
				return nil, fmt.Errorf("admin key is required. Set it in config file or ADC_APISIX_ADMIN_KEY environment variable")
			}
			cfg.APISIX.AdminKey = adminKey
		}
	}

	return &cfg, nil
}

// newAPISIXClient creates a new APISIX client
func newAPISIXClient(cfg *config.ADCConfig) *apisix.Client {
	return apisix.NewClient(cfg)
}

// versionCommand prints version information
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ADC Version: dev")
			fmt.Println("Build Time: 2026-02-11")
			fmt.Println("Git Commit: unknown")

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
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "adc")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("config file already exists at %s. Use --force to overwrite", configPath)
	}

	exampleConfig := config.ADCConfig{
		Version: "1.0.0",
		Debug:   false,
		APISIX: config.AdminAPI{
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

	var cfg declarative.DeclarativeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	if cfg.Version == "" {
		return fmt.Errorf("version is required in declarative config")
	}

	validationErrors := []string{}

	for i, route := range cfg.Routes {
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

	for i, service := range cfg.Services {
		if service.ID == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("service at index %d: id is required", i))
		}
	}

	for i, consumer := range cfg.Consumers {
		if consumer.Username == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("consumer at index %d: username is required", i))
		}
	}

	for i, ssl := range cfg.SSLs {
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

	upstreamCount := countUpstreams(&cfg)

	fmt.Printf("✓ Configuration file %s is valid\n", file)
	fmt.Printf("  Version: %s\n", cfg.Version)
	fmt.Printf("  Routes: %d\n", len(cfg.Routes))
	fmt.Printf("  Services: %d\n", len(cfg.Services))
	fmt.Printf("  Upstreams: %d\n", upstreamCount)
	fmt.Printf("  Consumers: %d\n", len(cfg.Consumers))
	fmt.Printf("  SSLs: %d\n", len(cfg.SSLs))
	fmt.Printf("  Global Rules: %d\n", len(cfg.GlobalRules))
	fmt.Printf("  Plugin Configs: %d\n", len(cfg.PluginConfigs))
	fmt.Printf("  Stream Routes: %d\n", len(cfg.StreamRoutes))

	if len(cfg.Routes) > 0 {
		fmt.Println("\n  Routes details:")
		for _, route := range cfg.Routes {
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

func countUpstreams(cfg *declarative.DeclarativeConfig) int {
	seen := make(map[string]bool)
	count := 0

	for _, upstream := range cfg.Upstreams {
		if !seen[upstream.ID] {
			count++
			seen[upstream.ID] = true
		}
	}

	for _, route := range cfg.Routes {
		if route.Upstream != nil && !seen[route.Upstream.ID] {
			count++
			seen[route.Upstream.ID] = true
		}
	}

	return count
}

// pingCommand checks connection to APISIX
func newPingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Check connection to APISIX Admin API",
		Long:  `Verify that ADC can connect to APISIX Admin API and authenticate successfully.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Connecting to APISIX at: %s\n", cfg.APISIX.BaseURL)

			client := newAPISIXClient(cfg)

			if err := client.Ping(); err != nil {
				fmt.Printf("✗ Connection failed: %v\n", err)
				return err
			}

			fmt.Println("✓ Successfully connected to APISIX Admin API")
			return nil
		},
	}
}

// dumpCommand exports current APISIX configuration
func newDumpCommand() *cobra.Command {
	var outputFile string
	var format string

	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Export current APISIX configuration",
		Long:  `Dump current APISIX configuration to a declarative YAML/JSON file.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Dumping configuration from: %s\n", cfg.APISIX.BaseURL)

			client := newAPISIXClient(cfg)
			syncer := sync.NewSyncer(client)

			remoteConfig, err := syncer.GetRemoteState()
			if err != nil {
				return fmt.Errorf("failed to get remote state: %w", err)
			}

			var data []byte
			if format == "json" {
				data, err = json.MarshalIndent(remoteConfig, "", "  ")
			} else {
				data, err = yaml.Marshal(remoteConfig)
			}
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, data, 0644); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}
				fmt.Printf("✓ Configuration dumped to: %s\n", outputFile)
			} else {
				fmt.Println(string(data))
			}

			fmt.Printf("\nSummary:\n")
			fmt.Printf("  Routes: %d\n", len(remoteConfig.Routes))
			fmt.Printf("  Services: %d\n", len(remoteConfig.Services))
			fmt.Printf("  Upstreams: %d\n", len(remoteConfig.Upstreams))
			fmt.Printf("  Consumers: %d\n", len(remoteConfig.Consumers))
			fmt.Printf("  SSLs: %d\n", len(remoteConfig.SSLs))
			fmt.Printf("  Global Rules: %d\n", len(remoteConfig.GlobalRules))
			fmt.Printf("  Plugin Configs: %d\n", len(remoteConfig.PluginConfigs))
			fmt.Printf("  Stream Routes: %d\n", len(remoteConfig.StreamRoutes))

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&format, "format", "yaml", "Output format: yaml or json")

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

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var localConfig declarative.DeclarativeConfig
			if err := yaml.Unmarshal(data, &localConfig); err != nil {
				return fmt.Errorf("invalid YAML: %w", err)
			}

			cfg, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Comparing with APISIX at: %s\n\n", cfg.APISIX.BaseURL)

			client := newAPISIXClient(cfg)
			syncer := sync.NewSyncer(client)

			fmt.Println("Fetching current APISIX state...")
			remoteConfig, err := syncer.GetRemoteState()
			if err != nil {
				return fmt.Errorf("failed to get remote state: %w", err)
			}

			diffResult := syncer.CalculateDiff(&localConfig, remoteConfig)

			if !diffResult.HasChanges() {
				fmt.Println("✓ No differences found. Configurations are in sync.")
				return nil
			}

			diff.PrintDiff(diffResult)
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Declarative configuration file (YAML/JSON)")
	cmd.MarkFlagRequired("file")
	return cmd
}

// applyCommand applies declarative configuration
func newApplyCommand() *cobra.Command {
	var dryRun bool
	var force bool
	var noBackup bool

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

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var localConfig declarative.DeclarativeConfig
			if err := yaml.Unmarshal(data, &localConfig); err != nil {
				return fmt.Errorf("invalid YAML: %w", err)
			}

			cfg, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Applying configuration to: %s\n", cfg.APISIX.BaseURL)

			client := newAPISIXClient(cfg)
			syncer := sync.NewSyncer(client)

			fmt.Println("Fetching current APISIX state...")
			remoteConfig, err := syncer.GetRemoteState()
			if err != nil {
				return fmt.Errorf("failed to get remote state: %w", err)
			}

			fmt.Println("Calculating differences...")
			diffResult := syncer.CalculateDiff(&localConfig, remoteConfig)

			if !diffResult.HasChanges() {
				fmt.Println("✓ No changes needed. Configuration is up to date.")
				return nil
			}

			diff.PrintDiff(diffResult)

			if dryRun {
				fmt.Println("\nDry run mode. No changes were applied.")
				return nil
			}

			// Create backup before applying changes
			if !noBackup {
				fmt.Println("\nCreating backup before applying changes...")
				bm, err := backup.NewBackupManager("")
				if err == nil {
					backupPath, err := bm.Backup(remoteConfig, fmt.Sprintf("Auto backup before apply from %s", file))
					if err != nil {
						fmt.Printf("⚠️  Warning: Failed to create backup: %v\n", err)
					} else {
						fmt.Printf("✓ Backup created: %s\n", backupPath)
					}
				}
			}

			if !force {
				fmt.Print("\nDo you want to apply these changes? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" && response != "y" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			fmt.Println("\nApplying changes...")
			if err := syncer.ApplyDiff(diffResult, false); err != nil {
				return fmt.Errorf("failed to apply changes: %w", err)
			}

			fmt.Println("\n✓ Configuration applied successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Declarative configuration file (YAML/JSON)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&force, "force", false, "Force apply without confirmation")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "Skip automatic backup before applying")
	cmd.MarkFlagRequired("file")

	return cmd
}

// syncCommand synchronizes configuration
func newSyncCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "sync -f FILE",
		Short: "Synchronize configuration with APISIX",
		Long: `Synchronize declarative configuration with APISIX, ensuring they match exactly.
This command will create, update, and DELETE resources to match the local configuration.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			if file == "" {
				return fmt.Errorf("file is required. Use -f flag to specify configuration file")
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var localConfig declarative.DeclarativeConfig
			if err := yaml.Unmarshal(data, &localConfig); err != nil {
				return fmt.Errorf("invalid YAML: %w", err)
			}

			cfg, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Synchronizing with APISIX at: %s\n", cfg.APISIX.BaseURL)
			fmt.Println("⚠️  WARNING: This will DELETE resources not present in the local configuration!")

			client := newAPISIXClient(cfg)
			syncer := sync.NewSyncer(client)

			fmt.Println("\nFetching current APISIX state...")
			remoteConfig, err := syncer.GetRemoteState()
			if err != nil {
				return fmt.Errorf("failed to get remote state: %w", err)
			}

			fmt.Println("Calculating differences...")
			diffResult := syncer.CalculateDiff(&localConfig, remoteConfig)

			if !diffResult.HasChanges() {
				fmt.Println("✓ No changes needed. Configuration is already in sync.")
				return nil
			}

			diff.PrintDiff(diffResult)

			if !force {
				fmt.Print("\n⚠️  Do you want to synchronize (including deletions)? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" && response != "y" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			fmt.Println("\nSynchronizing...")
			if err := syncer.ApplyDiff(diffResult, true); err != nil {
				return fmt.Errorf("failed to synchronize: %w", err)
			}

			fmt.Println("\n✓ Configuration synchronized successfully!")
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Declarative configuration file (YAML/JSON)")
	cmd.Flags().BoolVar(&force, "force", false, "Force sync without confirmation")
	cmd.MarkFlagRequired("file")

	return cmd
}
