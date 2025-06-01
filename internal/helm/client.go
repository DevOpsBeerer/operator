package helm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client provides helm operations
type Client struct {
	workDir string
}

// NewClient creates a new helm client
func NewClient() (*Client, error) {
	// Create work directory for git clones
	workDir := "/tmp/devopsbeerer-helm"
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Check if helm is available
	if _, err := exec.LookPath("helm"); err != nil {
		return nil, fmt.Errorf("helm not found in PATH: %w", err)
	}

	return &Client{
		workDir: workDir,
	}, nil
}

// Install installs a helm chart
func (c *Client) Install(ctx context.Context, releaseName, namespace, repoURL, chartPath, values string) error {
	// Clone or update the git repository
	repoPath, err := c.cloneOrUpdateRepo(ctx, repoURL)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Construct full chart path
	fullChartPath := filepath.Join(repoPath, chartPath)

	// Build helm install command
	args := []string{
		"upgrade", "--install",
		releaseName,
		fullChartPath,
		"--namespace", namespace,
		"--create-namespace",
		"--wait",
		"--timeout", "10m",
	}

	// Add values if provided
	if values != "" {
		// Write values to temp file
		valuesFile := filepath.Join(c.workDir, fmt.Sprintf("values-%s.yaml", releaseName))
		if err := os.WriteFile(valuesFile, []byte(values), 0644); err != nil {
			return fmt.Errorf("failed to write values file: %w", err)
		}
		defer os.Remove(valuesFile)

		args = append(args, "--values", valuesFile)
	}

	// Execute helm install
	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm install failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Uninstall removes a helm release
func (c *Client) Uninstall(ctx context.Context, releaseName, namespace string) error {
	args := []string{
		"uninstall",
		releaseName,
		"--namespace", namespace,
		"--wait",
		"--timeout", "5m",
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if release doesn't exist
		if strings.Contains(string(output), "not found") {
			return nil // Already uninstalled
		}
		return fmt.Errorf("helm uninstall failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Status checks the status of a helm release
func (c *Client) Status(ctx context.Context, releaseName, namespace string) (string, error) {
	args := []string{
		"status",
		releaseName,
		"--namespace", namespace,
		"--output", "json",
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("helm status failed: %w", err)
	}

	return string(output), nil
}

// cloneOrUpdateRepo clones or updates a git repository
func (c *Client) cloneOrUpdateRepo(ctx context.Context, repoURL string) (string, error) {
	// Generate repo directory name from URL
	repoName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	repoPath := filepath.Join(c.workDir, repoName)

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		// Repository exists, update it
		cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull", "--rebase")
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
		}
	} else {
		// Clone the repository
		cmd := exec.CommandContext(ctx, "git", "clone", repoURL, repoPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
		}
	}

	return repoPath, nil
}

// Cleanup removes temporary files
func (c *Client) Cleanup() error {
	return os.RemoveAll(c.workDir)
}
