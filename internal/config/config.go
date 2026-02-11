package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ADCConfig struct {
	APISIX  AdminAPI `yaml:"apisix" json:"apisix"`
	Debug   bool     `yaml:"debug" json:"debug"`
	Version string   `yaml:"version" json:"version"`
}

type AdminAPI struct {
	BaseURL      string `yaml:"base_url" json:"base_url"`
	AdminKey     string `yaml:"admin_key" json:"admin_key"`
	AdminKeyName string `yaml:"admin_key_name,omitempty" json:"admin_key_name,omitempty"`
}

// LoadConfig загружает конфигурацию из файла или переменных окружения
func LoadConfig(configPath string) (*ADCConfig, error) {
	viper.SetEnvPrefix("ADC")
	viper.AutomaticEnv()

	// Устанавливаем значения по умолчанию
	viper.SetDefault("apisix.base_url", "http://127.0.0.1:9180")
	viper.SetDefault("apisix.admin_key_name", "X-API-Key")
	viper.SetDefault("debug", false)
	viper.SetDefault("version", "1.0.0")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Ищем конфиг в стандартных местах
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
		// Конфиг не найден, используем значения по умолчанию
	}

	var config ADCConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(config *ADCConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir := filepath.Join(home, ".config", "adc")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		configPath = filepath.Join(configDir, "config.yaml")
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
