package httpd

var (
	_ error = apiError{}
)

type (
	apiError struct {
		err    error
		msg    string
		status int
	}
)

func wrapAPIError(err error, msg string, status int) apiError {
	return apiError{
		err:    err,
		msg:    msg,
		status: status,
	}
}

func (e apiError) Error() string {
	return e.err.Error()
}
