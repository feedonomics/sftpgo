package vfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/option"
)

const (
	testBaseURL = `http://localhost/gcs-test`
)

var (
	apr28 = time.Date(2025, 04, 28, 15, 04, 05, 00, time.UTC)
	jan1  = time.Date(2025, 01, 01, 12, 00, 00, 00, time.UTC)
)

type GCSFsSuite struct {
	suite.Suite
	Fs *GCSFs
}

func (Suite *GCSFsSuite) SetupSuite() {
	gcsClient, err := storage.NewClient(
		context.Background(),
		option.WithHTTPClient(&http.Client{}),
		option.WithEndpoint(testBaseURL),
	)
	if err != nil {
		Suite.FailNowf(`storage.NewClient`, `failed to setup storage client: %s`, err)
	}

	Suite.Fs = &GCSFs{
		svc: gcsClient,
		config: &GCSFsConfig{
			KeyPrefix: `users/test1/`,
			Bucket:    `bucket1`,
		},
		localTempDir:   os.TempDir(),
		ctxTimeout:     30 * time.Second,
		ctxLongTimeout: 300 * time.Second,
	}
}

func (Suite *GCSFsSuite) TestStat_NoCustomTime() {
	defer gock.Off()

	gock.New(testBaseURL).
		Get("/b/bucket1/o/users/test1/test.txt").
		Reply(200).
		JSON(map[string]any{
			"bucket":      "bucket1",
			"name":        "test.txt",
			"contentType": "text/plain",
			"size":        "8",
			"updated":     apr28.Format(time.RFC3339),
			//"customTime":  "",
		})

	info, err := Suite.Fs.Stat("users/test1/test.txt")
	Suite.NoError(err)
	Suite.Equal("test.txt", info.Name())
	Suite.Equal(int64(8), info.Size())
	Suite.Equal(apr28, info.ModTime())
	Suite.False(info.IsDir())

	Suite.True(gock.IsDone(), "pending mocks: %s", printPendingMocks())
}

func (Suite *GCSFsSuite) TestStat_ZeroCustomTime() {
	defer gock.Off()

	gock.New(testBaseURL).
		Get("/b/bucket1/o/users/test1/test.txt").
		Reply(200).
		JSON(map[string]any{
			"bucket":      "bucket1",
			"name":        "test.txt",
			"contentType": "text/plain",
			"size":        "8",
			"updated":     apr28.Format(time.RFC3339),
			"customTime":  "0001-01-01T00:00:00Z",
		})

	info, err := Suite.Fs.Stat("users/test1/test.txt")
	Suite.NoError(err)
	Suite.Equal("test.txt", info.Name())
	Suite.Equal(int64(8), info.Size())
	Suite.Equal(apr28, info.ModTime())
	Suite.False(info.IsDir())

	Suite.True(gock.IsDone(), "pending mocks: %s", printPendingMocks())
}

func (Suite *GCSFsSuite) TestStat_ValidCustomTime() {
	defer gock.Off()

	gock.New(testBaseURL).
		Get("/b/bucket1/o/users/test1/test.txt").
		Reply(200).
		JSON(map[string]any{
			"bucket":      "bucket1",
			"name":        "test.txt",
			"contentType": "text/plain",
			"size":        "8",
			"updated":     apr28.Format(time.RFC3339),
			"customTime":  jan1.Format(time.RFC3339),
		})

	info, err := Suite.Fs.Stat("users/test1/test.txt")
	Suite.NoError(err)
	Suite.Equal("test.txt", info.Name())
	Suite.Equal(int64(8), info.Size())
	Suite.Equal(jan1, info.ModTime())
	Suite.False(info.IsDir())

	Suite.True(gock.IsDone(), "pending mocks: %s", printPendingMocks())
}

func (Suite *GCSFsSuite) TestReadDir_IsObject_NoCustomTime() {
	defer gock.Off()

	gock.New(testBaseURL).
		Get("/b/bucket1/o/users/test1/test.txt").
		Reply(200).
		JSON(map[string]any{
			"bucket":      "bucket1",
			"name":        "test.txt",
			"contentType": "text/plain",
			"size":        "8",
			"updated":     apr28.Format(time.RFC3339),
			//"customTime":  "",
		})

	results, err := Suite.Fs.ReadDir("users/test1/test.txt")
	Suite.NoError(err)
	Suite.Len(results, 1)

	Suite.Equal("test.txt", results[0].Name())
	Suite.Equal(int64(8), results[0].Size())
	Suite.Equal(apr28, results[0].ModTime())
	Suite.False(results[0].IsDir())

	Suite.True(gock.IsDone(), "pending mocks: %s", printPendingMocks())
}

func (Suite *GCSFsSuite) TestReadDir_IsObject_WithCustomTime() {
	defer gock.Off()

	gock.New(testBaseURL).
		Get("/b/bucket1/o/users/test1/test.txt").
		Reply(200).
		JSON(map[string]any{
			"bucket":      "bucket1",
			"name":        "test.txt",
			"contentType": "text/plain",
			"size":        "8",
			"updated":     apr28.Format(time.RFC3339),
			"customTime":  jan1.Format(time.RFC3339),
		})

	results, err := Suite.Fs.ReadDir("users/test1/test.txt")
	Suite.NoError(err)
	Suite.Len(results, 1)

	Suite.Equal("test.txt", results[0].Name())
	Suite.Equal(int64(8), results[0].Size())
	Suite.Equal(jan1, results[0].ModTime())
	Suite.False(results[0].IsDir())

	Suite.True(gock.IsDone(), "pending mocks: %s", printPendingMocks())
}

func (Suite *GCSFsSuite) TestReadDir_IsDir() {
	defer gock.Off()

	gock.New(testBaseURL).
		Get("/b/bucket1/o").
		AddMatcher(customTimeInFieldsList()).
		Reply(200).
		JSON(map[string]any{
			"kind": "storage#objects",
			"items": []map[string]any{
				{
					"bucket":      "bucket1",
					"name":        "test.txt",
					"contentType": "text/plain",
					"size":        "8",
					"updated":     apr28.Format(time.RFC3339),
					//"customTime":  jan1.Format(time.RFC3339),
				},
				{
					"bucket":      "bucket1",
					"name":        "test2.txt",
					"contentType": "text/plain",
					"size":        "16",
					"updated":     apr28.Format(time.RFC3339),
					"customTime":  jan1.Format(time.RFC3339),
				},
			},
		})

	results, err := Suite.Fs.ReadDir("users/test1/")
	Suite.NoError(err)
	Suite.Len(results, 2)

	Suite.Equal("test.txt", results[0].Name())
	Suite.Equal(int64(8), results[0].Size())
	Suite.Equal(apr28, results[0].ModTime())
	Suite.False(results[0].IsDir())

	Suite.Equal("test2.txt", results[1].Name())
	Suite.Equal(int64(16), results[1].Size())
	Suite.Equal(jan1, results[1].ModTime())
	Suite.False(results[1].IsDir())

	Suite.True(gock.IsDone(), "pending mocks: %s", printPendingMocks())
}

func (Suite *GCSFsSuite) TestCreate_CustomTimeAttribute() {
	Suite.Fs.nowFunc = func() time.Time { return jan1 }
	defer func() {
		gock.Off()
		Suite.Fs.nowFunc = time.Now
	}()

	gock.New(testBaseURL).
		Post("/upload/storage/v1/b/bucket1/o").
		AddMatcher(customTimeInUploadRequest(jan1)).
		Reply(200)

	_, writer, _, err := Suite.Fs.Create("new.txt", 0)
	Suite.NoError(err)

	go func() {
		_, err = writer.Write([]byte("hello world"))
		Suite.NoError(err)
		err = writer.Close()
		Suite.NoError(err)
	}()

	select {
	case <-writer.done:
		break
	case <-time.After(time.Second * 5):
		Suite.FailNow("timeout for writer done chan")
	}

	Suite.True(gock.IsDone(), "pending mocks: %s", printPendingMocks())
}

func TestGCSFsSuite(t *testing.T) {
	suite.Run(t, new(GCSFsSuite))
}

func customTimeInFieldsList() gock.MatchFunc {
	return func(r *http.Request, _ *gock.Request) (bool, error) {
		return strings.Contains(r.URL.Query().Get("fields"), "customTime"), nil
	}
}

func customTimeInUploadRequest(expectedTime time.Time) gock.MatchFunc {
	return func(r *http.Request, _ *gock.Request) (bool, error) {
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, r.Body); err != nil {
			return false, err
		}
		return strings.Contains(
			buf.String(),
			fmt.Sprintf("\"customTime\":\"%s\"", expectedTime.Format(time.RFC3339)),
		), nil
	}
}

func printPendingMocks() string {
	var urls []string
	for _, mock := range gock.Pending() {
		urls = append(urls, mock.Request().URLStruct.String())
	}
	return strings.Join(urls, ", ")
}
