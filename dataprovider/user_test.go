package dataprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemProvider_String(t *testing.T) {
	for fsProvider, expectedStr := range map[FilesystemProvider]string{
		LocalFilesystemProvider:     "local",
		S3FilesystemProvider:        "s3",
		GCSFilesystemProvider:       "gcs",
		AzureBlobFilesystemProvider: "azure",
		SFTPFilesystemProvider:      "sftp",
		CryptedFilesystemProvider:   "local-encrypted",
		FilesystemProvider(99):      "",
	} {
		assert.Equal(t, expectedStr, fsProvider.String())
	}
}
