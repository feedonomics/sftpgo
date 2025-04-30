package httpd

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/drakkan/sftpgo/dataprovider"
	"github.com/drakkan/sftpgo/httpd/translate"
)

// Deprecated: Use generic postTranslatePath instead
// For backwards compatability this will return an error if the backend filesystem is not S3
func userS3Translate(w http.ResponseWriter, r *http.Request) {
	resp, err := handleTranslateRequest(r)
	if err != nil {
		var apiErr apiError
		if errors.As(err, &apiErr) {
			sendAPIResponse(w, r, apiErr.err, apiErr.msg, apiErr.status)
		} else {
			sendAPIResponse(w, r, err, "", http.StatusBadRequest)
		}
		return
	}

	if resp.Provider != dataprovider.S3FilesystemProvider.String() {
		sendAPIResponse(w, r, translate.ErrFileSystemNotS3, "", http.StatusBadRequest)
		return
	}

	render.JSON(w, r, resp)
}
