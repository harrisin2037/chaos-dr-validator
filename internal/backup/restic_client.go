package backup

import (
	"fmt"
	"os/exec"
)

type ResticClient struct{}

func (c *ResticClient) CreateBackup(name string, selector map[string]string) error {
	// Example: Restic backup command
	cmd := exec.Command("restic", "backup", "--tag", name, "/data")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restic backup failed: %v, output: %s", err, output)
	}
	return nil
}

func (c *ResticClient) CreateRestore(backupName, namespace string) error {
	cmd := exec.Command("restic", "restore", backupName, "--target", "/restore")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restic restore failed: %v, output: %s", err, output)
	}
	return nil
}
