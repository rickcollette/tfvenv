package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"tfvenv/snaps"
)

const (
	terraformDownloadURL  = "https://releases.hashicorp.com/terraform/"
	terragruntDownloadURL = "https://github.com/gruntwork-io/terragrunt/releases/download/"
	tfvenvrcFileName      = ".tfvenvrc"
	lockFileName          = ".lock"
)

// Global logger
var logger = logrus.New()

// Config structure parsed from tfvenvrc
type Config struct {
	TfVersion          string            `mapstructure:"TF_VERSION"`
	TgVersion          string            `mapstructure:"TG_VERSION"`
	S3StateBucket      string            `mapstructure:"S3_STATE_BUCKET"`
	S3StatePath        string            `mapstructure:"S3_STATE_PATH"`
	S3ModulesBucket    string            `mapstructure:"S3_MODULES_BUCKET"`
	S3ModulesPath      string            `mapstructure:"S3_MODULES_PATH"`
	Region             string            `mapstructure:"REGION"`
	AccessKey          string            `mapstructure:"ACCESS_KEY"`
	SecretKey          string            `mapstructure:"SECRET_KEY"`
	RemoteSnapEndpoint string            `mapstructure:"REMOTE_SNAP_ENDPOINT"`
	RemoteSnapAuth     string            `mapstructure:"REMOTE_SNAP_AUTH"`
	RemoteSnapType     string            `mapstructure:"REMOTE_SNAP_TYPE"`
	EnvVars            map[string]string `mapstructure:"ENV_VARS"`
}

// EnvironmentState holds the structure of the environment's state.
type EnvironmentState struct {
	TerraformVersion   string            `json:"terraform_version"`
	TerragruntVersion  string            `json:"terragrunt_version"`
	OS                 string            `json:"os"`
	Architecture       string            `json:"architecture"`
	EnvironmentVars    map[string]string `json:"environment_vars"`
	Plugins            map[string]string `json:"plugins"`
	AdditionalMetadata map[string]string `json:"additional_metadata"`
}

// TerraformRelease represents a specific version release of Terraform.
// It contains the version string in semantic versioning format.
type TerraformRelease struct {
	Version string `json:"version"`
}

// TerraformReleaseIndex holds the index of Terraform releases.
// It maps version strings to their corresponding builds.
type TerraformReleaseIndex struct {
	Versions map[string]struct {
		Builds []TerraformRelease `json:"builds"`
	} `json:"versions"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	// Additional fields are ignored
}

func main() {
	initializeLogger()

	var rootCmd = &cobra.Command{
		Use:   "tfvenv",
		Short: "tfvenv manages virtual environments for Terraform and Terragrunt",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Validate env-dir exists
			envDir := viper.GetString("env-dir")
			if _, err := os.Stat(envDir); os.IsNotExist(err) {
				logger.Fatalf("Environment directory %s does not exist", envDir)
			}
		},
	}

	// Define persistent flags
	rootCmd.PersistentFlags().StringP("env-dir", "e", ".", "Base directory for environments")
	viper.BindPFlag("env-dir", rootCmd.PersistentFlags().Lookup("env-dir"))

	// Add all subcommands to rootCmd
	rootCmd.AddCommand(createCmd())
	rootCmd.AddCommand(deleteCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(activateCmd())
	rootCmd.AddCommand(deactivateCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(upgradeCmd())
	rootCmd.AddCommand(mergeCmd())
	rootCmd.AddCommand(lockCmd())
	rootCmd.AddCommand(unlockCmd())
	rootCmd.AddCommand(hclfmtCmd())
	rootCmd.AddCommand(completionCmd(rootCmd))
	rootCmd.AddCommand(listVersionsCmd())
	rootCmd.AddCommand(switchCmd())
	rootCmd.AddCommand(terraformInstallCmd())
	rootCmd.AddCommand(terragruntInstallCmd())
	rootCmd.AddCommand(cleanupCmd())
	rootCmd.AddCommand(snapCmd()) // Only add once

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("Error executing command: %v", err)
	}
}
// fetchEnvironmentState gathers and returns the current state of the environment.
func fetchEnvironmentState() []byte {
	// Initialize the environment state structure
	envState := EnvironmentState{
		OS:                 runtime.GOOS,
		Architecture:       runtime.GOARCH,
		EnvironmentVars:    getRelevantEnvVars(),
		Plugins:            getPlugins(),
		AdditionalMetadata: getAdditionalMetadata(),
	}

	// Fetch Terraform version
	tfVersion, err := getTerraformVersion()
	if err != nil {
		fmt.Printf("Error fetching Terraform version: %v\n", err)
		envState.TerraformVersion = "unknown"
	} else {
		envState.TerraformVersion = tfVersion
	}

	// Fetch Terragrunt version
	tgVersion, err := getTerragruntVersion()
	if err != nil {
		fmt.Printf("Error fetching Terragrunt version: %v\n", err)
		envState.TerragruntVersion = "unknown"
	} else {
		envState.TerragruntVersion = tgVersion
	}

	// Convert the state to JSON format
	stateJSON, _ := json.Marshal(envState)
	return stateJSON
}

// getRelevantEnvVars fetches environment variables that are relevant to the environment state.
func getRelevantEnvVars() map[string]string {
	relevantVars := []string{"TF_VAR_region", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "GOOS", "GOARCH"}
	envVars := make(map[string]string)

	for _, envVar := range relevantVars {
		if value, exists := os.LookupEnv(envVar); exists {
			envVars[envVar] = value
		}
	}
	return envVars
}

// getAdditionalMetadata returns additional metadata related to the environment.
func getAdditionalMetadata() map[string]string {
	return map[string]string{
		"hostname": getHostname(),
		"user":     getUsername(),
	}
}

// getHostname returns the system's hostname.
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getUsername returns the current user's username.
func getUsername() string {
	user := os.Getenv("USER")
	if user == "" {
		return "unknown"
	}
	return user
}
// snapCmd defines the "snap" command and its subcommands.
func snapCmd() *cobra.Command {
	snapCmd := &cobra.Command{
		Use:   "snap",
		Short: "Manage environment snaps",
	}

	// Add subcommands to snapCmd correctly
	snapCmd.AddCommand(saveSnapCmd())
	snapCmd.AddCommand(getSnapCmd())
	snapCmd.AddCommand(updateSnapCmd())
	snapCmd.AddCommand(removeSnapCmd())
	snapCmd.AddCommand(snapRemoteConfigCmd())
	snapCmd.AddCommand(snapRemoteGetCmd())
	snapCmd.AddCommand(snapRemoteSaveCmd())
	snapCmd.AddCommand(snapRemoteListCmd())
	snapCmd.AddCommand(snapRemoteRemoveCmd())

	return snapCmd
}

// getTerraformVersion dynamically fetches the Terraform version installed.
func getTerraformVersion() (string, error) {
	out, err := exec.Command("terraform", "-version").Output()
	if err != nil {
		return "", err
	}
	return parseVersion(string(out), "Terraform")
}

// getTerragruntVersion dynamically fetches the Terragrunt version installed.
func getTerragruntVersion() (string, error) {
	out, err := exec.Command("terragrunt", "-version").Output()
	if err != nil {
		return "", err
	}
	return parseVersion(string(out), "Terragrunt")
}

// parseVersion extracts the version from the output string.
func parseVersion(output, toolName string) (string, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, toolName) {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				return parts[1], nil
			}
		}
	}
	return "", fmt.Errorf("failed to parse version for %s", toolName)
}

// getEnvVars retrieves the relevant environment variables.
func getEnvVars() map[string]string {
	return map[string]string{
		"TF_VAR_region": os.Getenv("TF_VAR_region"),
	}
}

// getPlugins retrieves the list of plugins and their versions.
func getPlugins() map[string]string {
	return map[string]string{
		"aws":   "3.46.0",
		"azure": "2.29.0",
	}
}
func saveSnapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save <env-name> <filename>",
		Short: "Save the specified environment to a snap file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			filename := args[1]

			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)

			// Fetch the current environment state
			snapData := fetchEnvironmentState()

			// Convert snapData to the required snap object
			var snap snaps.Snap
			err := json.Unmarshal(snapData, &snap)
			if err != nil {
				logger.Errorf("error parsing snap data: %v", err)
				fmt.Printf("Error parsing snap data: %v\n", err)
				os.Exit(1)
			}

			// Pass envPath and filename to GetSnapFilePath
			filePath := snaps.GetSnapFilePath(envPath, filename)

			// Save the snap using the SaveSnap function
			err = snaps.SaveSnap(filePath, &snap)
			if err != nil {
				logger.Errorf("error saving snap: %v", err)
				fmt.Printf("Error saving snap: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Snap saved successfully to %s\n", filePath)
			logger.Infof("Snap saved successfully to %s", filePath)
		},
	}
}
func getSnapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <env-name> <snap-name>",
		Short: "Get a snap from the specified environment's local storage",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			snapName := args[1]

			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)
			filePath := snaps.GetSnapFilePath(envPath, snapName)

			// Check if the snap file exists
			if !fileExists(filePath) {
				fmt.Printf("Snap '%s' does not exist in environment '%s'.\n", snapName, envName)
				logger.Warnf("snap '%s' does not exist in environment '%s'", snapName, envName)
				return
			}

			// Get the snap using the GetSnap function
			snap, err := snaps.GetSnap(filePath)
			if err != nil {
				fmt.Printf("Error retrieving snap: %v\n", err)
				logger.Errorf("error retrieving snap: %v", err)
				return
			}

			// Use the retrieved snap (e.g., print it)
			fmt.Printf("Snap '%s' retrieved successfully: %+v\n", snapName, snap)
			logger.Infof("Snap '%s' retrieved successfully.", snapName)
		},
	}
}

func updateSnapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update <env-name> <filename>",
		Short: "Update an existing snap file for the specified environment",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			snapName := args[1]

			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)

			filePath := snaps.GetSnapFilePath(envPath, snapName)

			terraformVersion, err := getTerraformVersion()
			if err != nil {
				fmt.Printf("Error getting Terraform version: %v\n", err)
				logger.Errorf("error getting Terraform version: %v", err)
				return
			}
			terragruntVersion, err := getTerragruntVersion()
			if err != nil {
				fmt.Printf("Error getting Terragrunt version: %v\n", err)
				logger.Errorf("error getting Terragrunt version: %v", err)
				return
			}
			plugins := getPlugins()
			envVars := getEnvVars()

			updatedSnap := snaps.Snap{
				TerraformVersion:  terraformVersion,
				TerragruntVersion: terragruntVersion,
				Plugins:           plugins,
				EnvVars:           envVars,
			}

			err = snaps.UpdateSnap(filePath, &updatedSnap)
			if err != nil {
				fmt.Printf("Error updating snap: %v\n", err)
				logger.Errorf("error updating snap: %v", err)
			} else {
				fmt.Printf("Snap updated successfully in %s\n", filePath)
				logger.Infof("Snap updated successfully in %s", filePath)
			}
		},
	}
}
func removeSnapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <env-name> <snap-name>",
		Short: "Remove a snap file from the specified environment's local storage",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			snapName := args[1]

			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)
			filePath := snaps.GetSnapFilePath(envPath, snapName)

			err := snaps.RemoveSnap(filePath)
			if err != nil {
				fmt.Printf("Error removing snap: %v\n", err)
				logger.Errorf("error removing snap: %v", err)
			} else {
				fmt.Printf("Snap '%s' removed successfully from '%s'.\n", snapName, filePath)
				logger.Infof("Snap '%s' removed successfully from '%s'.", snapName, filePath)
			}
		},
	}
}
func snapRemoteConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Configure remote snap settings",
		Run: func(cmd *cobra.Command, args []string) {
			endpoint := os.Getenv("REMOTE_SNAP_ENDPOINT")
			auth := os.Getenv("REMOTE_SNAP_AUTH")
			snapType := os.Getenv("REMOTE_SNAP_TYPE")

			if snapType != "S3" {
				fmt.Println("Only S3 is supported in this iteration.")
				logger.Warnf("unsupported snap type: %s", snapType)
				return
			}

			if endpoint == "" || auth == "" {
				fmt.Println("REMOTE_SNAP_ENDPOINT and REMOTE_SNAP_AUTH must be set.")
				logger.Warn("REMOTE_SNAP_ENDPOINT or REMOTE_SNAP_AUTH not set")
				return
			}

			fmt.Printf("Remote Snap Configured: %s (Type: %s)\n", endpoint, snapType)
			logger.Infof("Remote Snap Configured: %s (Type: %s)", endpoint, snapType)
		},
	}
}
func snapRemoteGetCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "get <env-name> <snap-name>",
        Short: "Get a snap from the specified environment's remote S3 storage",
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
            envName := args[0]
            snapName := args[1]

            envDir := viper.GetString("env-dir")
            envPath := filepath.Join(envDir, envName)
            filePath := snaps.GetSnapFilePath(envPath, snapName)

            auth := os.Getenv("REMOTE_SNAP_AUTH")
            if auth == "" {
                logger.Error("REMOTE_SNAP_AUTH must be set")
                fmt.Println("Error: REMOTE_SNAP_AUTH must be set.")
                os.Exit(1)
            }

            accessKey := os.Getenv("AWS_ACCESS_KEY")
            secretKey := os.Getenv("AWS_SECRET_KEY")
            region := os.Getenv("AWS_REGION")

            if accessKey == "" || secretKey == "" || region == "" {
                logger.Error("AWS_ACCESS_KEY, AWS_SECRET_KEY, and AWS_REGION must be set")
                fmt.Println("Error: AWS_ACCESS_KEY, AWS_SECRET_KEY, and AWS_REGION must be set.")
                os.Exit(1)
            }

            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            sanitizedSnapName, err := snaps.SanitizeSnapName(snapName)
            if err != nil {
                logger.Errorf("invalid snap name '%s': %v", snapName, err)
                fmt.Printf("Error: %v\n", err)
                os.Exit(1)
            }

            encryptedSnap, err := snaps.GetRemoteSnap(ctx, sanitizedSnapName, accessKey, secretKey, region)
            if err != nil {
                logger.Errorf("error retrieving snap '%s': %v", sanitizedSnapName, err)
                fmt.Printf("Error retrieving snap: %v\n", err)
                os.Exit(1)
            }

            snapData, err := snaps.Decrypt(encryptedSnap)
            if err != nil {
                logger.Errorf("error decrypting snap '%s': %v", sanitizedSnapName, err)
                fmt.Printf("Error decrypting snap: %v\n", err)
                os.Exit(1)
            }

            // Use filePath to save the decrypted snap
            err = os.WriteFile(filePath, snapData, 0644)
            if err != nil {
                logger.Errorf("error saving snap file '%s': %v", filePath, err)
                fmt.Printf("Error saving snap file: %v\n", err)
                os.Exit(1)
            }

            fmt.Printf("Snap '%s' retrieved, decrypted, and saved to '%s' successfully.\n", sanitizedSnapName, filePath)
            logger.Infof("Snap '%s' retrieved and saved to '%s'.", sanitizedSnapName, filePath)
        },
    }
}

func snapRemoteSaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save <env-name> <snap-name>",
		Short: "Save a snap to the specified environment's remote S3 storage",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			snapName := args[1]

			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)
			filePath := snaps.GetSnapFilePath(envPath, snapName)

			auth := os.Getenv("REMOTE_SNAP_AUTH")
			if auth == "" {
				fmt.Println("REMOTE_SNAP_AUTH must be set.")
				logger.Warn("REMOTE_SNAP_AUTH not set")
				return
			}

			accessKey := os.Getenv("AWS_ACCESS_KEY")
			secretKey := os.Getenv("AWS_SECRET_KEY")
			region := os.Getenv("AWS_REGION")

			if accessKey == "" || secretKey == "" || region == "" {
				fmt.Println("AWS_ACCESS_KEY, AWS_SECRET_KEY, and AWS_REGION must be set.")
				logger.Warn("AWS_ACCESS_KEY, AWS_SECRET_KEY, or AWS_REGION not set")
				return
			}

			// Read the snap data from the local file
			snapData, err := os.ReadFile(fmt.Sprintf("%s.snap", snapName))
			if err != nil {
				fmt.Printf("Error reading snap: %v\n", err)
				logger.Errorf("error reading snap file: %v", err)
				return
			}

			// Encrypt the snap before uploading
			encryptedSnap, err := snaps.Encrypt(snapData)
			if err != nil {
				fmt.Printf("Error encrypting snap: %v\n", err)
				logger.Errorf("error encrypting snap: %v", err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err = snaps.SaveRemoteSnap(ctx, snapName, []byte(encryptedSnap), accessKey, secretKey, region)
			if err != nil {
				fmt.Printf("Error uploading snap: %v\n", err)
				logger.Errorf("error uploading snap: %v", err)
				return
			}

			fmt.Printf("Snap '%s' encrypted and uploaded successfully to S3.\n", snapName)
			logger.Infof("Snap '%s' encrypted and uploaded successfully to S3 at %s.", snapName, filePath)
		},
	}
}
func snapRemoteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <env-name>",
		Short: "List all snaps from the specified environment's remote S3 storage",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)

			accessKey := os.Getenv("AWS_ACCESS_KEY")
			secretKey := os.Getenv("AWS_SECRET_KEY")
			region := os.Getenv("AWS_REGION")

			if accessKey == "" || secretKey == "" || region == "" {
				fmt.Println("AWS_ACCESS_KEY, AWS_SECRET_KEY, and AWS_REGION must be set.")
				logger.Warn("AWS_ACCESS_KEY, AWS_SECRET_KEY, or AWS_REGION not set")
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			snapsList, err := snaps.ListRemoteSnaps(ctx, accessKey, secretKey, region)
			if err != nil {
				fmt.Printf("Error listing snaps: %v\n", err)
				logger.Errorf("error listing snaps: %v", err)
				return
			}

			if len(snapsList) == 0 {
				fmt.Println("No snaps found in the remote S3 bucket.")
				return
			}

			fmt.Println("Snaps found in the remote S3 bucket:")
			for _, snap := range snapsList {
				fmt.Println(" -", snap)
			}
			logger.Infof("Listed %d snaps from remote S3 at %s.", len(snapsList), envPath)
		},
	}
}
func snapRemoteRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <env-name> <snap-name>",
		Short: "Remove a snap from the specified environment's remote S3 storage",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			snapName := args[1]

			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)

			accessKey := os.Getenv("AWS_ACCESS_KEY")
			secretKey := os.Getenv("AWS_SECRET_KEY")
			region := os.Getenv("AWS_REGION")

			if accessKey == "" || secretKey == "" || region == "" {
				fmt.Println("AWS_ACCESS_KEY, AWS_SECRET_KEY, and AWS_REGION must be set.")
				logger.Warn("AWS_ACCESS_KEY, AWS_SECRET_KEY, or AWS_REGION not set")
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := snaps.RemoveRemoteSnap(ctx, snapName, accessKey, secretKey, region)
			if err != nil {
				fmt.Printf("Error removing snap: %v\n", err)
				logger.Errorf("error removing snap: %v", err)
				return
			}

			fmt.Printf("Snap '%s' removed successfully from remote S3 storage.\n", snapName)
			logger.Infof("Snap '%s' removed successfully from remote S3 storage at %s.", snapName, envPath)
		},
	}
}
func cleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup <env-name>",
		Short: "Clean up duplicate or unused provider plugins from the specified environment's plugin cache",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)

			pluginCacheDir := filepath.Join(os.Getenv("HOME"), ".tfvenv", "plugin-cache")
			if !fileExists(pluginCacheDir) {
				fmt.Printf("Plugin cache directory %s does not exist.\n", pluginCacheDir)
				logger.Warnf("plugin cache directory %s does not exist", pluginCacheDir)
				return
			}

			// Remove duplicates
			err := cleanDuplicateProviders(pluginCacheDir)
			if err != nil {
				logger.Errorf("error cleaning duplicate providers: %v", err)
				fmt.Printf("Error cleaning duplicate providers: %v\n", err)
				os.Exit(1)
			}

			// Remove unused providers
			err = cleanUnusedProviders(pluginCacheDir, envPath)
			if err != nil {
				logger.Errorf("error cleaning unused providers: %v", err)
				fmt.Printf("Error cleaning unused providers: %v\n", err)
			}

			fmt.Println("Cleanup completed successfully.")
			logger.Infof("Cleanup completed successfully for environment '%s' at %s.", envName, envPath)
		},
	}

	return cmd
}

// cleanUnusedProviders scans the plugin cache and removes providers not used by any existing environment
func cleanUnusedProviders(pluginCacheDir string, baseEnvDir string) error {
	logger.Infof("Cleaning unused providers from plugin cache at %s", pluginCacheDir)

	// List all existing environments
	envs, err := getEnvironments(baseEnvDir)
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	if len(envs) == 0 {
		logger.Warn("No environments found. Skipping cleanup.")
		return nil
	}

	// Collect all used providers (source and version)
	usedProviders, err := collectUsedProviders(envs, baseEnvDir)
	if err != nil {
		return fmt.Errorf("failed to collect used providers: %w", err)
	}

	if len(usedProviders) == 0 {
		logger.Warn("No providers found in any environment. Skipping cleanup.")
		return nil
	}

	logger.Infof("Total used providers found: %d", len(usedProviders))

	// Iterate through the plugin cache and remove unused providers
	err = filepath.Walk(pluginCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Determine if the file is a provider plugin
		// Terraform provider plugins typically follow the naming convention:
		// terraform-provider-<NAME>_v<VERSION>_<OS>_<ARCH>[.zip]
		// Examples:
		// terraform-provider-aws_v3.74.0_linux_amd64.zip
		// terraform-provider-google_v4.8.0_darwin_amd64

		filename := info.Name()
		if strings.HasPrefix(filename, "terraform-provider-") {
			// Remove any file extensions like .zip, .exe, etc.
			baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

			// Split to extract provider name and version
			// Example split: ["terraform-provider-aws", "v3.74.0", "linux", "amd64"]
			parts := strings.Split(baseName, "_")
			if len(parts) < 2 {
				logger.Warnf("Unexpected provider plugin filename format: %s", filename)
				return nil
			}

			providerPart := parts[0] // "terraform-provider-aws"
			versionPart := parts[1]  // "v3.74.0"

			providerName := strings.TrimPrefix(providerPart, "terraform-provider-")
			providerVersion := strings.TrimPrefix(versionPart, "v")

			// Check if this provider and version is in the usedProviders map
			key := fmt.Sprintf("%s_%s", providerName, providerVersion)
			if !usedProviders[key] {
				// Remove the unused provider plugin
				logger.Infof("Removing unused provider plugin: %s", path)
				err := os.Remove(path)
				if err != nil {
					logger.Warnf("Failed to remove %s: %v", path, err)
				} else {
					fmt.Printf("Removed unused provider plugin: %s\n", path)
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error cleaning plugin cache: %w", err)
	}

	logger.Info("Cleanup of unused providers completed successfully.")
	return nil
}

// collectUsedProviders parses Terraform configuration files in all environments to collect used providers
func collectUsedProviders(envs []string, baseEnvDir string) (map[string]bool, error) {
	usedProviders := make(map[string]bool)
	parser := hclparse.NewParser()

	for _, env := range envs {
		envConfigDir := filepath.Join(baseEnvDir, env, "config", env)

		// Find all .tf files in the environment's configuration directory
		tfFiles, err := filepath.Glob(filepath.Join(envConfigDir, "*.tf"))
		if err != nil {
			logger.Warnf("Failed to list .tf files in %s: %v", envConfigDir, err)
			continue
		}

		if len(tfFiles) == 0 {
			logger.Warnf("No .tf files found in %s. Skipping environment.", envConfigDir)
			continue
		}

		for _, tfFile := range tfFiles {
			fileContent, diags := parser.ParseHCLFile(tfFile)
			if diags.HasErrors() {
				logger.Warnf("Failed to parse %s: %s", tfFile, diags.Error())
				continue
			}

			// Extract the terraform block
			content, _, diags := fileContent.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "terraform",
						LabelNames: []string{},
					},
				},
			})

			if diags.HasErrors() {
				logger.Warnf("Failed to get terraform blocks in %s: %s", tfFile, diags.Error())
				continue
			}

			for _, block := range content.Blocks.OfType("terraform") {
				// Within terraform block, look for required_providers
				terraformBodyContent, _, diags := block.Body.PartialContent(&hcl.BodySchema{
					Blocks: []hcl.BlockHeaderSchema{
						{
							Type:       "required_providers",
							LabelNames: []string{},
						},
					},
				})

				if diags.HasErrors() {
					logger.Warnf("Failed to get required_providers in %s: %s", tfFile, diags.Error())
					continue
				}

				for _, rpBlock := range terraformBodyContent.Blocks.OfType("required_providers") {
					// Each label in required_providers block corresponds to a provider
					// Example:
					// required_providers {
					//     aws = {
					//         source  = "hashicorp/aws"
					//         version = "~> 3.0"
					//     }
					// }

					// Extract provider configurations
					rpContent, diags := rpBlock.Body.Content(&hcl.BodySchema{
						Attributes: []hcl.AttributeSchema{
							{Name: "source"},
							{Name: "version"},
						},
					})

					if diags.HasErrors() {
						logger.Warnf("Failed to parse required_providers in %s: %s", tfFile, diags.Error())
						continue
					}

					// Iterate over each provider defined in required_providers
					for providerName, attr := range rpContent.Attributes {
						// Extract source
						sourceVal, diags := attr.Expr.Value(nil)
						if diags.HasErrors() {
							logger.Warnf("Failed to get source for provider %s in %s: %s", providerName, tfFile, diags.Error())
							continue
						}
						sourceStr := sourceVal.AsString()

						// Extract version constraint
						versionAttr, exists := rpContent.Attributes["version"]
						if !exists {
							logger.Warnf("Provider %s in %s missing 'version' attribute. Skipping.", providerName, tfFile)
							continue
						}
						versionVal, diags := versionAttr.Expr.Value(nil)
						if diags.HasErrors() {
							logger.Warnf("Failed to get version for provider %s in %s: %s", providerName, tfFile, diags.Error())
							continue
						}
						versionStr := versionVal.AsString()

						// Resolve the exact version using version constraints
						exactVersion, err := resolveProviderVersion(sourceStr, versionStr)
						if err != nil {
							logger.Warnf("Failed to resolve version for provider %s: %v", providerName, err)
							continue
						}

						// Add to usedProviders map
						key := fmt.Sprintf("%s_%s", extractProviderShortName(sourceStr), exactVersion)
						usedProviders[key] = true
						logger.Debugf("Environment '%s' uses provider '%s' version '%s'", env, extractProviderShortName(sourceStr), exactVersion)
					}
				}
			}
		}
	}

	return usedProviders, nil
}

// extractProviderShortName extracts the provider name from the source string
// For example, "hashicorp/aws" -> "aws"
func extractProviderShortName(source string) string {
	parts := strings.Split(source, "/")
	if len(parts) != 2 {
		return source // Fallback to the whole source if unexpected format
	}
	return parts[1]
}

// resolveProviderVersion resolves the exact provider version based on the version constraint
// This is a simplified resolver that assumes the latest version satisfying the constraint is used
func resolveProviderVersion(source, constraint string) (string, error) {
	// Fetch available versions for the provider
	availableVersions, err := getProviderAvailableVersions(source)
	if err != nil {
		return "", fmt.Errorf("failed to fetch available versions for provider %s: %w", source, err)
	}

	// Parse the constraint
	vConstraint, err := version.NewConstraint(constraint)
	if err != nil {
		return "", fmt.Errorf("invalid version constraint '%s' for provider %s: %w", constraint, source, err)
	}

	// Find the latest version that satisfies the constraint
	sort.Sort(sort.Reverse(version.Collection(availableVersions)))
	for _, v := range availableVersions {
		if vConstraint.Check(v) {
			return v.Original(), nil
		}
	}

	return "", fmt.Errorf("no versions found for provider %s matching constraint '%s'", source, constraint)
}

// getProviderAvailableVersions fetches all available versions for a given provider source
// This function needs to be implemented to fetch provider versions from the provider registry
func getProviderAvailableVersions(source string) ([]*version.Version, error) {
	// Example implementation for HashiCorp providers
	// For other sources, adjust accordingly
	parts := strings.Split(source, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unsupported provider source format: %s", source)
	}

	publisher := parts[0]
	provider := parts[1]

	if publisher != "hashicorp" {
		return nil, fmt.Errorf("only hashicorp provider sources are supported currently")
	}

	// Construct the API URL for the provider
	// Example: https://registry.terraform.io/v1/providers/hashicorp/aws/versions
	apiURL := fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/%s/versions", publisher, provider)

	// Fetch the provider versions from the API
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch provider versions from %s: %w", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch provider versions from %s: status code %d", apiURL, resp.StatusCode)
	}

	// Define the structure to parse the JSON response
	type providerVersionResponse struct {
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	}

	var pvResp providerVersionResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&pvResp); err != nil {
		return nil, fmt.Errorf("failed to decode provider versions JSON: %w", err)
	}

	// Parse and collect the versions
	var versions []*version.Version
	for _, v := range pvResp.Versions {
		ver, err := version.NewVersion(v.Version)
		if err != nil {
			logger.Warnf("Invalid provider version '%s': %v", v.Version, err)
			continue
		}
		versions = append(versions, ver)
	}

	return versions, nil
}

// cleanDuplicateProviders removes duplicate provider versions from the plugin cache
func cleanDuplicateProviders(pluginCacheDir string) error {
	logger.Infof("Cleaning duplicate providers in plugin cache at %s", pluginCacheDir)

	// Map to track unique providers
	uniqueProviders := make(map[string]bool)

	err := filepath.Walk(pluginCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Identify provider plugin files
		filename := info.Name()
		if strings.HasPrefix(filename, "terraform-provider-") {
			// Example filename: terraform-provider-aws_v3.74.0_x5
			parts := strings.Split(filename, "_v")
			if len(parts) != 2 {
				return nil
			}
			provider := strings.TrimPrefix(parts[0], "terraform-provider-")
			version := strings.TrimSuffix(parts[1], ".zip") // Adjust based on actual file extension

			key := fmt.Sprintf("%s_%s", provider, version)
			if uniqueProviders[key] {
				// Duplicate found, remove the file
				logger.Infof("Removing duplicate provider plugin: %s", path)
				err := os.Remove(path)
				if err != nil {
					logger.Warnf("Failed to remove %s: %v", path, err)
				} else {
					fmt.Printf("Removed duplicate provider plugin: %s\n", path)
				}
			} else {
				uniqueProviders[key] = true
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error cleaning duplicate providers: %w", err)
	}

	return nil
}

// initializeLogger sets up the logger with Logrus and ensures sensitive data is not logged
func initializeLogger() {
    logFile := os.Getenv("TFVENV_LOG_FILE")
    if logFile != "" {
        file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            fmt.Println("error opening log file:", err) // Starts with lowercase
            os.Exit(1)
        }
        logger.Out = file
    } else {
        logger.Out = os.Stdout
    }

    // Check for verbose logging
    if os.Getenv("TFVENV_VERBOSE") == "true" {
        logger.SetLevel(logrus.DebugLevel)
    } else {
        logger.SetLevel(logrus.InfoLevel)
    }

    // Use JSON formatter for structured logging
    logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })

    // Add MaskHook to sanitize logs
    logger.AddHook(&MaskHook{
        SensitiveKeys: []string{
            "AWS_ACCESS_KEY",
            "AWS_SECRET_KEY",
            "SNAP_KEY",
            // Add other sensitive keys here
        },
    })
}
// MaskHook is a Logrus hook to mask sensitive fields
type MaskHook struct {
    SensitiveKeys []string
}

// Levels specifies the log levels at which the MaskHook should be triggered.
// This implementation returns all available log levels.
func (hook *MaskHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

// Fire is called by Logrus when a log entry is fired.
// It masks the values of sensitive keys by replacing them with asterisks.
func (hook *MaskHook) Fire(entry *logrus.Entry) error {
    for _, key := range hook.SensitiveKeys {
        if _, ok := entry.Data[key]; ok {
            entry.Data[key] = "******" // Mask the value
        }
    }
    return nil
}

// isSensitiveKey determines if the environment variable is sensitive
func isSensitiveKey(key string) bool {
    sensitiveKeys := []string{
        "AWS_ACCESS_KEY",
        "AWS_SECRET_KEY",
        "SNAP_KEY",
        // Add other sensitive keys here
    }
    for _, k := range sensitiveKeys {
        if key == k {
            return true
        }
    }
    return false
}

// downloadFile downloads a file from a URL and saves it to the specified destination path
func downloadFile(url, dest string) error {
	logger.Infof("Downloading from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to initiate download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// getLatestVersion fetches the latest version for the specified tool
// tool: "terraform" or "terragrunt"
// includePreReleases: applicable only for Terragrunt
func getLatestVersion(tool string, includePreReleases bool) (string, error) {
	switch strings.ToLower(tool) {
	case "terraform":
		return getLatestTerraformVersion()
	case "terragrunt":
		return getLatestTerragruntVersion(includePreReleases)
	default:
		return "", fmt.Errorf("unsupported tool: %s", tool)
	}
}

// getLatestTerragruntVersion fetches the latest Terragrunt version from GitHub Releases
// If includePreReleases is true, it includes pre-releases in the search
func getLatestTerragruntVersion(includePreReleases bool) (string, error) {
	apiURL := "https://api.github.com/repos/gruntwork-io/terragrunt/releases"

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Terragrunt releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch Terragrunt releases: status code %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&releases); err != nil {
		return "", fmt.Errorf("failed to decode Terragrunt releases: %w", err)
	}

	versions := make([]*version.Version, 0)
	for _, release := range releases {
		if !includePreReleases && release.Prerelease {
			continue
		}
		// Clean version string, remove leading 'v' if present
		cleanVerStr := strings.TrimPrefix(release.TagName, "v")
		ver, err := version.NewVersion(cleanVerStr)
		if err != nil {
			logger.Warnf("invalid Terragrunt version format: %s", release.TagName)
			continue
		}
		versions = append(versions, ver)
	}

	if len(versions) == 0 {
		return "", errors.New("no valid Terragrunt versions found")
	}

	// Sort versions in descending order
	sort.Sort(sort.Reverse(version.Collection(versions)))

	latestVersion := versions[0].Original()
	return latestVersion, nil
}

func getLatestTerraformVersion() (string, error) {
	indexURL := "https://releases.hashicorp.com/terraform/index.json"

	resp, err := http.Get(indexURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Terraform release index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch Terraform release index: status code %d", resp.StatusCode)
	}

	var index TerraformReleaseIndex
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&index); err != nil {
		return "", fmt.Errorf("failed to decode Terraform release index: %w", err)
	}

	versions := make([]*version.Version, 0, len(index.Versions))
	for verStr := range index.Versions {
		cleanVerStr := strings.TrimPrefix(verStr, "v")
		ver, err := version.NewVersion(cleanVerStr)
		if err != nil {
			logger.Warnf("invalid Terraform version format: %s", verStr)
			continue
		}
		versions = append(versions, ver)
	}

	if len(versions) == 0 {
		return "", errors.New("no valid Terraform versions found")
	}

	// Sort versions in descending order
	sort.Sort(sort.Reverse(version.Collection(versions)))

	latestVersion := versions[0].Original()
	return latestVersion, nil
}

// unzipStandard extracts a zip file to the specified destination using the standard library
func unzipStandard(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", src, err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// Create directory
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create necessary directories
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(fpath), err)
		}

		// Create the file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fpath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open file %s in zip: %w", f.Name, err)
		}

		// Copy contents
		_, err = io.Copy(outFile, rc)

		// Close files
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to copy contents to file %s: %w", fpath, err)
		}
	}
	return nil
}

// copyAndCustomizeConfig copies the template config and replaces placeholders with actual values
func copyAndCustomizeConfig(templatePath, destPath, tfVersion, tgVersion, environment string) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Replace placeholders with actual values
	contentStr := string(content)
	contentStr = strings.ReplaceAll(contentStr, "{{TF_VERSION}}", tfVersion)
	contentStr = strings.ReplaceAll(contentStr, "{{TG_VERSION}}", tgVersion)
	contentStr = strings.ReplaceAll(contentStr, "{{ENVIRONMENT}}", environment)

	err = os.WriteFile(destPath, []byte(contentStr), 0644)
	if err != nil {
		return fmt.Errorf("failed to write customized config to %s: %w", destPath, err)
	}

	logger.Infof("Customized config written to %s", destPath)
	return nil
}

// customizeTerragruntHcl customizes the terragrunt.hcl file with S3 backend configuration
func customizeTerragruntHcl(templatePath, destPath string, config Config) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Replace placeholders with actual S3 backend configuration
	contentStr := string(content)
	contentStr = strings.ReplaceAll(contentStr, "{{S3_STATE_BUCKET}}", config.S3StateBucket)
	contentStr = strings.ReplaceAll(contentStr, "{{S3_STATE_PATH}}", config.S3StatePath)
	contentStr = strings.ReplaceAll(contentStr, "{{REGION}}", config.Region)

	err = os.WriteFile(destPath, []byte(contentStr), 0644)
	if err != nil {
		return fmt.Errorf("failed to write customized terragrunt.hcl to %s: %w", destPath, err)
	}

	logger.Infof("Customized terragrunt.hcl written to %s", destPath)
	return nil
}

// getEnvironments lists all environment types in the envDir
func getEnvironments(envDir string) ([]string, error) {
	configPath := filepath.Join(envDir, "config")
	entries, err := os.ReadDir(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory %s: %w", configPath, err)
	}

	envs := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			envs = append(envs, entry.Name())
		}
	}

	return envs, nil
}

// runHclfmt runs hclfmt on terragrunt.hcl
func runHclfmt(envDir, envType string, check bool) error {
	terragruntPath := filepath.Join(envDir, "config", envType, fmt.Sprintf("terragrunt.%s.hcl", envType))
	if !fileExists(terragruntPath) {
		logger.Warnf("terragrunt.hcl file %s not found. Skipping hclfmt.", terragruntPath)
		return nil
	}

	tgBinary := filepath.Join(envDir, "bin", "terragrunt")

	// Check if the terragrunt binary exists
	if !fileExists(tgBinary) {
		// Log a warning and skip formatting if the binary is not found
		logger.Warnf("terragrunt binary not found at %s. Skipping hclfmt.", tgBinary)
		fmt.Printf("Warning: terragrunt binary not found at %s. Skipping hclfmt.\n", tgBinary)
		return nil
	}

	// Prepare the hclfmt command
	var cmdTg *exec.Cmd
	if check {
		cmdTg = exec.Command(tgBinary, "hclfmt", "--terragrunt-check", terragruntPath)
	} else {
		cmdTg = exec.Command(tgBinary, "hclfmt", terragruntPath)
	}

	cmdTg.Dir = filepath.Dir(terragruntPath)
	output, err := cmdTg.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hclfmt failed: %v, output: %s", err, string(output))
	}

	logger.Infof("hclfmt completed successfully for %s", terragruntPath)
	return nil
}
// lockCmd locks the environment to prevent concurrent modifications
func lockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock <env-name>",
		Short: "Lock the specified environment to prevent concurrent operations",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)

			lockPath := filepath.Join(envPath, lockFileName)
			if fileExists(lockPath) {
				fmt.Println("Environment is already locked.")
				os.Exit(1)
			}

			file, err := os.Create(lockPath)
			if err != nil {
				logger.Errorf("Failed to create lock file: %v", err)
				fmt.Printf("Error locking environment: %v\n", err)
				os.Exit(1)
			}
			file.Close()
			fmt.Println("Environment locked successfully.")
		},
	}

	return cmd
}
// unlockCmd unlocks the environment
func unlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock <env-name>",
		Short: "Unlock the specified environment to allow operations",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)

			lockPath := filepath.Join(envPath, lockFileName)
			if !fileExists(lockPath) {
				fmt.Println("Environment is not locked.")
				os.Exit(1)
			}

			err := os.Remove(lockPath)
			if err != nil {
				logger.Errorf("Failed to remove lock file: %v", err)
				fmt.Printf("Error unlocking environment: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Environment unlocked successfully.")
		},
	}

	return cmd
}
// completionCmd generates shell completion scripts
func completionCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(tfvenv completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ tfvenv completion bash > /etc/bash_completion.d/tfvenv
  # macOS:
  $ tfvenv completion bash > /usr/local/etc/bash_completion.d/tfvenv

Zsh:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  $ tfvenv completion zsh > "${fpath[1]}/_tfvenv"

Fish:

  $ tfvenv completion fish | source

  # To load completions for each session, execute once:
  $ tfvenv completion fish > ~/.config/fish/completions/tfvenv.fish

PowerShell:

  PS> tfvenv completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> tfvenv completion powershell > tfvenv.ps1
  # and source this file from your PowerShell profile.
`,
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			cobra.OnlyValidArgs,
		),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	return cmd
}

// createCmd handles the creation of a new environment
func createCmd() *cobra.Command {
	var tfVersion, tgVersion string

	cmd := &cobra.Command{
		Use:   "create <env-name> [tf-version] [tg-version]",
		Short: "Create a new virtual environment for Terraform and optionally Terragrunt",
		Args:  cobra.MinimumNArgs(1), // Ensure at least 1 argument is provided
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0] // First argument: environment name

			// Retrieve envDir from persistent flags via Viper
			envDir := viper.GetString("env-dir")

			// If `tf-version` is not provided, default to "latest"
			tfVersion := "latest"
			if len(args) > 1 {
				tfVersion = args[1]
			}

			// If `tg-version` is not provided, set to "none"
			tgVersion := "none"
			if len(args) > 2 {
				tgVersion = args[2]
			}

			// Construct environment directory
			envDirPath := filepath.Join(envDir, envName)
			if strings.ToLower(envName) == "previous" {
				fmt.Println("Cannot use 'previous' as an environment name.")
				logger.Error("attempted to use 'previous' as environment name")
				os.Exit(1)
			}
			// Initialize the environment
			err := initEnv(envDirPath, tfVersion, tgVersion, envName)
			if err != nil {
				logger.Errorf("error creating environment %s: %v", envName, err) // Lowercase and use logger
				fmt.Printf("Error creating environment '%s': %v\n", envName, err)
				os.Exit(1)
			}

			fmt.Printf("Environment '%s' created successfully with Terraform %s and Terragrunt %s.\n", envName, tfVersion, tgVersion)
			logger.Infof("Environment '%s' created successfully with Terraform %s and Terragrunt %s.", envName, tfVersion, tgVersion) // Log success
		},
	}

	// Define specific flags for the create command
	cmd.Flags().StringVar(&tfVersion, "tf-version", "latest", "Terraform version to create the environment with")
	cmd.Flags().StringVar(&tgVersion, "tg-version", "none", "Terragrunt version to create the environment with")

	return cmd
}
// deleteCmd handles the deletion of an environment
func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <env-name>",
		Short: "Delete an existing virtual environment",
		Args:  cobra.ExactArgs(1), // Ensure exactly one argument is provided
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]

			// Retrieve envDir from persistent flags via Viper
			envDir := viper.GetString("env-dir")

			// Construct full path to the environment
			envPath := filepath.Join(envDir, envName)

			// Delete environment
			err := os.RemoveAll(envPath)
			if err != nil {
				logger.Errorf("error deleting environment at %s: %v", envPath, err) // Lowercase and use logger
				fmt.Printf("Error deleting environment '%s': %v\n", envName, err)
				os.Exit(1)
			}
			fmt.Printf("Environment '%s' deleted successfully.\n", envName)
			logger.Infof("Environment '%s' deleted successfully.", envName) // Log success
		},
	}

	return cmd
}
// listCmd lists all managed environments
func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all existing environments managed by tfvenv",
		Run: func(cmd *cobra.Command, args []string) {
			// Retrieve envDir from persistent flags via Viper
			envDir := viper.GetString("env-dir")

			envs, err := getEnvironments(envDir)
			if err != nil {
				logger.Errorf("error listing environments: %v", err) // Lowercase and use logger
				fmt.Printf("Error listing environments: %v\n", err)
				os.Exit(1)
			}

			if len(envs) == 0 {
				fmt.Println("No environments found.")
				logger.Info("no environments found") // Log info
				return
			}

			fmt.Println("Available Environments:")
			for _, env := range envs {
				fmt.Println(" -", env)
			}
			logger.Infof("listed %d environments", len(envs)) // Log success
		},
	}

	return cmd
}
// activateCmd activates an environment by generating the activate scripts and instructing the user to source the appropriate one.
func activateCmd() *cobra.Command {
	var envType string
	var customEnv string

	cmd := &cobra.Command{
		Use:   "activate <env-name>",
		Short: "Activate a virtual environment",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)
			configPath := filepath.Join(envPath, "config", envName, tfvenvrcFileName)

			// Read configuration
			config, err := readConfig(configPath)
			if err != nil {
				logger.Errorf("error reading %s: %v", configPath, err) // Lowercase and use logger
				fmt.Printf("Error reading configuration: %v\n", err)
				os.Exit(1)
			}

			// Parse customEnv
			if customEnv != "" {
				customVars := strings.Split(customEnv, ",")
				for _, kv := range customVars {
					parts := strings.SplitN(kv, "=", 2)
					if len(parts) != 2 {
						fmt.Printf("Invalid environment variable format: %s\n", kv)
						continue
					}
					key := parts[0]
					value := parts[1]
					config.EnvVars[key] = value
				}
			}

			// Generate activate scripts for all supported shells
			err = generateActivateScript(envPath, envName, config)
			if err != nil {
				logger.Errorf("error generating activate scripts: %v", err) // Lowercase and use logger
				fmt.Printf("Error generating activate scripts: %v\n", err)
				os.Exit(1)
			}

			// Generate deactivate scripts for all supported shells
			err = generateDeactivateScript(envPath, config)
			if err != nil {
				logger.Errorf("error generating deactivate scripts: %v", err) // Lowercase and use logger
				fmt.Printf("Error generating deactivate scripts: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Environment '%s' activated successfully.\n", envName)
			logger.Infof("environment '%s' activated successfully", envName) // Log success
			fmt.Println("To activate the environment in your current shell, run the appropriate command below:")
			fmt.Printf("  Bash/Zsh:   source %s\n", filepath.Join(envPath, "bin", "activate.sh"))
			fmt.Printf("  Fish:       source %s\n", filepath.Join(envPath, "bin", "activate.fish"))
			fmt.Printf("  PowerShell: .\\%s\n", filepath.Join(envPath, "bin", "Activate.ps1"))
		},
	}

	// Define command-line flags
	cmd.Flags().StringVar(&envType, "env-type", "dev", "Environment type (e.g., dev, prod)")
	cmd.Flags().StringVar(&customEnv, "var", "", "Custom environment variables in key=value format, separated by commas")

	return cmd
}
// deactivateCmd deactivates an environment by generating the deactivate scripts and instructing the user to source the appropriate one.
func deactivateCmd() *cobra.Command {
	var envType string

	cmd := &cobra.Command{
		Use:   "deactivate <env-name>",
		Short: "Deactivate the current virtual environment",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)
			configPath := filepath.Join(envPath, "config", envName, tfvenvrcFileName)

			// Read configuration
			config, err := readConfig(configPath)
			if err != nil {
				logger.Errorf("error reading %s: %v", configPath, err) // Lowercase and use logger
				fmt.Printf("Error reading configuration: %v\n", err)
				os.Exit(1)
			}

			// Generate deactivate scripts for all supported shells
			err = generateDeactivateScript(envPath, config)
			if err != nil {
				logger.Errorf("error generating deactivate scripts: %v", err) // Lowercase and use logger
				fmt.Printf("Error generating deactivate scripts: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Environment '%s' deactivated successfully.\n", envName)
			logger.Infof("environment '%s' deactivated successfully", envName) // Log success
			fmt.Println("To deactivate the environment in your current shell, run the appropriate command below:")
			fmt.Printf("  Bash/Zsh:   source %s\n", filepath.Join(envPath, "bin", "deactivate.sh"))
			fmt.Printf("  Fish:       source %s\n", filepath.Join(envPath, "bin", "deactivate.fish"))
			fmt.Printf("  PowerShell: .\\%s\n", filepath.Join(envPath, "bin", "Deactivate.ps1"))
		},
	}

	// Define command-line flags
	cmd.Flags().StringVar(&envType, "env-type", "dev", "Environment type (e.g., dev, prod)")

	return cmd
}
// validateCmd validates .tfvars and terragrunt.hcl files
func validateCmd() *cobra.Command {
	var envType string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate .tfvars and terragrunt.hcl files",
		Run: func(cmd *cobra.Command, args []string) {
			// Construct the path to the tfvenvrc configuration file
			envDir := viper.GetString("env-dir")
			configPath := filepath.Join(envDir, "config", envType, tfvenvrcFileName)

			// Read and parse the configuration
			config, err := readConfig(configPath)
			if err != nil {
				logger.Errorf("error reading %s: %v", configPath, err) // Lowercase and use logger
				fmt.Printf("Error reading configuration: %v\n", err)
				os.Exit(1)
			}

			// Apply environment variables from the configuration
			applyEnvVars(config.EnvVars)

			// Paths to the .tfvars and terragrunt.hcl files
			tfvarsPath := filepath.Join(envDir, "config", envType, fmt.Sprintf("%s.tfvars", envType))
			terragruntPath := filepath.Join(envDir, "config", envType, fmt.Sprintf("terragrunt.%s.hcl", envType))

			// Validate .tfvars file using Terraform
			if fileExists(tfvarsPath) {
				tfBinary := filepath.Join(envDir, "bin", "terraform")

				// Check if Terraform binary exists
				if !fileExists(tfBinary) {
					logger.Errorf("terraform binary not found at %s", tfBinary) // Lowercase and use logger
					fmt.Printf("Terraform binary not found at %s\n", tfBinary)
					os.Exit(1)
				}

				cmdTf := exec.Command(tfBinary, "validate", "-var-file", tfvarsPath)
				cmdTf.Env = append(os.Environ(), "TF_DATA_DIR="+filepath.Join(envDir, "terraform-data"))

				output, err := cmdTf.CombinedOutput()
				if err != nil {
					logger.Errorf("validation failed for %s: %v", tfvarsPath, err) // Lowercase and use logger
					fmt.Println("Validation Error:", string(output))
					os.Exit(1)
				}
				fmt.Printf(".tfvars file %s is valid.\n", tfvarsPath)
				logger.Infof(".tfvars file %s is valid.", tfvarsPath) // Log success
			} else {
				fmt.Printf(".tfvars file %s not found.\n", tfvarsPath)
				logger.Warnf(".tfvars file %s not found.", tfvarsPath) // Log warning
			}

			// Validate terragrunt.hcl file using Terragrunt
			if fileExists(terragruntPath) {
				tgBinary := filepath.Join(envDir, "bin", "terragrunt")

				// Check if Terragrunt binary exists
				if !fileExists(tgBinary) {
					logger.Errorf("terragrunt binary not found at %s", tgBinary) // Lowercase and use logger
					fmt.Printf("Terragrunt binary not found at %s\n", tgBinary)
					os.Exit(1)
				}

				cmdTg := exec.Command(tgBinary, "hclfmt", "--terragrunt-check", terragruntPath)
				cmdTg.Dir = filepath.Dir(terragruntPath)

				output, err := cmdTg.CombinedOutput()
				if err != nil {
					logger.Errorf("validation failed for %s: %v", terragruntPath, err) // Lowercase and use logger
					fmt.Println("Validation Error:", string(output))
					os.Exit(1)
				}
				fmt.Printf("terragrunt.hcl file %s is valid.\n", terragruntPath)
				logger.Infof("terragrunt.hcl file %s is valid.", terragruntPath) // Log success
			} else {
				fmt.Printf("terragrunt.hcl file %s not found.\n", terragruntPath)
				logger.Warnf("terragrunt.hcl file %s not found.", terragruntPath) // Log warning
			}

			logger.Info("validation completed successfully") // Log success
		},
	}

	// Define command-line flags
	cmd.Flags().StringVar(&envType, "env-type", "dev", "Environment type (e.g., dev, prod)")
	return cmd
}
// statusCmd shows the status of the environment
func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <env-name>",
		Short: "Display the status of the specified environment",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)
			configPath := filepath.Join(envPath, "config", envName, tfvenvrcFileName)

			// Read configuration
			config, err := readConfig(configPath)
			if err != nil {
				logger.Errorf("error reading %s: %v", configPath, err) // Lowercase and use logger
				fmt.Printf("Error reading configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Status of environment '%s' (%s):\n", envName, envDir)

			// Check if Terraform and Terragrunt are installed
			tfPath := filepath.Join(envDir, "bin", "terraform")
			tgPath := filepath.Join(envDir, "bin", "terragrunt")
			if fileExists(tfPath) {
				fmt.Printf("Terraform installed at %s\n", tfPath)
				logger.Infof("Terraform installed at %s", tfPath) // Log info
			} else {
				fmt.Println("Terraform not found in environment.")
				logger.Warnf("Terraform not found in environment at %s", tfPath) // Log warning
			}
			if fileExists(tgPath) {
				fmt.Printf("Terragrunt installed at %s\n", tgPath)
				logger.Infof("Terragrunt installed at %s", tgPath) // Log info
			} else {
				fmt.Println("Terragrunt not found in environment.")
				logger.Warnf("Terragrunt not found in environment at %s", tgPath) // Log warning
			}

			// Display active environment variables
			fmt.Println("Environment Variables:")
			for key, value := range config.EnvVars {
				if isSensitiveKey(key) {
					fmt.Printf(" - %s=******\n", key) // Masked output
				} else {
					fmt.Printf(" - %s=%s\n", key, value)
				}
				// Optionally log environment variables without values
				logger.Infof("environment variable set: %s", key)
			}

			logger.Info("status displayed successfully") // Log success
		},
	}

	return cmd
}

// / downloadAndInstallBinary downloads and installs the specified binary.
// It handles different OS and package types for Windows, Linux, and macOS.
func downloadAndInstallBinary(baseURL, version, binDir, tool string) error {
	binaryName := tool
	binaryPath := filepath.Join(binDir, binaryName)

	// Append .exe extension for Windows binaries
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
		binaryPath = filepath.Join(binDir, binaryName)
	}

	// Check if the binary is already installed
	if fileExists(binaryPath) {
		logger.Infof("Checking existing %s version...", tool)
		existingVersion, err := getBinaryVersion(binaryPath, tool)
		if err != nil {
			logger.Warnf("Failed to get existing %s version: %v", tool, err)
		} else {
			if version == "latest" {
				latestVersion, err := getLatestVersion(tool, false)
				if err != nil {
					logger.Warnf("Failed to fetch latest version for %s: %v", tool, err)
				} else if existingVersion == latestVersion {
					fmt.Printf("`%s version %s` already installed at %s\n", tool, existingVersion, binaryPath)
					return nil
				}
			} else {
				if existingVersion == version {
					fmt.Printf("`%s version %s` already installed at %s\n", tool, existingVersion, binaryPath)
					return nil
				}
			}
		}
	}

	// Handle "latest" version specification
	if version == "latest" {
		latest, err := getLatestVersion(tool, false) // Set to true if pre-releases should be included
		if err != nil {
			return fmt.Errorf("failed to fetch latest version for %s: %w", tool, err)
		}
		version = latest
		logger.Infof("Using latest version for %s: %s", tool, version)
	}

	// Define file paths and URLs
	var downloadURL, destPath string

	// Construct the download URL based on the tool and OS
	switch tool {
	case "terraform":
		// Terraform is distributed as a zip archive across all OSes
		// Example: https://releases.hashicorp.com/terraform/1.9.7/terraform_1.9.7_linux_amd64.zip
		downloadURL = fmt.Sprintf("%s%s/terraform_%s_%s_%s.zip", baseURL, version, version, runtime.GOOS, runtime.GOARCH)
		destPath = filepath.Join(binDir, "terraform.zip")
	case "terragrunt":
		// Terragrunt binaries are direct downloads, with .exe for Windows
		// Example for Linux: https://github.com/gruntwork-io/terragrunt/releases/download/v0.67.16/terragrunt_linux_amd64
		// Example for Windows: https://github.com/gruntwork-io/terragrunt/releases/download/v0.67.16/terragrunt_windows_amd64.exe
		if runtime.GOOS == "windows" {
			downloadURL = fmt.Sprintf("%sv%s/terragrunt_%s_%s.exe", baseURL, version, runtime.GOOS, runtime.GOARCH)
			destPath = filepath.Join(binDir, "terragrunt.exe")
		} else {
			downloadURL = fmt.Sprintf("%sv%s/terragrunt_%s_%s", baseURL, version, runtime.GOOS, runtime.GOARCH)
			destPath = filepath.Join(binDir, "terragrunt")
		}
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}

	fmt.Printf("Downloading %s version %s...\n", tool, version)
	logger.Infof("Downloading %s from %s", tool, downloadURL)

	// Download the binary
	err := downloadFile(downloadURL, destPath)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", tool, err)
	}

	// Post-download processing based on the tool
	switch tool {
	case "terraform":
		// Unzip the Terraform archive
		err = unzipStandard(destPath, binDir)
		if err != nil {
			return fmt.Errorf("failed to unzip Terraform: %w", err)
		}
		// Clean up zip file after extraction
		os.Remove(destPath)

		// Rename the binary if necessary (e.g., adding .exe for Windows)
		extractedBinaryName := "terraform"
		if runtime.GOOS == "windows" {
			extractedBinaryName += ".exe"
		}
		extractedBinaryPath := filepath.Join(binDir, extractedBinaryName)
		if fileExists(extractedBinaryPath) {
			// Overwrite the existing binaryPath with the extracted binary
			err = os.Rename(extractedBinaryPath, binaryPath)
			if err != nil {
				return fmt.Errorf("failed to rename Terraform binary: %w", err)
			}
		}
	case "terragrunt":
		// Ensure Terragrunt binary is executable
		if runtime.GOOS != "windows" {
			if err := os.Chmod(destPath, 0755); err != nil {
				return fmt.Errorf("failed to set execute permissions on %s: %w", destPath, err)
			}
		}
	default:
		// No additional processing for unknown tools
	}

	// Verify the installed version
	installedVersion, err := getBinaryVersion(binaryPath, tool)
	if err != nil {
		return fmt.Errorf("failed to verify installed %s version: %w", tool, err)
	}

	if version != "latest" && installedVersion != version {
		return fmt.Errorf("%s version mismatch: expected %s, got %s", tool, version, installedVersion)
	}

	fmt.Printf("Installed: `%s version %s`\n", tool, installedVersion)
	logger.Infof("%s version %s installed successfully at %s", tool, installedVersion, binaryPath)

	return nil
}

// getBinaryVersion retrieves the version of the installed binary.
func getBinaryVersion(binaryPath, tool string) (string, error) {
	var cmd *exec.Cmd
	if tool == "terraform" {
		cmd = exec.Command(binaryPath, "version")
	} else if tool == "terragrunt" {
		cmd = exec.Command(binaryPath, "--version")
	} else {
		return "", fmt.Errorf("unsupported tool for version retrieval: %s", tool)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute %s: %w", tool, err)
	}

	outputStr := string(output)
	// Parse the version from the output
	var version string
	if tool == "terraform" {
		// Example output: Terraform v1.9.7
		parts := strings.Fields(outputStr)
		if len(parts) >= 2 {
			version = strings.TrimPrefix(parts[1], "v")
		}
	} else if tool == "terragrunt" {
		// Example output: terragrunt version v0.67.16
		parts := strings.Fields(outputStr)
		if len(parts) >= 3 {
			version = strings.TrimPrefix(parts[2], "v")
		}
	}

	if version == "" {
		return "", fmt.Errorf("could not parse %s version from output: %s", tool, outputStr)
	}

	return version, nil
}

// upgradeBinaries upgrades Terraform and Terragrunt binaries to specified versions.
// It provides detailed logging for each step.
func upgradeBinaries(envDir, tfVersion, tgVersion string) error {
    logger.Infof("upgrading binaries in environment %s", envDir) // Lowercase log message

    // Paths to the bin directory
    binDir := filepath.Join(envDir, "bin")

    // Upgrade Terraform
    fmt.Printf("Upgrading Terraform to version %s...\n", tfVersion)
    logger.Infof("upgrading Terraform to version %s", tfVersion) // Lowercase log message
    err := downloadAndInstallBinary(terraformDownloadURL, tfVersion, binDir, "terraform")
    if err != nil {
        logger.Errorf("error upgrading Terraform: %v", err) // Lowercase and use logger
        return fmt.Errorf("failed to upgrade Terraform: %w", err)
    }
    fmt.Printf("Terraform upgraded to version %s\n", tfVersion)
    logger.Infof("Terraform upgraded to version %s", tfVersion) // Log success

    // Upgrade Terragrunt
    fmt.Printf("Upgrading Terragrunt to version %s...\n", tgVersion)
    logger.Infof("upgrading Terragrunt to version %s", tgVersion) // Lowercase log message
    err = downloadAndInstallBinary(terragruntDownloadURL, tgVersion, binDir, "terragrunt")
    if err != nil {
        logger.Errorf("error upgrading Terragrunt: %v", err) // Lowercase and use logger
        return fmt.Errorf("failed to upgrade Terragrunt: %w", err)
    }
    fmt.Printf("Terragrunt upgraded to version %s\n", tgVersion)
    logger.Infof("Terragrunt upgraded to version %s", tgVersion) // Log success

    return nil
}
// upgradeCmd upgrades Terraform and Terragrunt binaries to specified versions.
func upgradeCmd() *cobra.Command {
	var tfVersion, tgVersion string

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Terraform and Terragrunt binaries to specified versions",
		Run: func(cmd *cobra.Command, args []string) {
			envDir := viper.GetString("env-dir")

			// Use specified version or default to "latest"
			if tfVersion == "" {
				tfVersion = "latest"
			}
			if tgVersion == "" {
				tgVersion = "latest"
			}

			// Upgrade binaries
			err := upgradeBinaries(envDir, tfVersion, tgVersion)
			if err != nil {
				logger.Errorf("Upgrade failed: %v", err)
				fmt.Printf("Error upgrading binaries: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Upgrade completed successfully.")
		},
	}

	// Define specific flags for the upgrade command
	cmd.Flags().StringVar(&tfVersion, "tf-version", "latest", "Terraform version to upgrade to")
	cmd.Flags().StringVar(&tgVersion, "tg-version", "latest", "Terragrunt version to upgrade to")

	return cmd
}
// hclfmtCmd formats or checks .hcl files in the environment
func hclfmtCmd() *cobra.Command {
	var envType string
	var check bool

	cmd := &cobra.Command{
		Use:   "hclfmt <env-name>",
		Short: "Format or check .hcl files within the specified environment",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			envDir := viper.GetString("env-dir")
			envPath := filepath.Join(envDir, envName)
			configPath := filepath.Join(envPath, "config", envType, tfvenvrcFileName)

			// Read and parse the configuration
			config, err := readConfig(configPath)
			if err != nil {
				logger.Errorf("error reading %s: %v", configPath, err)
				fmt.Printf("Error reading configuration: %v\n", err)
				os.Exit(1)
			}

			// Apply environment variables from the configuration
			applyEnvVars(config.EnvVars)

			// Run hclfmt
			err = runHclfmt(envDir, envType, check)
			if err != nil {
				logger.Errorf("hclfmt failed: %v", err)
				fmt.Printf("hclfmt Error: %v\n", err)
				os.Exit(1)
			}

			if check {
				fmt.Println("hclfmt check passed successfully.")
			} else {
				fmt.Println("hclfmt completed successfully.")
			}

			logger.Info("hclfmt completed successfully") // Log success
		},
	}

	// Define command-line flags
	cmd.Flags().StringVar(&envType, "env-type", "dev", "Environment type (e.g., dev, prod)")
	cmd.Flags().BoolVar(&check, "check", false, "Check formatting without making changes")

	return cmd
}

// initEnv initializes a new environment with detailed logging and ensures plugin cache is centralized.
func initEnv(envDir, tfVersion, tgVersion, environment string) error {
	logger.Infof("Initializing environment %s/%s", envDir, environment)

	// Define centralized plugin cache directory and Terraform data directory
	pluginCacheDir := filepath.Join(os.Getenv("HOME"), ".tfvenv", "plugin-cache")
	tfDataDir := filepath.Join(envDir, "terraform-data")

	// Create the plugin cache and Terraform data directories
	if err := os.MkdirAll(pluginCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin cache directory: %w", err)
	}
	if err := os.MkdirAll(tfDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create Terraform data directory: %w", err)
	}

	// Create necessary directory structure
	binDir := filepath.Join(envDir, "bin")
	configEnvDir := filepath.Join(envDir, "config", environment)
	templatesDir := filepath.Join(envDir, "templates")

	directories := []string{binDir, configEnvDir, templatesDir}
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Download and install Terraform
	fmt.Printf("Installing Terraform...\n")
	logger.Infof("Downloading and installing Terraform version %s", tfVersion)
	err := downloadAndInstallBinary(terraformDownloadURL, tfVersion, binDir, "terraform")
	if err != nil {
		return fmt.Errorf("failed to download Terraform: %w", err)
	}

	// Download and install Terragrunt if not specified as "none"
	if tgVersion != "none" {
		fmt.Printf("Installing Terragrunt...\n")
		logger.Infof("Downloading and installing Terragrunt version %s", tgVersion)
		err = downloadAndInstallBinary(terragruntDownloadURL, tgVersion, binDir, "terragrunt")
		if err != nil {
			return fmt.Errorf("failed to download Terragrunt: %w", err)
		}
	}

	// Paths for template files
	tfvarsTemplatePath := filepath.Join(templatesDir, fmt.Sprintf("%s.tfvars.template", environment))
	terragruntTemplatePath := filepath.Join(templatesDir, fmt.Sprintf("terragrunt.%s.hcl.template", environment))

	// Create default template files if they don't exist
	if !fileExists(tfvarsTemplatePath) {
		if err := createDefaultTfvarsTemplate(tfvarsTemplatePath); err != nil {
			return fmt.Errorf("failed to create default .tfvars template: %w", err)
		}
		logger.Infof("Default .tfvars template created at %s", tfvarsTemplatePath)
	}

	if !fileExists(terragruntTemplatePath) {
		if err := createDefaultTerragruntTemplate(terragruntTemplatePath); err != nil {
			return fmt.Errorf("failed to create default terragrunt.hcl template: %w", err)
		}
		logger.Infof("Default terragrunt.hcl template created at %s", terragruntTemplatePath)
	}

	// Customize and create .tfvars file
	tfvarsPath := filepath.Join(configEnvDir, fmt.Sprintf("%s.tfvars", environment))
	err = copyAndCustomizeConfig(tfvarsTemplatePath, tfvarsPath, tfVersion, tgVersion, environment)
	if err != nil {
		return fmt.Errorf("failed to create .tfvars file from template: %w", err)
	}

	// Customize and create terragrunt.hcl file with EnvVars including TF_PLUGIN_CACHE_DIR and TF_DATA_DIR
	terragruntPath := filepath.Join(configEnvDir, fmt.Sprintf("terragrunt.%s.hcl", environment))
	err = customizeTerragruntHcl(terragruntTemplatePath, terragruntPath, Config{
		S3StateBucket: "your_s3_state_bucket", // Replace with actual values or pass through parameters
		S3StatePath:   "your_s3_state_path",
		Region:        "your_aws_region",
		EnvVars: map[string]string{
			"TF_PLUGIN_CACHE_DIR": pluginCacheDir,
			"TF_DATA_DIR":         tfDataDir,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create terragrunt.hcl file from template: %w", err)
	}

	// Apply hclfmt automatically using runHclfmt
	fmt.Printf("Formatting terragrunt.hcl...\n")
	err = runHclfmt(envDir, environment, false)
	if err != nil {
		logger.Errorf("hclfmt failed: %v", err)
		fmt.Printf("hclfmt Error: %v\n", err)
		return fmt.Errorf("failed to format terragrunt.hcl: %w", err)
	}
	logger.Infof("terragrunt.hcl formatted successfully")
	fmt.Printf("terragrunt.hcl formatted successfully.\n")

	// Read the complete configuration to pass to script generators
	completeConfig := Config{
		S3StateBucket: "your_s3_state_bucket",
		S3StatePath:   "your_s3_state_path",
		Region:        "your_aws_region",
		EnvVars: map[string]string{
			"TF_PLUGIN_CACHE_DIR": pluginCacheDir,
			"TF_DATA_DIR":         tfDataDir,
		},
	}

	// Generate activate scripts for all supported shells
	err = generateActivateScript(envDir, environment, completeConfig)
	if err != nil {
		return fmt.Errorf("failed to generate activation scripts: %w", err)
	}

	// Generate deactivate scripts for all supported shells
	err = generateDeactivateScript(envDir, completeConfig)
	if err != nil {
		return fmt.Errorf("failed to generate deactivation scripts: %w", err)
	}

	logger.Infof("Environment %s/%s initialized successfully", envDir, environment)
	fmt.Printf("Environment '%s' created successfully with Terraform %s and Terragrunt %s.\n", environment, tfVersion, tgVersion)
	fmt.Println(`Run the following to add the environment binaries to your PATH (if not already present):
      source ~/.bashrc

Or ensure the bin directory is added to your shell configuration.`)

	return nil
}


// createDefaultTfvarsTemplate creates a default .tfvars template file
func createDefaultTfvarsTemplate(path string) error {
	defaultContent := `
# Default .tfvars template
variable "example_variable" {
  description = "An example variable"
  type        = string
  default     = "default_value"
}
`
	if err := os.WriteFile(path, []byte(defaultContent), 0644); err != nil {
		return fmt.Errorf("failed to create .tfvars template: %w", err)
	}
	return nil
}

// createDefaultTerragruntTemplate creates a default terragrunt.hcl template file
func createDefaultTerragruntTemplate(path string) error {
	defaultContent := `
# Default terragrunt.hcl template
terraform {
  source = "./terraform"
}

inputs = {
  example_variable = "default_value"
}
`
	if err := os.WriteFile(path, []byte(defaultContent), 0644); err != nil {
		return fmt.Errorf("failed to create terragrunt.hcl template: %w", err)
	}
	return nil
}

// mergeConfigurations merges template configurations into the environment
func mergeConfigurations(envDir, envType string) error {
	logger.Infof("Merging configurations for environment %s/%s", envDir, envType)

	templatesDir := filepath.Join(envDir, "templates")
	configEnvDir := filepath.Join(envDir, "config", envType)

	// Paths to template files
	tfvarsTemplatePath := filepath.Join(templatesDir, fmt.Sprintf("%s.tfvars.template", envType))
	terragruntTemplatePath := filepath.Join(templatesDir, fmt.Sprintf("terragrunt.%s.hcl.template", envType))

	// Paths to environment files
	tfvarsPath := filepath.Join(configEnvDir, fmt.Sprintf("%s.tfvars", envType))
	terragruntPath := filepath.Join(configEnvDir, fmt.Sprintf("terragrunt.%s.hcl", envType))

	// Merge .tfvars
	if fileExists(tfvarsTemplatePath) && fileExists(tfvarsPath) {
		err := smartMerge(tfvarsTemplatePath, tfvarsPath)
		if err != nil {
			return fmt.Errorf("failed to merge .tfvars: %w", err)
		}
		fmt.Printf(".tfvars file merged successfully at %s\n", tfvarsPath)
	}

	// Merge terragrunt.hcl
	if fileExists(terragruntTemplatePath) && fileExists(terragruntPath) {
		err := smartMerge(terragruntTemplatePath, terragruntPath)
		if err != nil {
			return fmt.Errorf("failed to merge terragrunt.hcl: %w", err)
		}
		fmt.Printf("terragrunt.hcl file merged successfully at %s\n", terragruntPath)
	}

	return nil
}

// smartMerge merges the template file into the environment file without overwriting user modifications
func smartMerge(templatePath, envPath string) error {
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read environment file %s: %w", envPath, err)
	}

	// Simple merge: Append any lines from template that are missing in env file
	envLines := make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewReader(envContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			envLines[line] = true
		}
	}

	templateScanner := bufio.NewScanner(bytes.NewReader(templateContent))
	var mergedContent bytes.Buffer
	mergedContent.Write(envContent) // Start with existing env content

	for templateScanner.Scan() {
		line := strings.TrimSpace(templateScanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !envLines[line] {
			mergedContent.WriteString("\n" + line)
		}
	}

	// Write the merged content back to the environment file
	err = os.WriteFile(envPath, mergedContent.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write merged content to %s: %w", envPath, err)
	}

	return nil
}

// readConfig reads the tfvenvrc configuration file
func readConfig(configPath string) (Config, error) {
	var config Config
	viper.SetConfigFile(configPath)
	viper.SetConfigType("env") // Assuming tfvenvrc is in KEY=VALUE format

	if err := viper.ReadInConfig(); err != nil {
		return config, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// applyEnvVars applies custom environment variables from the configuration
func applyEnvVars(envVars map[string]string) {
	for key, value := range envVars {
		if currentVal, exists := os.LookupEnv(key); exists {
			fmt.Printf("Environment variable %s is already set to %s.\n", key, currentVal)
			fmt.Printf("Do you want to overwrite it with %s? (y/n): ", value)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Printf("Skipping %s\n", key)
				continue
			}
		}
		fmt.Printf("Setting environment variable %s to %s\n", key, value)
		logger.Infof("Setting environment variable %s to %s", key, value)
		os.Setenv(key, value)
	}
}
func mergeCmd() *cobra.Command {
    var envType string

    cmd := &cobra.Command{
        Use:   "merge <env-name>",
        Short: "Merge changes from template configurations into the specified environment",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            envName := args[0]
            envDir := viper.GetString("env-dir")
            envPath := filepath.Join(envDir, envName)

            // Call mergeConfigurations with envPath and envType
            err := mergeConfigurations(envPath, envType)
            if err != nil {
                logger.Errorf("Merge failed: %v", err)
                fmt.Printf("Error merging configurations: %v\n", err)
                os.Exit(1)
            }
            fmt.Println("Configuration files merged successfully.")
        },
    }

    // Define command-line flags
    cmd.Flags().StringVar(&envType, "env-type", "dev", "Environment type (e.g., dev, prod)")

    return cmd
}


// listVersionsCmd lists the last 5 versions of Terraform and Terragrunt
func listVersionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "List the last 5 versions of Terraform and Terragrunt",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Fetching the last 5 versions for Terraform and Terragrunt...")

			tfVersions, err := getLastFiveTerraformVersions()
			if err != nil {
				logger.Errorf("Failed to get Terraform versions: %v", err)
				fmt.Printf("Error getting Terraform versions: %v\n", err)
				return
			}
			fmt.Println("Terraform Versions:")
			for _, ver := range tfVersions {
				fmt.Println(" -", ver)
			}

			tgVersions, err := getLastFiveTerragruntVersions()
			if err != nil {
				logger.Errorf("Failed to get Terragrunt versions: %v", err)
				fmt.Printf("Error getting Terragrunt versions: %v\n", err)
				return
			}
			fmt.Println("Terragrunt Versions:")
			for _, ver := range tgVersions {
				fmt.Println(" -", ver)
			}
		},
	}
	return cmd
}

func getLastFiveTerraformVersions() ([]string, error) {
	indexURL := "https://releases.hashicorp.com/terraform/index.json"

	resp, err := http.Get(indexURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Terraform release index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Terraform release index: status code %d", resp.StatusCode)
	}

	var index map[string]TerraformRelease
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode Terraform release index: %w", err)
	}

	versions := make([]*version.Version, 0, len(index))
	for verStr := range index {
		cleanVerStr := strings.TrimPrefix(verStr, "v")
		ver, err := version.NewVersion(cleanVerStr)
		if err != nil {
			logger.Warnf("invalid Terraform version format: %s", verStr)
			continue
		}
		versions = append(versions, ver)
	}

	sort.Sort(sort.Reverse(version.Collection(versions)))

	latestVersions := []string{}
	for i := 0; i < 5 && i < len(versions); i++ {
		latestVersions = append(latestVersions, versions[i].Original())
	}
	return latestVersions, nil
}

func getLastFiveTerragruntVersions() ([]string, error) {
	apiURL := "https://api.github.com/repos/gruntwork-io/terragrunt/releases"

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Terragrunt releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Terragrunt releases: status code %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode Terragrunt releases: %w", err)
	}

	versions := make([]*version.Version, 0)
	for _, release := range releases {
		if release.Prerelease {
			continue
		}
		cleanVerStr := strings.TrimPrefix(release.TagName, "v")
		ver, err := version.NewVersion(cleanVerStr)
		if err != nil {
			logger.Warnf("invalid Terragrunt version format: %s", release.TagName)
			continue
		}
		versions = append(versions, ver)
	}

	sort.Sort(sort.Reverse(version.Collection(versions)))

	latestVersions := []string{}
	for i := 0; i < 5 && i < len(versions); i++ {
		latestVersions = append(latestVersions, versions[i].Original())
	}
	return latestVersions, nil
}
// switchCmd switches to a different Terraform environment by name or full path
func switchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <new-env-name|full-path|previous>",
		Short: "Switch to a different Terraform environment by name or full path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get envDir from the persistent flags via Viper
			envDir := viper.GetString("env-dir")
			if envDir == "" {
				logger.Error("TFVENV_PATH environment variable is not set") // Lowercase and use logger
				fmt.Println("Error: TFVENV_PATH environment variable is not set.")
				os.Exit(1)
			}

			newEnvInput := args[0]
			var newEnvPath string

			// Handle switching back to the previous environment
			if newEnvInput == "previous" {
				prevEnv := os.Getenv("TFVENV_PREV")
				if prevEnv == "" {
					fmt.Println("No previous environment found to switch back to.")
					logger.Warn("no previous environment found to switch back to") // Log warning
					os.Exit(1)
				}
				newEnvInput = prevEnv
			}

			// Determine if newEnvInput is an absolute path
			if filepath.IsAbs(newEnvInput) {
				newEnvPath = newEnvInput
			} else {
				// Assume it's an environment name and search standard directories
				searchPaths := []string{filepath.Join(envDir)}
				found := false
				for _, basePath := range searchPaths {
					potentialPath := filepath.Join(basePath, newEnvInput)
					activateScriptPath := filepath.Join(potentialPath, "bin", "activate.sh")
					deactivateScriptPath := filepath.Join(potentialPath, "bin", "deactivate.sh")

					if fileExists(activateScriptPath) && fileExists(deactivateScriptPath) {
						newEnvPath = potentialPath
						found = true
						break
					}
				}

				if !found {
					fmt.Printf("Environment '%s' not found in %s.\n", newEnvInput, envDir)
					logger.Warnf("environment '%s' not found in %s", newEnvInput, envDir) // Log warning
					os.Exit(1)
				}
			}

			// Validate that the path has activate/deactivate scripts
			activateScriptPath := filepath.Join(newEnvPath, "bin", "activate.sh")
			deactivateScriptPath := filepath.Join(newEnvPath, "bin", "deactivate.sh")

			if !fileExists(activateScriptPath) || !fileExists(deactivateScriptPath) {
				fmt.Printf("The environment at '%s' does not contain required activate and deactivate scripts.\n", newEnvPath)
				logger.Warnf("missing activate/deactivate scripts at '%s'", newEnvPath) // Log warning
				os.Exit(1)
			}

			// Determine the current environment based on envDir
			currentEnvPath := envDir
			currentDeactivateScriptPath := filepath.Join(currentEnvPath, "bin", "deactivate.sh")

			// Verify that the deactivate script for the current environment exists
			if !fileExists(currentDeactivateScriptPath) {
				fmt.Printf("Deactivate script for the current environment does not exist at %s.\n", currentDeactivateScriptPath)
				logger.Warnf("deactivate script not found at '%s'", currentDeactivateScriptPath) // Log warning
				os.Exit(1)
			}

			// Track the previously active environment
			err := os.Setenv("TFVENV_PREV", filepath.Base(envDir))
			if err != nil {
				logger.Errorf("error setting TFVENV_PREV: %v", err) // Lowercase and use logger
				fmt.Printf("Error setting TFVENV_PREV: %v\n", err)
				os.Exit(1)
			}

			// Print instructions to switch environments
			fmt.Printf("Switched to environment at '%s' successfully.\n", newEnvPath)
			logger.Infof("switched to environment at '%s'", newEnvPath) // Log success
			fmt.Println("To switch environments in your current shell, run:")
			fmt.Printf("  source %s\n", currentDeactivateScriptPath)
			fmt.Printf("  source %s\n", activateScriptPath)
		},
	}

	return cmd
}
// terraformInstallCmd handles the installation of Terraform with detailed logging.
func terraformInstallCmd() *cobra.Command {
	var tfVersion string

	cmd := &cobra.Command{
		Use:   "install-terraform",
		Short: "Install Terraform",
		Run: func(cmd *cobra.Command, args []string) {
			envDir := viper.GetString("env-dir")
			if envDir == "" {
				logger.Error("TFVENV_PATH environment variable is not set")
				fmt.Println("Error: TFVENV_PATH environment variable is not set.")
				os.Exit(1)
			}

			binDir := filepath.Join(envDir, "bin") // Set binDir based on envDir

			// Ensure the bin directory exists
			if err := os.MkdirAll(binDir, 0755); err != nil {
				logger.Errorf("Failed to create bin directory: %v", err)
				fmt.Printf("Error creating bin directory: %v\n", err)
				os.Exit(1)
			}

			// Use specified version or default to "latest"
			if tfVersion == "" {
				tfVersion = "latest"
			}

			// Download and install Terraform into binDir
			err := downloadAndInstallBinary(terraformDownloadURL, tfVersion, binDir, "terraform")
			if err != nil {
				logger.Errorf("Terraform installation failed: %v", err)
				fmt.Printf("Error installing Terraform: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Terraform version %s installed successfully.\n", tfVersion)
			logger.Infof("Terraform version %s installed successfully at %s", tfVersion, filepath.Join(binDir, "terraform"))
		},
	}

	// Add a flag for the Terraform version
	cmd.Flags().StringVar(&tfVersion, "version", "", "Specify the Terraform version to install (default is latest)")

	return cmd
}
// terragruntInstallCmd handles the installation of Terragrunt with detailed logging.
func terragruntInstallCmd() *cobra.Command {
	var tgVersion string

	cmd := &cobra.Command{
		Use:   "install-terragrunt",
		Short: "Install Terragrunt",
		Run: func(cmd *cobra.Command, args []string) {
			envDir := viper.GetString("env-dir")
			if envDir == "" {
				logger.Error("TFVENV_PATH environment variable is not set")
				fmt.Println("Error: TFVENV_PATH environment variable is not set.")
				os.Exit(1)
			}

			binDir := filepath.Join(envDir, "bin") // Set binDir based on envDir

			// Ensure the bin directory exists
			if err := os.MkdirAll(binDir, 0755); err != nil {
				logger.Errorf("Failed to create bin directory: %v", err)
				fmt.Printf("Error creating bin directory: %v\n", err)
				os.Exit(1)
			}

			// Use specified version or default to "latest"
			if tgVersion == "" {
				tgVersion = "latest"
			}

			// Download and install Terragrunt into binDir
			err := downloadAndInstallBinary(terragruntDownloadURL, tgVersion, binDir, "terragrunt")
			if err != nil {
				logger.Errorf("Terragrunt installation failed: %v", err)
				fmt.Printf("Error installing Terragrunt: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Terragrunt version %s installed successfully.\n", tgVersion)
			logger.Infof("Terragrunt version %s installed successfully at %s", tgVersion, filepath.Join(binDir, "terragrunt"))
		},
	}

	// Add a flag for the Terragrunt version
	cmd.Flags().StringVar(&tgVersion, "version", "", "Specify the Terragrunt version to install (default is latest)")

	return cmd
}


// Helper function to escape values for Bash and Zsh
func escapeBash(value string) string {
	// Enclose the string in single quotes and escape existing single quotes
	// Replace ' with '\'' to properly escape single quotes within single-quoted strings
	escaped := strings.ReplaceAll(value, `'`, `'\''`)
	return fmt.Sprintf("'%s'", escaped)
}

// Helper function to escape values for Fish shell
func escapeFish(value string) string {
	// Escape backslashes and double quotes
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return fmt.Sprintf("\"%s\"", escaped)
}

// Helper function to escape values for PowerShell
func escapePowerShell(value string) string {
	// Escape backticks and double quotes
	escaped := strings.ReplaceAll(value, "`", "``")
	escaped = strings.ReplaceAll(escaped, `"`, "`\"")
	return escaped
}
// generateActivateScript generates activation scripts for multiple shells
func generateActivateScript(envDir, envName string, config Config) error {
	binDir := filepath.Join(envDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Generate for Bash and Zsh
	activateShPath := filepath.Join(binDir, "activate.sh")
	var bufferBash bytes.Buffer
	bufferBash.WriteString("#!/bin/bash\n")
	bufferBash.WriteString("# Activate script for tfvenv\n\n")

	// Save original PATH if not already saved
	bufferBash.WriteString("if [ -z \"$INITIAL_PATH\" ]; then\n")
	bufferBash.WriteString("  export INITIAL_PATH=\"$PATH\"\n")
	bufferBash.WriteString("fi\n\n")

	// Set TFVENV_PATH if not already set
	bufferBash.WriteString("if [ -z \"$TFVENV_PATH\" ]; then\n")
	bufferBash.WriteString(fmt.Sprintf("  export TFVENV_PATH=\"%s\"\n", escapeBash(envDir)))
	bufferBash.WriteString("fi\n\n")

	// Prepend bin directory to PATH if not already in PATH
	bufferBash.WriteString("if [[ \":$PATH:\" != *\":$TFVENV_PATH/bin:\"* ]]; then\n")
	bufferBash.WriteString("  export PATH=\"$TFVENV_PATH/bin:$PATH\"\n")
	bufferBash.WriteString("fi\n\n")

	// Set TFVENV_ENV to the environment name
	bufferBash.WriteString(fmt.Sprintf("export TFVENV_ENV=\"%s\"\n", escapeBash(envName)))

	// Set additional environment variables, avoiding duplicates
	for key, value := range config.EnvVars {
		bufferBash.WriteString(fmt.Sprintf("export %s=%s\n", key, escapeBash(value)))
	}
	bufferBash.WriteString("\n")

	// Save original PS1 if not already saved
	bufferBash.WriteString("if [ -z \"$INITIAL_PS1\" ]; then\n")
	bufferBash.WriteString("  export INITIAL_PS1=\"$PS1\"\n")
	bufferBash.WriteString("fi\n\n")

	// Prepend environment name to PS1 if not already done
	bufferBash.WriteString(fmt.Sprintf("if [[ \"$PS1\" != \"(%s) *\" ]]; then\n", escapeBash(envName)))
	bufferBash.WriteString(fmt.Sprintf("  export PS1=\"(%s) $PS1\"\n", escapeBash(envName)))
	bufferBash.WriteString("fi\n\n")

	// Inform the user to source the script
	bufferBash.WriteString("# To deactivate, run 'source deactivate.sh'\n")

	// Optionally, add diagnostic messages
	bufferBash.WriteString(fmt.Sprintf("echo \"Environment '%s' activated.\"\n", escapeBash(envName)))
	bufferBash.WriteString("echo \"TF_PLUGIN_CACHE_DIR is set to $TF_PLUGIN_CACHE_DIR\"\n")
	bufferBash.WriteString("echo \"TF_DATA_DIR is set to $TF_DATA_DIR\"\n")

	// Write the Bash/Zsh activate script
	err := os.WriteFile(activateShPath, bufferBash.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write activate.sh: %w", err)
	}
	logger.Infof("Bash/Zsh activate script generated at %s", activateShPath)

	// Generate for Fish
	activateFishPath := filepath.Join(binDir, "activate.fish")
	var bufferFish bytes.Buffer
	bufferFish.WriteString("#!/usr/bin/env fish\n")
	bufferFish.WriteString("# Activate script for tfvenv\n\n")

	// Save original PATH if not already saved
	bufferFish.WriteString("if not set -q INITIAL_PATH\n")
	bufferFish.WriteString("  set -gx INITIAL_PATH $PATH\n")
	bufferFish.WriteString("end\n\n")

	// Set TFVENV_PATH if not already set
	bufferFish.WriteString("if not set -q TFVENV_PATH\n")
	bufferFish.WriteString(fmt.Sprintf("  set -gx TFVENV_PATH %s\n", escapeFish(envDir)))
	bufferFish.WriteString("end\n\n")

	// Prepend bin directory to PATH if not already in PATH
	bufferFish.WriteString("if not contains \"$TFVENV_PATH/bin\" $PATH\n")
	bufferFish.WriteString("  set -gx PATH \"$TFVENV_PATH/bin\" $PATH\n")
	bufferFish.WriteString("end\n\n")

	// Set TFVENV_ENV to the environment name
	bufferFish.WriteString(fmt.Sprintf("set -gx TFVENV_ENV \"%s\"\n", escapeFish(envName)))

	// Set additional environment variables, avoiding duplicates
	for key, value := range config.EnvVars {
		bufferFish.WriteString(fmt.Sprintf("set -gx %s %s\n", key, escapeFish(value)))
	}
	bufferFish.WriteString("\n")

	// Save original fish_prompt if not already saved
	bufferFish.WriteString("if not set -q INITIAL_PROMPT\n")
	bufferFish.WriteString("  set -gx INITIAL_PROMPT (functions fish_prompt | tail -n +2)\n")
	bufferFish.WriteString("end\n\n")

	// Prepend environment name to prompt if not already done
	bufferFish.WriteString(fmt.Sprintf("if not string match -qr \"^\\(%s\\) \" (functions -c fish_prompt)\n", escapeFish(envName)))
	bufferFish.WriteString(fmt.Sprintf("  function fish_prompt\n        echo \"(%s) \" (functions fish_prompt | tail -n +2)\n  end\n", escapeFish(envName)))
	bufferFish.WriteString("end\n\n")

	// Inform the user to source the script
	bufferFish.WriteString("# To deactivate, run 'source deactivate.fish'\n")

	// Optionally, add diagnostic messages
	bufferFish.WriteString(fmt.Sprintf("echo \"Environment '%s' activated.\"\n", escapeFish(envName)))
	bufferFish.WriteString("echo \"TF_PLUGIN_CACHE_DIR is set to $TF_PLUGIN_CACHE_DIR\"\n")
	bufferFish.WriteString("echo \"TF_DATA_DIR is set to $TF_DATA_DIR\"\n")

	// Write the Fish activate script
	err = os.WriteFile(activateFishPath, bufferFish.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write activate.fish: %w", err)
	}
	logger.Infof("Fish activate script generated at %s", activateFishPath)

	// Generate for PowerShell
	activatePs1Path := filepath.Join(binDir, "Activate.ps1")
	var bufferPs1 bytes.Buffer
	bufferPs1.WriteString("# Activate script for tfvenv\n\n")

	// Save original PATH if not already saved
	bufferPs1.WriteString("if (-not (Test-Path Env:INITIAL_PATH)) {\n")
	bufferPs1.WriteString("    $env:INITIAL_PATH = $env:PATH\n")
	bufferPs1.WriteString("}\n\n")

	// Set TFVENV_PATH if not already set
	bufferPs1.WriteString("if (-not (Test-Path Env:TFVENV_PATH)) {\n")
	bufferPs1.WriteString(fmt.Sprintf("    $env:TFVENV_PATH = \"%s\"\n", escapePowerShell(envDir)))
	bufferPs1.WriteString("}\n\n")

	// Prepend bin directory to PATH if not already in PATH
	bufferPs1.WriteString("$tfvenvBin = Join-Path $env:TFVENV_PATH 'bin'\n")
	bufferPs1.WriteString("if (-not ($env:PATH -split ';' | Where-Object { $_ -ieq $tfvenvBin })) {\n")
	bufferPs1.WriteString("    $env:PATH = \"$tfvenvBin;$env:PATH\"\n")
	bufferPs1.WriteString("}\n\n")

	// Set TFVENV_ENV to the environment name
	bufferPs1.WriteString(fmt.Sprintf("$env:TFVENV_ENV = \"%s\"\n\n", escapePowerShell(envName)))

	// Set additional environment variables, avoiding duplicates
	for key, value := range config.EnvVars {
		bufferPs1.WriteString(fmt.Sprintf("$env:%s = \"%s\"\n", key, escapePowerShell(value)))
	}
	bufferPs1.WriteString("\n")

	// Save original prompt if not already saved
	bufferPs1.WriteString("if (-not (Test-Path Env:INITIAL_PROMPT)) {\n")
	bufferPs1.WriteString("    $env:INITIAL_PROMPT = $function:prompt\n")
	bufferPs1.WriteString("}\n\n")

	// Prepend environment name to prompt if not already done
	bufferPs1.WriteString(fmt.Sprintf("if (-not ($function:prompt -contains \"(%s)\")) {\n", escapePowerShell(envName)))
	bufferPs1.WriteString(fmt.Sprintf("    function prompt { \"(%s) \" + (& $env:INITIAL_PROMPT) }\n", escapePowerShell(envName)))
	bufferPs1.WriteString("}\n\n")

	// Inform the user to source the script
	bufferPs1.WriteString("# To deactivate, run '.\\Deactivate.ps1'\n\n")

	// Optionally, add diagnostic messages
	bufferPs1.WriteString(fmt.Sprintf("Write-Output \"Environment '%s' activated.\"\n", escapePowerShell(envName)))
	bufferPs1.WriteString("Write-Output \"TF_PLUGIN_CACHE_DIR is set to $env:TF_PLUGIN_CACHE_DIR\"\n")
	bufferPs1.WriteString("Write-Output \"TF_DATA_DIR is set to $env:TF_DATA_DIR\"\n")

	// Write the PowerShell activate script
	err = os.WriteFile(activatePs1Path, bufferPs1.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write Activate.ps1: %w", err)
	}
	logger.Infof("PowerShell activate script generated at %s", activatePs1Path)

	return nil
}
// generateDeactivateScript generates deactivation scripts for multiple shells
func generateDeactivateScript(envDir string, config Config) error {
	binDir := filepath.Join(envDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Generate for Bash and Zsh
	deactivateShPath := filepath.Join(binDir, "deactivate.sh")
	var bufferBash bytes.Buffer
	bufferBash.WriteString("#!/bin/bash\n")
	bufferBash.WriteString("# Deactivate script for tfvenv\n\n")

	// Restore original PATH
	bufferBash.WriteString("if [ -n \"$INITIAL_PATH\" ]; then\n")
	bufferBash.WriteString("  export PATH=\"$INITIAL_PATH\"\n")
	bufferBash.WriteString("  unset INITIAL_PATH\n")
	bufferBash.WriteString("fi\n\n")

	// Unset TFVENV_PATH
	bufferBash.WriteString("unset TFVENV_PATH\n")

	// Unset TFVENV_ENV
	bufferBash.WriteString("unset TFVENV_ENV\n\n")

	// Unset additional environment variables
	for key := range config.EnvVars {
		bufferBash.WriteString(fmt.Sprintf("unset %s\n", key))
	}
	bufferBash.WriteString("\n")

	// Restore original PS1
	bufferBash.WriteString("if [ -n \"$INITIAL_PS1\" ]; then\n")
	bufferBash.WriteString("  export PS1=\"$INITIAL_PS1\"\n")
	bufferBash.WriteString("  unset INITIAL_PS1\n")
	bufferBash.WriteString("fi\n\n")

	// Inform the user
	bufferBash.WriteString("# Environment deactivated.\n")

	// Write the Bash/Zsh deactivate script
	err := os.WriteFile(deactivateShPath, bufferBash.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write deactivate.sh: %w", err)
	}
	logger.Infof("Bash/Zsh deactivate script generated at %s", deactivateShPath)

	// Generate for Fish
	deactivateFishPath := filepath.Join(binDir, "deactivate.fish")
	var bufferFish bytes.Buffer
	bufferFish.WriteString("#!/usr/bin/env fish\n")
	bufferFish.WriteString("# Deactivate script for tfvenv\n\n")

	// Restore original PATH
	bufferFish.WriteString("if set -q INITIAL_PATH\n")
	bufferFish.WriteString("    set -gx PATH $INITIAL_PATH\n")
	bufferFish.WriteString("    set -e INITIAL_PATH\n")
	bufferFish.WriteString("end\n\n")

	// Unset TFVENV_PATH
	bufferFish.WriteString("set -e TFVENV_PATH\n")

	// Unset TFVENV_ENV
	bufferFish.WriteString("set -e TFVENV_ENV\n\n")

	// Unset additional environment variables
	for key := range config.EnvVars {
		bufferFish.WriteString(fmt.Sprintf("set -e %s\n", key))
	}
	bufferFish.WriteString("\n")

	// Restore original prompt
	bufferFish.WriteString("if set -q INITIAL_PROMPT\n")
	bufferFish.WriteString("    functions --erase fish_prompt\n")
	bufferFish.WriteString("    functions fish_prompt\n")
	bufferFish.WriteString("        $INITIAL_PROMPT\n")
	bufferFish.WriteString("    end\n")
	bufferFish.WriteString("    set -e INITIAL_PROMPT\n")
	bufferFish.WriteString("end\n\n")

	// Inform the user
	bufferFish.WriteString("# Environment deactivated.\n")

	// Write the Fish deactivate script
	err = os.WriteFile(deactivateFishPath, bufferFish.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write deactivate.fish: %w", err)
	}
	logger.Infof("Fish deactivate script generated at %s", deactivateFishPath)

	// Generate for PowerShell
	deactivatePs1Path := filepath.Join(binDir, "Deactivate.ps1")
	var bufferPs1 bytes.Buffer
	bufferPs1.WriteString("# Deactivate script for tfvenv\n\n")

	// Restore original PATH
	bufferPs1.WriteString("if ($env:INITIAL_PATH) {\n")
	bufferPs1.WriteString("    $env:PATH = $env:INITIAL_PATH\n")
	bufferPs1.WriteString("    Remove-Item Env:INITIAL_PATH\n")
	bufferPs1.WriteString("}\n\n")

	// Unset TFVENV_PATH
	bufferPs1.WriteString("Remove-Item Env:TFVENV_PATH\n")

	// Unset TFVENV_ENV
	bufferPs1.WriteString("Remove-Item Env:TFVENV_ENV\n\n")

	// Unset additional environment variables
	for key := range config.EnvVars {
		bufferPs1.WriteString(fmt.Sprintf("Remove-Item Env:%s\n", key))
	}
	bufferPs1.WriteString("\n")

	// Restore original prompt
	bufferPs1.WriteString("if ($env:INITIAL_PROMPT) {\n")
	bufferPs1.WriteString("    function prompt { & $env:INITIAL_PROMPT }\n")
	bufferPs1.WriteString("    Remove-Item Env:INITIAL_PROMPT\n")
	bufferPs1.WriteString("}\n\n")

	// Inform the user
	bufferPs1.WriteString("# Environment deactivated.\n")

	// Write the PowerShell deactivate script
	err = os.WriteFile(deactivatePs1Path, bufferPs1.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write Deactivate.ps1: %w", err)
	}
	logger.Infof("PowerShell deactivate script generated at %s", deactivatePs1Path)

	return nil
}
