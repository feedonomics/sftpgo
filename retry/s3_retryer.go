package retry

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/drakkan/sftpgo/logger"
	"math"
	"time"
)

const (
	AwsRequestErrorCode     = `RequestError`
	AwsSendRequestFailedMsg = `send request failed`

	AwsClientDisconnectedErrorCode = `ClientDisconnected`
	AwsServiceUnavailableErrorCode = `ServiceUnavailable`
)

var (
	_ request.Retryer = S3Retryer{}

	// TODO: expand as necessary during testing
	// NOTE: GetObject operations are already cleanly handled/retried by the default retryer,
	// though our custom backoff duration will still apply for connection specific errors as expected
	customRetryOperations = []string{
		`UploadPart`,
	}
)

type (
	S3Retryer struct {
		client.DefaultRetryer
	}
)

func (s S3Retryer) RetryRules(r *request.Request) (duration time.Duration) {
	defer func() {
		logger.Debug(
			"S3Retryer - RetryRules()",
			"",
			"OPERATION: %s | RETRY ATTEMPT: %d/%d | DURATION: %s",
			r.Operation.Name, r.RetryCount+1, s.NumMaxRetries, duration,
		)
	}()

	if s.NumMaxRetries == 0 {
		duration = 0
		return
	}

	if isAwsConnectionError(r.Error) {
		// 2^n, where n is retry count + 1 (since RetryRules() is called before RetryCount is incremented)
		duration = time.Second * time.Duration(math.Pow(2, float64(r.RetryCount+1)))
		return
	}

	duration = s.DefaultRetryer.RetryRules(r)
	return
}

func (s S3Retryer) ShouldRetry(r *request.Request) (retry bool) {
	defer func() {
		logger.Debug(
			"S3Retryer - ShouldRetry()",
			"",
			"OPERATION: %s | RETRY: %v | ERROR: %s",
			r.Operation.Name, retry, r.Error,
		)
	}()

	if s.NumMaxRetries == 0 {
		retry = false
		return
	}

	// do not override if already set
	// ex: context deadline / cancel will have set this to false
	if r.Retryable != nil {
		retry = *r.Retryable
		return
	}

	if isAwsConnectionError(r.Error) && isCustomRetryOperation(r.Operation.Name) {
		retry = true
		return
	}

	retry = s.DefaultRetryer.ShouldRetry(r)
	return
}

func (s S3Retryer) MaxRetries() int {
	return s.NumMaxRetries
}

func isAwsConnectionError(err error) bool {
	var awsError awserr.Error
	if !errors.As(err, &awsError) {
		return false
	}

	return awsError.Code() == AwsServiceUnavailableErrorCode ||
		awsError.Code() == AwsClientDisconnectedErrorCode ||
		(awsError.Code() == AwsRequestErrorCode && awsError.Message() == AwsSendRequestFailedMsg)
}

func isCustomRetryOperation(operationName string) bool {
	for _, customOp := range customRetryOperations {
		if operationName == customOp {
			return true
		}
	}
	return false
}
