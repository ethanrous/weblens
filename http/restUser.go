package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// CreateUser godoc
//
//	@ID			CreateUser
//
//	@Security	ApiKeyAuth
//
//	@Summary	Create a new user
//	@Tags		Users
//	@Produce	json
//	@Param		newUserParams	body	rest.NewUserParams	true	"New user params"
//	@Success	201
//	@Failure	401
//	@Router		/users [post]
func createUser(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}
	if SafeErrorAndExit(err, w) {
		return
	}
	userParams, err := readCtxBody[rest.NewUserParams](w, r)
	if err != nil {
		return
	}

	if !u.Admin && (userParams.AutoActivate || userParams.Admin) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	newUser, err := models.NewUser(userParams.Username, userParams.Password, userParams.FullName, userParams.Admin, userParams.AutoActivate)
	if SafeErrorAndExit(err, w) {
		return
	}

	err = pack.FileService.CreateUserHome(newUser)
	if SafeErrorAndExit(err, w) {
		return
	}

	err = pack.UserService.Add(newUser)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// LoginUser godoc
//
//	@ID			LoginUser
//
//	@Summary	Login User
//	@Tags		Users
//	@Produce	json
//	@Param		loginParams	body		rest.LoginBody	true	"Login params"
//	@Success	200			{object}	rest.UserInfo	"Logged-in users info"
//	@Failure	401
//	@Router		/users/auth [post]
func loginUser(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	userCredentials, err := readCtxBody[rest.LoginBody](w, r)
	if err != nil {
		return
	}

	u := pack.UserService.Get(userCredentials.Username)
	if u == nil {
		SafeErrorAndExit(werror.ErrNoUserLogin, w)
		return
	}

	if !u.Activated {
		pack.Log.Warn().Msgf("[%s] attempted login but is not activated", u.Username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if u.CheckLogin(userCredentials.Password) {
		pack.Log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Valid login for [%s]", userCredentials.Username) })

		var token string
		var expires time.Time
		token, expires, err = pack.AccessService.GenerateJwtToken(u)
		if err != nil || token == "" {
			pack.Log.Error().Stack().Err(werror.Errorf("Could not get login token")).Msg("")
			w.WriteHeader(http.StatusInternalServerError)
		}

		userInfo := rest.UserToUserInfo(u)

		cookie := fmt.Sprintf("%s=%s;Path=/;Expires=%s;HttpOnly", SessionTokenCookie, token, expires.Format(time.RFC1123))

		pack.Log.Trace().Msgf("Setting cookie %s", cookie)
		w.Header().Set("Set-Cookie", cookie)

		userInfo.Token = token
		writeJson(w, http.StatusOK, userInfo)
	} else {
		SafeErrorAndExit(werror.ErrBadPassword, w)
		return
	}

}

// CheckUsernameUnique godoc
//
//	@ID		CheckUsernameUnique
//
//	@Summary	Check if username is unique
//	@Tags		Users
//	@Produce	json
//	@Param		username	query		string	true	"Username to check"
//	@Success	200
//	@Failure	400
//	@Failure	409
//	@Router		/users/unique [get]
func checkUsernameUnique(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := pack.UserService.Get(username)
	if user != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// LogoutUser godoc
//
//	@ID			LogoutUser
//
//	@Security	SessionAuth
//
//	@Summary	Logout User
//	@Tags		Users
//	@Success	200
//	@Router		/users/logout [post]
func logoutUser(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil || u.IsPublic() {
		// This should not happen. We must check for user before this point
		pack.Log.Error().Msg("Could not find user to logout")
	}

	cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", SessionTokenCookie)
	w.Header().Set("Set-Cookie", cookie)
	w.WriteHeader(http.StatusOK)

}

// GetUsers godoc
//
//	@ID			GetUsers
//
//	@Security	SessionAuth[admin]
//
//	@Summary	Get all users, including (possibly) sensitive information like password hashes
//	@Tags		Users
//	@Produce	json
//	@Success	200	{array}	rest.UserInfoArchive	"List of users"
//	@Failure	401
//	@Router		/users [get]
func getUsers(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil || !u.IsAdmin() {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	usersIter, err := pack.UserService.GetAll()
	if SafeErrorAndExit(err, w) {
		return
	}

	var usersInfo []rest.UserInfoArchive
	for user := range usersIter {
		usersInfo = append(usersInfo, rest.UserToUserInfoArchive(user))
	}

	writeJson(w, http.StatusOK, usersInfo)
}

// GetUser godoc
//
//	@ID			GetUser
//
//	@Security	SessionAuth
//
//	@Summary	Gets the user based on the auth token
//	@Tags		Users
//	@Produce	json
//	@Success	200	{object}	rest.UserInfo	"Logged-in users info"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/users/me [get]
func getUserInfo(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	if pack.InstanceService.GetLocal().GetRole() == models.InitServerRole {
		writeJson(w, http.StatusServiceUnavailable, rest.WeblensErrorInfo{Error: "weblens not initialized"})
		return
	}

	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	if u == nil || u.IsPublic() {
		SafeErrorAndExit(werror.ErrInvalidToken, w)
		return
	}

	res := rest.UserToUserInfo(u)

	trash, err := pack.FileService.GetFileSafe(u.TrashId, u, nil)
	if SafeErrorAndExit(err, w) {
		return
	}

	res.TrashSize = trash.Size()
	res.HomeSize = trash.GetParent().Size()

	writeJson(w, http.StatusOK, res)
}

// UpdateUserPassword godoc
//
//	@ID			UpdateUserPassword
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Update user password
//	@Tags		Users
//	@Produce	json
//
//	@Param		username				path	string						true	"Username of user to update"
//	@Param		passwordUpdateParams	body	rest.PasswordUpdateParams	true	"Password update params"
//	@Success	200
//	@Failure	400	{object}	rest.WeblensErrorInfo	"Both oldPassword and newPassword fields are required"
//	@Failure	403
//	@Failure	404
//	@Router		/users/{username}/password [patch]
func updateUserPassword(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	updateUsername := chi.URLParam(r, "username")
	updateUser := pack.UserService.Get(updateUsername)

	if updateUser == nil {
		w.WriteHeader(http.StatusNotFound)
	}

	passUpd, err := readCtxBody[rest.PasswordUpdateParams](w, r)
	if err != nil {
		return
	}

	if updateUser.GetUsername() != u.GetUsername() && !u.IsOwner() {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if (passUpd.OldPass == "" && !u.IsOwner()) || passUpd.NewPass == "" {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Both oldPassword and newPassword fields are required"})
		return
	}

	err = pack.UserService.UpdateUserPassword(
		updateUser.GetUsername(), passUpd.OldPass, passUpd.NewPass, u.IsOwner(),
	)

	if err != nil {
		pack.Log.Error().Stack().Err(err).Msg("")
		switch {
		case errors.Is(err, werror.ErrBadPassword):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SetUserAdmin godoc
//
//	@ID			SetUserAdmin
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Update admin status of user
//	@Tags		Users
//	@Produce	json
//
//	@Param		username	path	string	true	"Username of user to update"
//	@Param		setAdmin	query	bool	true	"Target admin status"
//	@Success	200
//	@Failure	400	{object}	rest.WeblensErrorInfo
//	@Failure	403
//	@Failure	404
//	@Router		/users/{username}/admin [patch]
func setUserAdmin(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	owner, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}
	if !owner.IsOwner() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	username := chi.URLParam(r, "username")
	u := pack.UserService.Get(username)
	if u == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	setAdminStr := r.URL.Query().Get("setAdmin")
	if setAdminStr == "" || (setAdminStr != "true" && setAdminStr != "false") {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "setAdmin query parameter is required and must be 'true' or 'false'"})
	}

	setAdmin := setAdminStr == "true"

	err = pack.UserService.SetUserAdmin(u, setAdmin)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ActivateUser godoc
//
//	@ID			ActivateUser
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Update active status of user
//	@Tags		Users
//	@Produce	json
//
//	@Param		username	path	string	true	"Username of user to update"
//	@Param		setActive	query	boolean	true	"Target activation status"
//	@Success	200
//	@Failure	400	{object}	rest.WeblensErrorInfo
//	@Failure	401
//	@Failure	404
//	@Router		/users/{username}/active [patch]
func activateUser(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	username := chi.URLParam(r, "username")
	u := pack.UserService.Get(username)

	setActiveStr := r.URL.Query().Get("setActive")
	if setActiveStr == "" || (setActiveStr != "true" && setActiveStr != "false") {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "setActive query parameter is required and must be 'true' or 'false'"})
	}

	setActive := setActiveStr == "true"

	err := pack.UserService.ActivateUser(u, setActive)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ChangeFullName godoc
//
//	@ID		ChangeFullName
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Update full name of user
//	@Tags		Users
//	@Produce	json
//
//	@Param		username	path	string	true	"Username of user to update"
//	@Param		newFullName	query	string	true	"New full name of user"
//	@Success	200
//	@Failure	400	{object}	rest.WeblensErrorInfo
//	@Failure	401 {object}	rest.WeblensErrorInfo
//	@Failure	404 {object}	rest.WeblensErrorInfo
//	@Router		/users/{username}/fullName [patch]
func changeFullName(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	accessor, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	username := chi.URLParam(r, "username")

	if accessor.Username != username {
		writeError(w, http.StatusForbidden, werror.ErrUserNotAuthorized)
		return
	}

	u := pack.UserService.Get(username)
	if u == nil {
		writeError(w, http.StatusNotFound, werror.ErrNoUser)
		return
	}

	newFullName := r.URL.Query().Get("newFullName")
	if newFullName == "" {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "newFullName query parameter is required"})
		return
	}

	err = pack.UserService.UpdateFullName(u, newFullName)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUser godoc
//
//	@ID			DeleteUser
//
//	@Security	SessionAuth[Admin]
//	@Security	ApiKeyAuth[Admin]
//
//	@Summary	Delete a user
//	@Tags		Users
//	@Produce	json
//
//	@Param		username	path	string	true	"Username of user to delete"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/users/{username} [delete]
func deleteUser(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	if !u.IsAdmin() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	username := chi.URLParam(r, "username")

	userToDelete := pack.UserService.Get(username)
	if userToDelete == nil {
		SafeErrorAndExit(werror.ErrNoUser, w)
		return
	}

	err = pack.UserService.Del(userToDelete.GetUsername())
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SearchUsers godoc
//
//	@ID			SearchUsers
//
//	@Security	SessionAuth
//
//	@Summary	Search for users by username
//	@Tags		Users
//	@Produce	json
//
//	@Param		search	query		string					true	"Partial username to search for"
//	@Success	200		{array}		rest.UserInfo			"List of users"
//	@Failure	400		{object}	rest.WeblensErrorInfo	"Username autocomplete must contain at least 2 characters"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/users/search [get]
func searchUsers(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	search := r.URL.Query().Get("search")
	if len(search) < 2 {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Username autocomplete must contain at least 2 characters"})
		return
	}

	users, err := pack.UserService.SearchByUsername(search)
	if err != nil {
		pack.Log.Error().Stack().Err(err).Msg("")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	usersInfo := []rest.UserInfo{}
	for user := range users {
		if user.Username == u.Username {
			continue
		}
		usersInfo = append(usersInfo, rest.UserToUserInfo(user))
	}

	writeJson(w, http.StatusOK, usersInfo)
}
