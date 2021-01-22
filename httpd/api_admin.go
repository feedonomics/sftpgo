package httpd

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"

	"github.com/drakkan/sftpgo/dataprovider"
)

type adminPwd struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func getAdmins(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0
	order := dataprovider.OrderASC
	var err error
	if _, ok := r.URL.Query()["limit"]; ok {
		limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			err = errors.New("Invalid limit")
			sendAPIResponse(w, r, err, "", http.StatusBadRequest)
			return
		}
		if limit > 500 {
			limit = 500
		}
	}
	if _, ok := r.URL.Query()["offset"]; ok {
		offset, err = strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			err = errors.New("Invalid offset")
			sendAPIResponse(w, r, err, "", http.StatusBadRequest)
			return
		}
	}
	if _, ok := r.URL.Query()["order"]; ok {
		order = r.URL.Query().Get("order")
		if order != dataprovider.OrderASC && order != dataprovider.OrderDESC {
			err = errors.New("Invalid order")
			sendAPIResponse(w, r, err, "", http.StatusBadRequest)
			return
		}
	}

	admins, err := dataprovider.GetAdmins(limit, offset, order)
	if err != nil {
		sendAPIResponse(w, r, err, "", getRespStatus(err))
		return
	}
	render.JSON(w, r, admins)
}

func getAdminByUsername(w http.ResponseWriter, r *http.Request) {
	username := getURLParam(r, "username")
	renderAdmin(w, r, username, http.StatusOK)
}

func renderAdmin(w http.ResponseWriter, r *http.Request, username string, status int) {
	admin, err := dataprovider.AdminExists(username)
	if err != nil {
		sendAPIResponse(w, r, err, "", getRespStatus(err))
		return
	}
	admin.HideConfidentialData()
	if status != http.StatusOK {
		ctx := context.WithValue(r.Context(), render.StatusCtxKey, http.StatusCreated)
		render.JSON(w, r.WithContext(ctx), admin)
	} else {
		render.JSON(w, r, admin)
	}
}

func addAdmin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	var admin dataprovider.Admin
	err := render.DecodeJSON(r.Body, &admin)
	if err != nil {
		sendAPIResponse(w, r, err, "", http.StatusBadRequest)
		return
	}
	err = dataprovider.AddAdmin(&admin)
	if err != nil {
		sendAPIResponse(w, r, err, "", getRespStatus(err))
		return
	}
	renderAdmin(w, r, admin.Username, http.StatusCreated)
}

func updateAdmin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	username := getURLParam(r, "username")
	admin, err := dataprovider.AdminExists(username)
	if err != nil {
		sendAPIResponse(w, r, err, "", getRespStatus(err))
		return
	}

	adminID := admin.ID
	err = render.DecodeJSON(r.Body, &admin)
	if err != nil {
		sendAPIResponse(w, r, err, "", http.StatusBadRequest)
		return
	}

	claims, err := getTokenClaims(r)
	if err != nil || claims.Username == "" {
		sendAPIResponse(w, r, err, "Invalid token claims", http.StatusBadRequest)
		return
	}
	if username == claims.Username {
		if claims.isCriticalPermRemoved(admin.Permissions) {
			sendAPIResponse(w, r, errors.New("You cannot remove these permissions to yourself"), "", http.StatusBadRequest)
			return
		}
		if admin.Status == 0 {
			sendAPIResponse(w, r, errors.New("You cannot disable yourself"), "", http.StatusBadRequest)
			return
		}
	}
	admin.ID = adminID
	admin.Username = username
	if err := dataprovider.UpdateAdmin(&admin); err != nil {
		sendAPIResponse(w, r, err, "", getRespStatus(err))
		return
	}
	sendAPIResponse(w, r, nil, "Admin updated", http.StatusOK)
}

func deleteAdmin(w http.ResponseWriter, r *http.Request) {
	username := getURLParam(r, "username")
	claims, err := getTokenClaims(r)
	if err != nil || claims.Username == "" {
		sendAPIResponse(w, r, err, "Invalid token claims", http.StatusBadRequest)
		return
	}
	if username == claims.Username {
		sendAPIResponse(w, r, errors.New("You cannot delete yourself"), "", http.StatusBadRequest)
		return
	}

	err = dataprovider.DeleteAdmin(username)
	if err != nil {
		sendAPIResponse(w, r, err, "", getRespStatus(err))
		return
	}
	sendAPIResponse(w, r, err, "Admin deleted", http.StatusOK)
}

func changeAdminPassword(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	var pwd adminPwd
	err := render.DecodeJSON(r.Body, &pwd)
	if err != nil {
		sendAPIResponse(w, r, err, "", http.StatusBadRequest)
		return
	}
	err = doChangeAdminPassword(r, pwd.CurrentPassword, pwd.NewPassword, pwd.NewPassword)
	if err != nil {
		sendAPIResponse(w, r, err, "", getRespStatus(err))
		return
	}
	sendAPIResponse(w, r, err, "Password updated", http.StatusOK)
}

func doChangeAdminPassword(r *http.Request, currentPassword, newPassword, confirmNewPassword string) error {
	if currentPassword == "" || newPassword == "" || confirmNewPassword == "" {
		return dataprovider.NewValidationError("Please provide the current password and the new one two times")
	}
	if newPassword != confirmNewPassword {
		return dataprovider.NewValidationError("The two password fields do not match")
	}
	if currentPassword == newPassword {
		return dataprovider.NewValidationError("The new password must be different from the current one")
	}
	claims, err := getTokenClaims(r)
	if err != nil {
		return err
	}
	admin, err := dataprovider.AdminExists(claims.Username)
	if err != nil {
		return err
	}
	match, err := admin.CheckPassword(currentPassword)
	if !match || err != nil {
		return dataprovider.NewValidationError("Current password does not match")
	}

	admin.Password = newPassword

	return dataprovider.UpdateAdmin(&admin)
}

func getTokenClaims(r *http.Request) (jwtTokenClaims, error) {
	tokenClaims := jwtTokenClaims{}
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return tokenClaims, err
	}
	tokenClaims.Decode(claims)

	return tokenClaims, nil
}
