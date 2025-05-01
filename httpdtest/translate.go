package httpdtest

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/drakkan/sftpgo/httpd/translate"
)

type APIError struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

func (err APIError) Error() string {
	if err.Err != `` {
		return err.Err
	}
	return err.Message
}

func UsersS3Translate(request translate.Request, expectedStatusCode int) (translate.Response, error) {
	return baseTranslate(request, expectedStatusCode, `/api/v2/users-s3/translate-path`)
}

func TranslatePath(request translate.Request, expectedStatusCode int) (translate.Response, error) {
	return baseTranslate(request, expectedStatusCode, `/api/v2/translate-path`)
}

func baseTranslate(request translate.Request, expectedStatusCode int, path string) (translate.Response, error) {
	var translated translate.Response
	var body []byte

	folderAsJSON, _ := json.Marshal(request)
	url := buildURLRelativeToBase(path)
	resp, err := sendHTTPRequest(http.MethodPost, url, bytes.NewBuffer(folderAsJSON), "", getDefaultToken())
	if err != nil {
		return translated, err
	}
	defer resp.Body.Close()

	body, _ = getResponseBody(resp)

	if err := checkResponse(resp.StatusCode, expectedStatusCode); err != nil {
		return translated, err
	}

	if resp.StatusCode == http.StatusOK {
		if err = json.Unmarshal(body, &translated); err != nil {
			return translated, err
		}
		return translated, nil
	}

	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return translated, err
	}
	return translated, apiErr
}
