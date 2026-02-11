package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/api7/adc-go/internal/declarative"
	"gopkg.in/yaml.v3"
)

// BackupManager manages configuration backups
type BackupManager struct {
	BackupDir string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(backupDir string) (*BackupManager, error) {
	if backupDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		backupDir = filepath.Join(home, ".config", "adc", "backups")
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &BackupManager{BackupDir: backupDir}, nil
}

// Backup creates a backup of the current configuration
func (bm *BackupManager) Backup(config *declarative.DeclarativeConfig, description string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("backup-%s.yaml", timestamp)
	backupPath := filepath.Join(bm.BackupDir, filename)

	// Create backup metadata
	metadata := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"description": description,
		"version":     config.Version,
	}

	// Create backup structure
	backupData := map[string]interface{}{
		"metadata":      metadata,
		"configuration": config,
	}

	data, err := yaml.Marshal(backupData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	// Also save as JSON for easier parsing
	jsonPath := filepath.Join(bm.BackupDir, fmt.Sprintf("backup-%s.json", timestamp))
	jsonData, err := json.MarshalIndent(backupData, "", "  ")
	if err == nil {
		os.WriteFile(jsonPath, jsonData, 0644)
	}

	return backupPath, nil
}

// Restore restores configuration from a backup
func (bm *BackupManager) Restore(backupPath string) (*declarative.DeclarativeConfig, error) {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	var backupData struct {
		Metadata      map[string]interface{}        `yaml:"metadata"`
		Configuration declarative.DeclarativeConfig `yaml:"configuration"`
	}

	if err := yaml.Unmarshal(data, &backupData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal backup: %w", err)
	}

	return &backupData.Configuration, nil
}

// List lists all available backups
func (bm *BackupManager) List() ([]BackupInfo, error) {
	files, err := os.ReadDir(bm.BackupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		backupPath := filepath.Join(bm.BackupDir, file.Name())
		info, err := bm.GetBackupInfo(backupPath)
		if err != nil {
			continue
		}

		backups = append(backups, *info)
	}

	return backups, nil
}

// GetBackupInfo gets information about a backup
func (bm *BackupManager) GetBackupInfo(backupPath string) (*BackupInfo, error) {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	var backupData struct {
		Metadata map[string]interface{} `yaml:"metadata"`
	}

	if err := yaml.Unmarshal(data, &backupData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal backup: %w", err)
	}

	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, err
	}

	info := &BackupInfo{
		Path:     backupPath,
		Filename: filepath.Base(backupPath),
		Size:     fileInfo.Size(),
	}

	if timestamp, ok := backupData.Metadata["timestamp"].(string); ok {
		info.Timestamp = timestamp
	}

	if description, ok := backupData.Metadata["description"].(string); ok {
		info.Description = description
	}

	if version, ok := backupData.Metadata["version"].(string); ok {
		info.Version = version
	}

	return info, nil
}

// Delete deletes a backup
func (bm *BackupManager) Delete(backupPath string) error {
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	// Also delete JSON version if exists
	jsonPath := backupPath[:len(backupPath)-5] + ".json"
	os.Remove(jsonPath) // Ignore error

	return nil
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	Path        string
	Filename    string
	Timestamp   string
	Description string
	Version     string
	Size        int64
}
