package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/gin-gonic/gin"
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
func createUser(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	userParams, err := readCtxBody[rest.NewUserParams](ctx)
	if err != nil {
		return
	}

	if !u.Admin && (userParams.AutoActivate || userParams.Admin) {
		ctx.Status(http.StatusForbidden)
		return
	}

	newUser, err := models.NewUser(userParams.Username, userParams.Password, userParams.Admin, userParams.AutoActivate)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	err = pack.FileService.CreateUserHome(newUser)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	err = pack.UserService.Add(newUser)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusCreated)
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
func loginUser(ctx *gin.Context) {
	pack := getServices(ctx)
	userCredentials, err := readCtxBody[rest.LoginBody](ctx)
	if err != nil {
		return
	}

	u := pack.UserService.Get(userCredentials.Username)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if !u.Activated {
		log.Warning.Printf("[%s] attempted login but is not activated", u.Username)
		ctx.Status(http.StatusUnauthorized)
		return
	}

	if u.CheckLogin(userCredentials.Password) {
		log.Debug.Printf("Valid login for [%s]\n", userCredentials.Username)

		var token string
		// var expires time.Time
		token, _, err = pack.AccessService.GenerateJwtToken(u)
		if err != nil || token == "" {
			log.ErrTrace(werror.Errorf("Could not get login token"))
			ctx.Status(http.StatusInternalServerError)
		}

		userInfo := rest.UserToUserInfo(u)

		// cookie := fmt.Sprintf("%s=%s; expires=%s;", SessionTokenCookie, token, expires.Format(time.RFC1123))
		cookie := fmt.Sprintf("%s=%s;Path=/;HttpOnly", SessionTokenCookie, token)

		log.Trace.Println("Setting cookie", cookie)
		ctx.Header("Set-Cookie", cookie)
		ctx.JSON(http.StatusOK, userInfo)
	} else {
		log.Error.Printf("Invalid login for [%s]", userCredentials.Username)
		ctx.Status(http.StatusNotFound)
	}

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
func logoutUser(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil || u.IsPublic() {
		// This should not happen. We must check for user before this point
		log.Error.Panicln("Could not find user to logout")
	}

	cookie := fmt.Sprintf("%s=;Path=/;expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", SessionTokenCookie)
	ctx.Header("Set-Cookie", cookie)
	ctx.Status(http.StatusOK)

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
func getUsers(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil || !u.IsAdmin() {
		ctx.Status(http.StatusNotFound)
		return
	}

	usersIter, err := pack.UserService.GetAll()
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var usersInfo []rest.UserInfoArchive
	for user := range usersIter {
		usersInfo = append(usersInfo, rest.UserToUserInfoArchive(user))
	}

	ctx.JSON(http.StatusOK, usersInfo)
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
//	@Router		/users/me [get]
func getUserInfo(ctx *gin.Context) {
	pack := getServices(ctx)
	if pack.InstanceService.GetLocal().GetRole() == models.InitServer {
		ctx.JSON(http.StatusTemporaryRedirect, gin.H{"error": "weblens not initialized"})
		return
	}

	u := getUserFromCtx(ctx)
	if u == nil || u.IsPublic() {
		log.Trace.Println("Could not find user")
		ctx.Status(http.StatusNotFound)
		return
	}

	res := rest.UserToUserInfo(u)
	ctx.JSON(http.StatusOK, res)
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
func updateUserPassword(ctx *gin.Context) {
	pack := getServices(ctx)
	reqUser := getUserFromCtx(ctx)

	updateUsername := ctx.Param("username")
	updateUser := pack.UserService.Get(updateUsername)

	if updateUser == nil {
		ctx.Status(http.StatusNotFound)
	}

	passUpd, err := readCtxBody[rest.PasswordUpdateParams](ctx)
	if err != nil {
		return
	}

	if updateUser.GetUsername() != reqUser.GetUsername() && !reqUser.IsOwner() {
		ctx.Status(http.StatusNotFound)
		return
	}

	if (passUpd.OldPass == "" && !reqUser.IsOwner()) || passUpd.NewPass == "" {
		ctx.JSON(http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Both oldPassword and newPassword fields are required"})
		return
	}

	err = pack.UserService.UpdateUserPassword(
		updateUser.GetUsername(), passUpd.OldPass, passUpd.NewPass, reqUser.IsOwner(),
	)

	if err != nil {
		log.ShowErr(err)
		switch {
		case errors.Is(err, werror.ErrBadPassword):
			ctx.Status(http.StatusForbidden)
		default:
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	ctx.Status(http.StatusOK)
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
func setUserAdmin(ctx *gin.Context) {
	pack := getServices(ctx)
	owner := getUserFromCtx(ctx)
	if !owner.IsOwner() {
		ctx.Status(http.StatusForbidden)
		return
	}

	username := ctx.Param("username")
	u := pack.UserService.Get(username)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	setAdminStr := ctx.Query("setAdmin")
	if setAdminStr == "" || (setAdminStr != "true" && setAdminStr != "false") {
		ctx.JSON(http.StatusBadRequest, rest.WeblensErrorInfo{Error: "setAdmin query parameter is required and must be 'true' or 'false'"})
	}

	setAdmin := setAdminStr == "true"

	err := pack.UserService.SetUserAdmin(u, setAdmin)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)
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
//	@Param		setActive	query	boolean	true	"Target admin status"
//	@Success	200
//	@Failure	400	{object}	rest.WeblensErrorInfo
//	@Failure	401
//	@Failure	404
//	@Router		/users/{username}/active [patch]
func activateUser(ctx *gin.Context) {
	pack := getServices(ctx)
	username := ctx.Param("username")
	u := pack.UserService.Get(username)

	setActiveStr := ctx.Query("setActive")
	if setActiveStr == "" || (setActiveStr != "true" && setActiveStr != "false") {
		ctx.JSON(http.StatusBadRequest, rest.WeblensErrorInfo{Error: "setActive query parameter is required and must be 'true' or 'false'"})
	}

	setActive := setActiveStr == "true"

	err := pack.UserService.ActivateUser(u, setActive)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)
}

// DeleteUser godoc
//
//	@ID			DeleteUser
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
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
func deleteUser(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if !u.IsAdmin() {
		ctx.Status(http.StatusForbidden)
		return
	}

	username := ctx.Param("username")

	deleteUser := pack.UserService.Get(username)
	if deleteUser == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	err := pack.UserService.Del(deleteUser.GetUsername())
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusOK)
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
func searchUsers(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	search := ctx.Query("search")
	if len(search) < 2 {
		ctx.JSON(http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Username autocomplete must contain at least 2 characters"})
		return
	}

	users, err := pack.UserService.SearchByUsername(search)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	var usersInfo []rest.UserInfo
	for user := range users {
		if user.Username == u.Username {
			continue
		}
		usersInfo = append(usersInfo, rest.UserToUserInfo(user))
	}

	ctx.JSON(http.StatusOK, usersInfo)
}
