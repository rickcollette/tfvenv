package snaps

import (
	"path/filepath"
)

// Snap represents the environment information to be saved in a .snap file.
type Snap struct {
	TerraformVersion  string            `json:"terraform_version"`
	TerragruntVersion string            `json:"terragrunt_version"`
	Plugins           map[string]string `json:"plugins"`  // provider: version
	EnvVars           map[string]string `json:"env_vars"` // optional environment variables
}
// GetSnapFilePath constructs the file path for a snap within a specific environment
func GetSnapFilePath(envPath, filename string) string {
	if filename == "" {
		filename = "default"
	}

	return filepath.Join(envPath, "snaps", filename+".snap")
}
