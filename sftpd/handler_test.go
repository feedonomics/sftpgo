package sftpd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetFolderPrefix(t *testing.T) {
	c := &Connection{}
	c.SetFolderPrefix(`files`)
	assert.Equal(t, `/files`, c.folderPrefix)
	c.SetFolderPrefix(`/files/`)
	assert.Equal(t, `/files`, c.folderPrefix)
	c.SetFolderPrefix(`files/`)
	assert.Equal(t, `/files`, c.folderPrefix)
}

func TestContainsFolderPrefix(t *testing.T) {
	c := &Connection{}
	c.SetFolderPrefix(`/files`)
	assert.True(t, c.containsFolderPrefix(`/files`))
	assert.True(t, c.containsFolderPrefix(`/files/test.csv`))
	assert.True(t, c.containsFolderPrefix(`files`))
	assert.True(t, c.containsFolderPrefix(`files/test.csv`))

	assert.False(t, c.containsFolderPrefix(`/files1`))
	assert.False(t, c.containsFolderPrefix(`/files1/test.csv`))
	assert.False(t, c.containsFolderPrefix(`files1/`))
	assert.False(t, c.containsFolderPrefix(`files1/test.csv`))
}

func TestContainsFolderPrefixNoPrefix(t *testing.T) {
	c := &Connection{}
	assert.True(t, c.containsFolderPrefix(`/files`))
	assert.True(t, c.containsFolderPrefix(`/files/test.csv`))
	assert.True(t, c.containsFolderPrefix(`files`))
	assert.True(t, c.containsFolderPrefix(`files/test.csv`))

	assert.True(t, c.containsFolderPrefix(`/files1`))
	assert.True(t, c.containsFolderPrefix(`/files1/test.csv`))
	assert.True(t, c.containsFolderPrefix(`files1/`))
	assert.True(t, c.containsFolderPrefix(`files1/test.csv`))
}

func TestRemoveFolderPrefix(t *testing.T) {
	c := Connection{}
	c.SetFolderPrefix(`files`)
	path, removed := c.removeFolderPrefix(`/files`)
	assert.Equal(t, `/`, path)
	assert.True(t, removed)

	path, removed = c.removeFolderPrefix(`/files1`)
	assert.Equal(t, `/files1`, path)
	assert.False(t, removed)
}

func TestRemoveFolderPrefixNoPrefix(t *testing.T) {
	c := Connection{}
	path, removed := c.removeFolderPrefix(`/files1`)
	assert.Equal(t, `/files1`, path)
	assert.True(t, removed)
}
