package backup

type BackupClient interface {
	CreateBackup(name string, selector map[string]string) error
	CreateRestore(backupName, namespace string) error
}
