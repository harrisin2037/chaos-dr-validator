package velero

import (
	"fmt"
	"os/exec"
)

type VeleroClient struct{}

func (c *VeleroClient) CreateBackup(name string, selector map[string]string) error {
	selectorStr := ""
	for k, v := range selector {
		selectorStr += k + "=" + v
	}
	cmd := exec.Command("velero", "backup", "create", name, "--selector", selectorStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("velero backup failed: %v, output: %s", err, output)
	}
	return nil
}

func (c *VeleroClient) CreateRestore(backupName, namespace string) error {
	cmd := exec.Command("velero", "restore", "create", "restore-"+backupName, "--from-backup", backupName, "--namespace", namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("velero restore failed: %v, output: %s", err, output)
	}
	return nil
}
