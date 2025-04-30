package httpd

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/drakkan/sftpgo/common"
	"github.com/drakkan/sftpgo/dataprovider"
	"github.com/drakkan/sftpgo/httpd/translate"
)

func postTranslatePath(w http.ResponseWriter, r *http.Request) {
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

	render.JSON(w, r, resp)
}

func handleTranslateRequest(r *http.Request) (translate.Response, error) {
	var req translate.Request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		return translate.Response{}, wrapAPIError(err, "", http.StatusBadRequest)
	}

	if err := req.Validate(); err != nil {
		return translate.Response{}, wrapAPIError(err, "", http.StatusBadRequest)
	}

	user, err := dataprovider.CheckUserAndPass(req.Username, req.Password, ``, common.ProtocolSSH)
	if err != nil {
		if errors.Is(err, dataprovider.ErrInvalidCredentials) {
			return translate.Response{}, wrapAPIError(err, "Access Denied", 403)
		}
		return translate.Response{}, wrapAPIError(err, "", getRespStatus(err))
	}

	resp, err := req.ResolvePath(user.FsConfig)
	if err != nil {
		return translate.Response{}, wrapAPIError(err, "", http.StatusBadRequest)
	}

	return resp, nil
}
