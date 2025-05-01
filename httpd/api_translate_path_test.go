package httpd_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/drakkan/sftpgo/dataprovider"
	"github.com/drakkan/sftpgo/httpd/translate"
	"github.com/drakkan/sftpgo/httpdtest"
	"github.com/drakkan/sftpgo/vfs"
)

func TestTranslatePath(t *testing.T) {
	basicUser := getTestUser()
	basicUser.FsConfig.Provider = dataprovider.GCSFilesystemProvider
	basicUser.FsConfig.GCSConfig = vfs.GCSFsConfig{
		Bucket:               `bucket1`,
		KeyPrefix:            `users/test1`,
		AutomaticCredentials: 1,
	}

	user, _, err := httpdtest.AddUser(basicUser, http.StatusCreated)
	assert.NoError(t, err)

	// successful response
	translated, err := httpdtest.TranslatePath(translate.Request{
		Username: defaultUsername,
		Password: defaultPassword,
		FilePath: `test.txt`,
	}, http.StatusOK)

	assert.NoError(t, err)
	assert.Equal(t, translate.Response{
		Provider: dataprovider.GCSFilesystemProvider.String(),
		Bucket:   `bucket1`,
		Key:      `/users/test1/test.txt`,
	}, translated)

	// validation exception
	_, err = httpdtest.TranslatePath(translate.Request{
		Username: defaultUsername,
		Password: defaultPassword,
	}, http.StatusBadRequest)

	assert.Equal(t, httpdtest.APIError{Err: `filepath is required`}, err)

	// user not found
	_, err = httpdtest.TranslatePath(translate.Request{
		Username: defaultUsername + `1`,
		Password: defaultPassword,
		FilePath: `test.txt`,
	}, http.StatusNotFound)

	assert.Error(t, err)

	// invalid credentials
	_, err = httpdtest.TranslatePath(translate.Request{
		Username: defaultUsername,
		Password: defaultPassword + `1`,
		FilePath: `test.txt`,
	}, http.StatusForbidden)

	assert.Equal(t, httpdtest.APIError{
		Message: `Access Denied`,
		Err:     `invalid credentials`,
	}, err)

	// invalid traversal error
	_, err = httpdtest.TranslatePath(translate.Request{
		Username: defaultUsername,
		Password: defaultPassword,
		FilePath: `/../user/test.txt`,
	}, http.StatusBadRequest)

	assert.Equal(t, httpdtest.APIError{
		Err: `filepath is invalid`,
	}, err)

	_, err = httpdtest.RemoveUser(user, http.StatusOK)
	assert.NoError(t, err)
}
