package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/api7/adc-go/internal/backup"
	"github.com/api7/adc-go/internal/sync"
	"github.com/spf13/cobra"
)

// newBackupCommand creates backup command
func newBackupCommand() *cobra.Command {
	var description string
	var backupDir string

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of current APISIX configuration",
		Long:  `Create a backup of the current APISIX configuration for later restoration.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load ADC config
			adcConfig, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Creating backup from: %s\n", adcConfig.APISIX.BaseURL)

			// Create client and syncer
			client := newAPISIXClient(adcConfig)
			syncer := sync.NewSyncer(client)

			// Get remote state
			fmt.Println("Fetching current APISIX state...")
			remoteConfig, err := syncer.GetRemoteState()
			if err != nil {
				return fmt.Errorf("failed to get remote state: %w", err)
			}

			// Create backup manager
			bm, err := backup.NewBackupManager(backupDir)
			if err != nil {
				return fmt.Errorf("failed to create backup manager: %w", err)
			}

			// Create backup
			backupPath, err := bm.Backup(remoteConfig, description)
			if err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}

			fmt.Printf("✓ Backup created successfully: %s\n", backupPath)

			// Show backup info
			info, err := bm.GetBackupInfo(backupPath)
			if err == nil {
				fmt.Printf("\nBackup details:\n")
				fmt.Printf("  Timestamp: %s\n", info.Timestamp)
				fmt.Printf("  Description: %s\n", info.Description)
				fmt.Printf("  Size: %d bytes\n", info.Size)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Backup description")
	cmd.Flags().StringVar(&backupDir, "backup-dir", "", "Backup directory (default: ~/.config/adc/backups)")

	return cmd
}

// newRestoreCommand creates restore command
func newRestoreCommand() *cobra.Command {
	var force bool
	var backupDir string

	cmd := &cobra.Command{
		Use:   "restore [backup-file]",
		Short: "Restore APISIX configuration from a backup",
		Long:  `Restore APISIX configuration from a previously created backup.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var backupPath string

			// Create backup manager
			bm, err := backup.NewBackupManager(backupDir)
			if err != nil {
				return fmt.Errorf("failed to create backup manager: %w", err)
			}

			// If no backup file specified, list available backups
			if len(args) == 0 {
				backups, err := bm.List()
				if err != nil {
					return fmt.Errorf("failed to list backups: %w", err)
				}

				if len(backups) == 0 {
					fmt.Println("No backups found")
					return nil
				}

				fmt.Println("Available backups:")
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "FILENAME\tTIMESTAMP\tDESCRIPTION\tSIZE")
				for _, b := range backups {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d bytes\n", b.Filename, b.Timestamp, b.Description, b.Size)
				}
				w.Flush()

				fmt.Println("\nUse 'adc restore <backup-file>' to restore a specific backup")
				return nil
			}

			backupPath = args[0]

			// Check if backup file exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				// Try to find in backup directory
				backupPath = filepath.Join(bm.BackupDir, backupPath)
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					return fmt.Errorf("backup file not found: %s", args[0])
				}
			}

			// Load backup
			fmt.Printf("Loading backup from: %s\n", backupPath)
			config, err := bm.Restore(backupPath)
			if err != nil {
				return fmt.Errorf("failed to restore backup: %w", err)
			}

			// Load ADC config
			adcConfig, err := loadConfig(false)
			if err != nil {
				return err
			}

			fmt.Printf("Restoring to: %s\n", adcConfig.APISIX.BaseURL)

			// Create client and syncer
			client := newAPISIXClient(adcConfig)
			syncer := sync.NewSyncer(client)

			// Get current remote state
			fmt.Println("Fetching current APISIX state...")
			remoteConfig, err := syncer.GetRemoteState()
			if err != nil {
				return fmt.Errorf("failed to get remote state: %w", err)
			}

			// Calculate diff
			fmt.Println("Calculating differences...")
			diffResult := syncer.CalculateDiff(config, remoteConfig)

			if !diffResult.HasChanges() {
				fmt.Println("✓ No changes needed. Configuration is already up to date.")
				return nil
			}

			// Print diff
			fmt.Println("\nChanges to be applied:")
			// TODO: Print diff

			// Ask for confirmation if not forced
			if !force {
				fmt.Print("\n⚠️  Do you want to restore this backup? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" && response != "y" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			// Apply changes
			fmt.Println("\nRestoring configuration...")
			if err := syncer.ApplyDiff(diffResult, true); err != nil {
				return fmt.Errorf("failed to restore configuration: %w", err)
			}

			fmt.Println("\n✓ Configuration restored successfully!")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force restore without confirmation")
	cmd.Flags().StringVar(&backupDir, "backup-dir", "", "Backup directory (default: ~/.config/adc/backups)")

	return cmd
}

// newBackupListCommand creates backup list command
func newBackupListCommand() *cobra.Command {
	var backupDir string

	cmd := &cobra.Command{
		Use:   "backup-list",
		Short: "List all available backups",
		Long:  `List all available configuration backups.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bm, err := backup.NewBackupManager(backupDir)
			if err != nil {
				return fmt.Errorf("failed to create backup manager: %w", err)
			}

			backups, err := bm.List()
			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			if len(backups) == 0 {
				fmt.Println("No backups found")
				return nil
			}

			fmt.Printf("Found %d backup(s):\n\n", len(backups))
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "FILENAME\tTIMESTAMP\tDESCRIPTION\tVERSION\tSIZE")
			for _, b := range backups {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d bytes\n",
					b.Filename, b.Timestamp, b.Description, b.Version, b.Size)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVar(&backupDir, "backup-dir", "", "Backup directory (default: ~/.config/adc/backups)")

	return cmd
}

// newBackupDeleteCommand creates backup delete command
func newBackupDeleteCommand() *cobra.Command {
	var backupDir string
	var force bool

	cmd := &cobra.Command{
		Use:   "backup-delete <backup-file>",
		Short: "Delete a backup",
		Long:  `Delete a specific backup file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupPath := args[0]

			bm, err := backup.NewBackupManager(backupDir)
			if err != nil {
				return fmt.Errorf("failed to create backup manager: %w", err)
			}

			// Check if backup file exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				// Try to find in backup directory
				backupPath = filepath.Join(bm.BackupDir, backupPath)
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					return fmt.Errorf("backup file not found: %s", args[0])
				}
			}

			// Ask for confirmation if not forced
			if !force {
				fmt.Printf("⚠️  Are you sure you want to delete backup: %s? (yes/no): ", filepath.Base(backupPath))
				var response string
				fmt.Scanln(&response)
				if response != "yes" && response != "y" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := bm.Delete(backupPath); err != nil {
				return fmt.Errorf("failed to delete backup: %w", err)
			}

			fmt.Printf("✓ Backup deleted: %s\n", filepath.Base(backupPath))
			return nil
		},
	}

	cmd.Flags().StringVar(&backupDir, "backup-dir", "", "Backup directory (default: ~/.config/adc/backups)")
	cmd.Flags().BoolVar(&force, "force", false, "Force delete without confirmation")

	return cmd
}
