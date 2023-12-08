package retry

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"math"
	"time"
)

const (
	AwsRequestErrorCode     = `RequestError`
	AwsSendRequestFailedMsg = `send request failed`
)

var (
	_ request.Retryer = S3Retryer{}
)

type (
	S3Retryer struct {
		client.DefaultRetryer
	}
)

func (s S3Retryer) RetryRules(r *request.Request) time.Duration {
	if s.NumMaxRetries == 0 {
		return 0
	}

	if isAwsRequestError(r) {
		// 2^n, where n is retry count + 1 (since RetryRules() is called before RetryCount is incremented)
		return time.Second * time.Duration(math.Pow(2, float64(r.RetryCount+1)))
	}

	return s.DefaultRetryer.RetryRules(r)
}

func (s S3Retryer) ShouldRetry(r *request.Request) bool {
	if s.NumMaxRetries == 0 {
		return false
	}

	// do not override if already set
	// ex: context deadline / cancel will have set this to false
	if r.Retryable != nil {
		return *r.Retryable
	}

	if isAwsRequestError(r) {
		return true
	}

	return s.DefaultRetryer.ShouldRetry(r)
}

func (s S3Retryer) MaxRetries() int {
	return s.NumMaxRetries
}

func isAwsRequestError(r *request.Request) bool {
	var awsError awserr.Error
	if !errors.As(r.Error, &awsError) {
		return false
	}

	return awsError.Code() == AwsRequestErrorCode && awsError.Message() == AwsSendRequestFailedMsg
}
