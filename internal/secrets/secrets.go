package secrets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/monkeymonk/homelabctl/internal/paths"
)

// LoadSecrets loads secrets/<stack>.yaml or secrets/<stack>.enc.yaml if it exists
// Automatically decrypts .enc.yaml files using SOPS
// Returns empty map if file doesn't exist (secrets are optional)
func LoadSecrets(stackName string) (map[string]interface{}, error) {
	// Try both .enc.yaml and .yaml extensions (encrypted first)
	secretsPaths := []string{
		paths.SecretsFilePath(stackName, paths.SecretsEncExt),
		paths.SecretsFilePath(stackName, paths.SecretsExt),
	}

	var secretsFile string
	for _, path := range secretsPaths {
		if _, err := os.Stat(path); err == nil {
			secretsFile = path
			break
		}
	}

	// Secrets are optional
	if secretsFile == "" {
		return make(map[string]interface{}), nil
	}

	var data []byte
	var err error

	// Check if file needs SOPS decryption
	if strings.HasSuffix(secretsFile, paths.SecretsEncExt) {
		data, err = decryptWithSOPS(secretsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secrets for %s: %w", stackName, err)
		}
	} else {
		// Plain YAML file - read directly
		data, err = os.ReadFile(secretsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read secrets for %s: %w", stackName, err)
		}
	}

	var secrets map[string]interface{}
	if err := yaml.Unmarshal(data, &secrets); err != nil {
		return nil, fmt.Errorf("failed to parse secrets for %s: %w", stackName, err)
	}

	if secrets == nil {
		secrets = make(map[string]interface{})
	}

	return secrets, nil
}

// decryptWithSOPS uses the sops command to decrypt an encrypted file
func decryptWithSOPS(filePath string) ([]byte, error) {
	// Check if sops is available
	sopsPath, err := exec.LookPath("sops")
	if err != nil {
		return nil, fmt.Errorf("sops not found in PATH - install from https://github.com/getsops/sops\nFile: %s", filePath)
	}

	// Run: sops -d <file>
	cmd := exec.Command(sopsPath, "-d", filePath)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's an exit error with stderr
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			return nil, fmt.Errorf("sops decryption failed: %s\nFile: %s", strings.TrimSpace(stderr), filepath.Base(filePath))
		}
		return nil, fmt.Errorf("failed to run sops: %w\nFile: %s", err, filepath.Base(filePath))
	}

	return output, nil
}
