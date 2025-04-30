package dataprovider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemProvider_MarshalJSON(t *testing.T) {
	for fsProvider, expectedStr := range map[FilesystemProvider]string{
		LocalFilesystemProvider:     `"local"`,
		S3FilesystemProvider:        `"s3"`,
		GCSFilesystemProvider:       `"gcs"`,
		AzureBlobFilesystemProvider: `"azure"`,
		SFTPFilesystemProvider:      `"sftp"`,
		CryptedFilesystemProvider:   `"local-encrypted"`,
		InvalidFilesystemProvider:   `""`,
	} {
		b, err := json.Marshal(fsProvider)
		assert.NoError(t, err)
		assert.Equal(t, expectedStr, string(b))
	}
}

func TestFilesystemProvider_UnmarshalJSON(t *testing.T) {
	for str, expectedFsProvider := range map[string]FilesystemProvider{
		`"local"`:           LocalFilesystemProvider,
		`"s3"`:              S3FilesystemProvider,
		`"gcs"`:             GCSFilesystemProvider,
		`"azure"`:           AzureBlobFilesystemProvider,
		`"sftp"`:            SFTPFilesystemProvider,
		`"local-encrypted"`: CryptedFilesystemProvider,
		`"invalid"`:         InvalidFilesystemProvider,
		`""`:                InvalidFilesystemProvider,
	} {
		var actual FilesystemProvider
		err := json.Unmarshal([]byte(str), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expectedFsProvider, actual)
	}
}
