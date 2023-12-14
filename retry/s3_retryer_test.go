package retry

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	testUploadPartRequest *request.Request

	testMatchingError = awserr.New(
		"RequestError",
		"send request failed",
		errors.New("test"),
	)

	testNonMatchingError = awserr.New(
		"something else",
		"different message",
		errors.New("test"),
	)
)

func _before() {
	testUploadPartRequest = &request.Request{
		Operation: &request.Operation{
			Name: "UploadPart",
		},
		Error:      nil,
		RetryCount: 0,
		Retryable:  nil,
	}
}

func TestS3Retryer_MaxRetries(t *testing.T) {
	_before()

	retryer := S3Retryer{}
	assert.Equal(t, 0, retryer.MaxRetries())

	retryer.NumMaxRetries = 3
	assert.Equal(t, 3, retryer.MaxRetries())
}

func TestS3Retryer_ShouldRetry(t *testing.T) {
	_before()

	retryer := S3Retryer{}

	// 0 max retries
	assert.False(t, retryer.ShouldRetry(testUploadPartRequest))

	retryer.NumMaxRetries = 3

	// request retry flag already set
	testUploadPartRequest.Retryable = aws.Bool(true)
	assert.True(t, retryer.ShouldRetry(testUploadPartRequest))

	testUploadPartRequest.Retryable = aws.Bool(false)
	assert.False(t, retryer.ShouldRetry(testUploadPartRequest))

	testUploadPartRequest.Retryable = nil

	// upload part request, matching error
	testUploadPartRequest.Error = testMatchingError
	assert.True(t, retryer.ShouldRetry(testUploadPartRequest))

	// default retryer fallback
	testUploadPartRequest.Error = testNonMatchingError
	assert.Equal(t, retryer.DefaultRetryer.ShouldRetry(testUploadPartRequest), retryer.ShouldRetry(testUploadPartRequest))
}

func TestS3Retryer_RetryRules(t *testing.T) {
	_before()

	retryer := S3Retryer{}
	retryer.NumMaxRetries = 0
	// set min/max to nanosecond so we don't have to account for jitter delay in tests
	retryer.MinRetryDelay = time.Nanosecond
	retryer.MaxRetryDelay = time.Nanosecond

	assert.Equal(t, time.Duration(0), retryer.RetryRules(testUploadPartRequest))

	retryer.NumMaxRetries = 5

	// upload part request with matching error
	testUploadPartRequest.Error = testMatchingError
	assert.Equal(t, time.Second*2, retryer.RetryRules(testUploadPartRequest))
	testUploadPartRequest.RetryCount++
	assert.Equal(t, time.Second*4, retryer.RetryRules(testUploadPartRequest))
	testUploadPartRequest.RetryCount++
	assert.Equal(t, time.Second*8, retryer.RetryRules(testUploadPartRequest))
	testUploadPartRequest.RetryCount++
	assert.Equal(t, time.Second*16, retryer.RetryRules(testUploadPartRequest))
	testUploadPartRequest.RetryCount++
	assert.Equal(t, time.Second*32, retryer.RetryRules(testUploadPartRequest))

	// everything else, no need to test built-in sdk logic
	// just make sure our custom retryer is falling back to default retryer rules
	testUploadPartRequest.RetryCount = 0
	testUploadPartRequest.Error = testNonMatchingError
	assert.Equal(t, retryer.DefaultRetryer.RetryRules(testUploadPartRequest), retryer.RetryRules(testUploadPartRequest))
}

func TestS3Retryer_isAwsRequestError(t *testing.T) {
	_before()

	assert.False(t, isAwsConnectionError(nil))

	assert.False(t, isAwsConnectionError(testUploadPartRequest.Error))
	testUploadPartRequest.Error = testNonMatchingError
	assert.False(t, isAwsConnectionError(testUploadPartRequest.Error))
	testUploadPartRequest.Error = testMatchingError
	assert.True(t, isAwsConnectionError(testUploadPartRequest.Error))
}

func TestS3Retryer_isCustomRetryOperation(t *testing.T) {
	_before()

	for _, scenario := range []struct {
		operation string
		expected  bool
	}{
		{`UploadPart`, true},
		{`GetObject`, false},
		{`AbortMultipartUpload`, false},
		{`HeadObject`, false},
		{`ListObjectsV2`, false},
		{``, false},
	} {
		assert.Equal(t, scenario.expected, isCustomRetryOperation(scenario.operation))
	}
}
