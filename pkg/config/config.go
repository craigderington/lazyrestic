package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/craigderington/lazyrestic/pkg/types"
)

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "lazyrestic", "config.yaml")
}

// Load reads and parses the configuration file
func Load(path string) (*types.ResticConfig, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into a temporary structure to check for deprecated fields
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Check for deprecated plain-text passwords
	if repos, ok := rawConfig["repositories"].([]interface{}); ok {
		for i, repo := range repos {
			if repoMap, ok := repo.(map[interface{}]interface{}); ok {
				if _, hasPassword := repoMap["password"]; hasPassword {
					repoName := "unknown"
					if name, ok := repoMap["name"].(string); ok {
						repoName = name
					}
					return nil, fmt.Errorf("repository %d (%s) uses deprecated 'password' field\n\n"+
						"Plain-text passwords are no longer supported for security.\n"+
						"Please migrate to one of these secure methods:\n\n"+
						"1. Password file (recommended):\n"+
						"   password_file: /path/to/password-file  # File must have 0400 or 0600 permissions\n\n"+
						"2. Password command (for password managers):\n"+
						"   password_command: pass show restic/%s\n\n"+
						"To migrate your existing password:\n"+
						"  mkdir -p ~/.config/lazyrestic/passwords\n"+
						"  echo 'YOUR_PASSWORD' > ~/.config/lazyrestic/passwords/%s.txt\n"+
						"  chmod 400 ~/.config/lazyrestic/passwords/%s.txt\n"+
						"  # Then update config to use: password_file: ~/.config/lazyrestic/passwords/%s.txt",
						i, repoName, repoName, repoName, repoName, repoName)
				}
			}
		}
	}

	// Parse YAML into proper structure
	var config types.ResticConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return &config, nil
}

// ValidateConfig checks the configuration for security issues
func ValidateConfig(config *types.ResticConfig, configPath string) error {
	// Check config file permissions
	if err := validateConfigFilePermissions(configPath); err != nil {
		return fmt.Errorf("config file security check failed: %w", err)
	}

	// Validate each repository configuration
	for i, repo := range config.Repositories {
		if err := validateRepositoryConfig(&repo, i); err != nil {
			return fmt.Errorf("repository %d (%s) validation failed: %w", i, repo.Name, err)
		}
	}

	return nil
}

// validateConfigFilePermissions checks that the config file has secure permissions
func validateConfigFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat config file: %w", err)
	}

	mode := info.Mode()
	// Config file should be readable/writable only by owner (0600)
	if mode.Perm() != 0600 {
		return fmt.Errorf("config file permissions are %s, should be 0600 for security", mode.Perm())
	}

	return nil
}

// validateRepositoryConfig validates a single repository configuration
func validateRepositoryConfig(repo *types.RepositoryConfig, index int) error {
	passwordMethods := 0
	if repo.PasswordFile != "" {
		passwordMethods++
	}
	if repo.PasswordCommand != "" {
		passwordMethods++
	}

	if passwordMethods == 0 {
		return fmt.Errorf("no password method specified (password_file or password_command required)")
	}

	if passwordMethods > 1 {
		return fmt.Errorf("multiple password methods specified, use only one of: password_file or password_command")
	}

	// Validate password file
	if repo.PasswordFile != "" {
		if err := validatePasswordFile(repo.PasswordFile); err != nil {
			return fmt.Errorf("password_file validation failed: %w", err)
		}
	}

	// Validate password command
	if repo.PasswordCommand != "" {
		if err := validatePasswordCommand(repo.PasswordCommand); err != nil {
			return fmt.Errorf("password_command validation failed: %w", err)
		}
	}

	return nil
}

// validatePasswordFile checks that the password file exists and has secure permissions
func validatePasswordFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("password file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access password file: %w", err)
	}

	mode := info.Mode()
	// Password file should be readable only by owner (0400 or 0600)
	if mode.Perm() != 0400 && mode.Perm() != 0600 {
		return fmt.Errorf("password file permissions are %s, should be 0400 or 0600 for security", mode.Perm())
	}

	return nil
}

// validatePasswordCommand checks that the password command doesn't contain dangerous shell metacharacters
func validatePasswordCommand(cmd string) error {
	// Dangerous characters that could allow command injection
	dangerousChars := []string{";", "|", "&", "`", "$", "(", ")", "<", ">", "\n", "\r"}

	for _, char := range dangerousChars {
		if strings.Contains(cmd, char) {
			return fmt.Errorf("password command contains dangerous character '%s': %s", char, cmd)
		}
	}

	// Check for common dangerous commands
	dangerousCommands := []string{"rm", "del", "format", "mkfs", "dd", "shred", "wget", "curl"}
	lowerCmd := strings.ToLower(cmd)

	for _, badCmd := range dangerousCommands {
		if strings.Contains(lowerCmd, badCmd) {
			return fmt.Errorf("password command contains potentially dangerous command '%s': %s", badCmd, cmd)
		}
	}

	return nil
}

// LoadAndValidate reads and parses the configuration file with validation
func LoadAndValidate(path string) (*types.ResticConfig, error) {
	// Use default path if not specified
	if path == "" {
		path = DefaultConfigPath()
	}

	config, err := Load(path)
	if err != nil {
		return nil, err
	}

	// Validate configuration for security issues
	if err := ValidateConfig(config, path); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// LoadOrDefault loads the config file, or returns an empty config if not found
func LoadOrDefault(path string) *types.ResticConfig {
	config, err := LoadAndValidate(path)
	if err != nil {
		// Return empty config - user needs to add repos or create config file
		return &types.ResticConfig{
			Repositories: []types.RepositoryConfig{},
		}
	}
	return config
}

// Save writes the configuration to a file
func Save(config *types.ResticConfig, path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateExample creates an example configuration file
func CreateExample(path string) error {
	example := &types.ResticConfig{
		Repositories: []types.RepositoryConfig{
			{
				Name:         "local-backup",
				Path:         "/path/to/local/repo",
				PasswordFile: "~/.config/lazyrestic/passwords/local-backup.txt",
			},
			{
				Name:            "home-backup",
				Path:            "/mnt/backup/restic",
				PasswordCommand: "pass show restic/home-backup",
			},
		},
	}

	return Save(example, path)
}

// RemoveRepository removes a repository from the config by name
func RemoveRepository(config *types.ResticConfig, name string) bool {
	for i, repo := range config.Repositories {
		if repo.Name == name {
			// Remove this repository by slicing around it
			config.Repositories = append(config.Repositories[:i], config.Repositories[i+1:]...)
			return true
		}
	}
	return false
}
