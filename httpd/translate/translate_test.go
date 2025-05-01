package translate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/drakkan/sftpgo/dataprovider"
	"github.com/drakkan/sftpgo/vfs"
)

func TestRequestValidation(t *testing.T) {
	Req := Request{}
	assert.Equal(t, ErrUsernameRequired, Req.Validate())
	Req.Username = `user1`
	assert.Equal(t, ErrPasswordRequired, Req.Validate())
	Req.Password = `pass1`
	assert.Equal(t, ErrFilePathRequired, Req.Validate())
	Req.FilePath = `/`
	assert.Equal(t, ErrFilePathInvalid, Req.Validate())
	Req.FilePath = `test.txt`
	assert.Nil(t, Req.Validate())
}

func TestResolvePath_S3_KeyPrefixTransversal(t *testing.T) {
	Req := Request{FilePath: `/../user/test.csv`}
	Resp, err := Req.ResolvePath(dataprovider.Filesystem{
		Provider: dataprovider.S3FilesystemProvider,
		S3Config: vfs.S3FsConfig{
			KeyPrefix: `users/user1/`,
		},
	})
	assert.Equal(t, Response{}, Resp)
	assert.Equal(t, ErrFilePathInvalid, err)
}

func TestResolvePath_S3_Success(t *testing.T) {
	Req := Request{FilePath: `test.csv`}
	Resp, err := Req.ResolvePath(dataprovider.Filesystem{
		Provider: dataprovider.S3FilesystemProvider,
		S3Config: vfs.S3FsConfig{
			Region:    `us-east-1`,
			Bucket:    `bucket1`,
			KeyPrefix: `users/user1/`,
		},
	})
	assert.Equal(t, Response{
		Provider: dataprovider.S3FilesystemProvider.String(),
		Region:   `us-east-1`,
		Bucket:   `bucket1`,
		Key:      `/users/user1/test.csv`,
	}, Resp)
	assert.Nil(t, err)
}

func TestResolvePath_GCS_KeyPrefixTransversal(t *testing.T) {
	Req := Request{FilePath: `/../../user/test.csv`}
	Resp, err := Req.ResolvePath(dataprovider.Filesystem{
		Provider: dataprovider.GCSFilesystemProvider,
		GCSConfig: vfs.GCSFsConfig{
			KeyPrefix: `users/user1/`,
		},
	})
	assert.Empty(t, Resp)
	assert.Equal(t, ErrFilePathInvalid, err)
}

func TestResolvePath_GCS_Success(t *testing.T) {
	Req := Request{FilePath: `test.csv`}
	Resp, err := Req.ResolvePath(dataprovider.Filesystem{
		Provider: dataprovider.GCSFilesystemProvider,
		GCSConfig: vfs.GCSFsConfig{
			Bucket:    `bucket1`,
			KeyPrefix: `users/user1/`,
		},
	})
	assert.Equal(t, Response{
		Provider: dataprovider.GCSFilesystemProvider.String(),
		Bucket:   `bucket1`,
		Key:      `/users/user1/test.csv`,
	}, Resp)
	assert.Nil(t, err)
}

func TestResolvePath_Unsupported(t *testing.T) {
	Req := Request{FilePath: `test.csv`}
	Resp, err := Req.ResolvePath(dataprovider.Filesystem{
		Provider: dataprovider.LocalFilesystemProvider,
	})
	assert.Empty(t, Resp)
	assert.Equal(t, ErrFileSystemNotSupported, err)
}
