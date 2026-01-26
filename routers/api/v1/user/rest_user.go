package user

import (
	"fmt"
	"net/http"

	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/cryptography"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/modules/wlerrors"
	access_service "github.com/ethanrous/weblens/services/auth"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/reshape"
)

// Create godoc
//
//	@ID			CreateUser
//
//	@Security	ApiKeyAuth
//
//	@Summary	Create a new user
//	@Tags		Users
//	@Produce	json
//	@Param		newUserParams	body	structs.NewUserParams	true	"New user params"
//	@Success	201
//	@Failure	401
//	@Router		/users [post]
func Create(ctx ctxservice.RequestContext) {
	userParams, err := netwrk.ReadRequestBody[structs.NewUserParams](ctx.Req)
	if err != nil {
		return
	}

	newUser := &user_model.User{
		Username:    userParams.Username,
		Password:    userParams.Password,
		DisplayName: userParams.FullName,
		Activated:   userParams.AutoActivate,
		UserPerms:   user_model.UserPermissionBasic,
	}

	if err := user_model.ValidateUser(ctx, newUser); err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	err = ctx.GetFileService().CreateUserHome(ctx, newUser)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	err = user_model.SaveUser(ctx, newUser)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusCreated)
}

// Login godoc
//
//	@ID			LoginUser
//
//	@Summary	Login User
//	@Tags		Users
//	@Produce	json
//	@Param		loginParams	body		structs.LoginParams	true	"Login params"
//	@Success	200			{object}	structs.UserInfo	"Logged-in users info"
//	@Failure	401
//	@Router		/users/auth [post]
func Login(ctx ctxservice.RequestContext) {
	userCredentials, err := netwrk.ReadRequestBody[structs.LoginParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	u, err := user_model.GetUserByUsername(ctx, userCredentials.Username)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	if !u.CheckLogin(userCredentials.Password) {
		ctx.Log().Debug().Msgf("Invalid login for [%s]", userCredentials.Username)
		ctx.Error(http.StatusUnauthorized, wlerrors.New("invalid login"))

		return
	}

	ctx.Requester = u

	err = access_service.SetSessionToken(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, reshape.UserToUserInfo(ctx, u))
}

// CheckExists godoc
//
//	@ID			CheckExists
//
//	@Summary	Check if username is already taken
//	@Tags		Users
//	@Produce	json
//	@Param		username	path	string	true	"Username of user to check"
//	@Success	200
//	@Failure	400
//	@Failure	404
//	@Router		/users/{username} [head]
func CheckExists(ctx ctxservice.RequestContext) {
	username := ctx.Path("username")
	if username == "" {
		ctx.Status(http.StatusBadRequest)

		return
	}

	_, err := user_model.GetUserByUsername(ctx, username)
	if err != nil {
		ctx.Status(http.StatusNotFound)

		return
	}

	ctx.Status(http.StatusOK)
}

// Logout godoc
//
//	@ID			LogoutUser
//
//	@Security	SessionAuth
//
//	@Summary	Logout User
//	@Tags		Users
//	@Success	200
//	@Router		/users/logout [post]
func Logout(ctx ctxservice.RequestContext) {
	if ctx.Requester == nil {
		ctx.Status(http.StatusNotFound)

		return
	}

	cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", cryptography.SessionTokenCookie)
	ctx.W.Header().Set("Set-Cookie", cookie)
	ctx.Status(http.StatusOK)
}

// GetAll godoc
//
//	@ID			GetUsers
//
//	@Security	SessionAuth[admin]
//
//	@Summary	Get all users, including (possibly) sensitive information like password hashes
//	@Tags		Users
//	@Produce	json
//	@Success	200	{array}	structs.UserInfoArchive	"List of users"
//	@Failure	401
//	@Router		/users [get]
func GetAll(ctx ctxservice.RequestContext) {
	users, err := user_model.GetAllUsers(ctx)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to get all users")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	results := make([]structs.UserInfo, 0, len(users))

	for _, u := range users {
		newU := reshape.UserToUserInfo(ctx, u)
		results = append(results, newU)
	}

	ctx.JSON(http.StatusOK, results)
}

// GetMe godoc
//
//	@ID			GetUser
//
//	@Security	SessionAuth
//
//	@Summary	Gets the user based on the auth token
//	@Tags		Users
//	@Produce	json
//	@Success	200	{object}	structs.UserInfo	"Logged-in users info"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/users/me [get]
func GetMe(ctx ctxservice.RequestContext) {
	if !ctx.IsLoggedIn {
		ctx.Status(http.StatusUnauthorized)

		return
	}

	newU := reshape.UserToUserInfo(ctx, ctx.Requester)
	ctx.JSON(http.StatusOK, newU)
}

// UpdatePassword godoc
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
//	@Param		username				path	string							true	"Username of user to update"
//	@Param		passwordUpdateParams	body	structs.PasswordUpdateParams	true	"Password update params"
//	@Success	200
//	@Failure	400	{object}	structs.WeblensErrorInfo	"Both oldPassword and newPassword fields are required"
//	@Failure	403
//	@Failure	404
//	@Router		/users/{username}/password [patch]
func UpdatePassword(ctx ctxservice.RequestContext) {
	updateUsername := ctx.Path("username")

	userToUpdate, err := user_model.GetUserByUsername(ctx, updateUsername)
	if err != nil {
		ctx.Status(http.StatusNotFound)

		return
	}

	if userToUpdate.Username != ctx.Requester.Username && !ctx.Requester.IsOwner() {
		ctx.Status(http.StatusForbidden)

		return
	}

	updateParams, err := netwrk.ReadRequestBody[structs.PasswordUpdateParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	if !ctx.Requester.IsOwner() || ctx.Requester.Username == updateUsername {
		ok := userToUpdate.CheckLogin(updateParams.OldPass)

		if !ok {
			ctx.Log().Debug().Msgf("Invalid password for user [%s]", updateUsername)
			ctx.Error(http.StatusUnauthorized, wlerrors.New("invalid password"))

			return
		}
	}

	err = userToUpdate.UpdatePassword(ctx, updateParams.NewPass)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to update user password")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

// SetAdmin godoc
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
//	@Failure	400	{object}	structs.WeblensErrorInfo
//	@Failure	403
//	@Failure	404
//	@Router		/users/{username}/admin [patch]
func SetAdmin(ctx ctxservice.RequestContext) {
	if !ctx.Requester.IsOwner() {
		ctx.Status(http.StatusForbidden)

		return
	}

	adminStr := ctx.Query("setAdmin")
	if adminStr == "" {
		ctx.Status(http.StatusBadRequest)

		return
	}

	permissionLevel := user_model.UserPermissionAdmin
	if adminStr != "true" {
		permissionLevel = user_model.UserPermissionBasic
	}

	username := ctx.Path("username")

	user, err := user_model.GetUserByUsername(ctx, username)
	if err != nil {
		ctx.Status(http.StatusNotFound)

		return
	}

	err = user.UpdatePermissionLevel(ctx, permissionLevel)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to update user permission level")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

// Activate godoc
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
//	@Failure	400	{object}	structs.WeblensErrorInfo
//	@Failure	401
//	@Failure	404
//	@Router		/users/{username}/active [patch]
func Activate(ctx ctxservice.RequestContext) {
	if !ctx.Requester.IsAdmin() {
		ctx.Status(http.StatusForbidden)

		return
	}

	activeStr := ctx.Query("setActive")
	if activeStr == "" {
		ctx.Status(http.StatusBadRequest)

		return
	}

	active := false
	if activeStr == "true" {
		active = true
	}

	username := ctx.Path("username")

	user, err := user_model.GetUserByUsername(ctx, username)
	if err != nil {
		ctx.Status(http.StatusNotFound)

		return
	}

	err = user.UpdateActivationStatus(ctx, active)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to update user activation status")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

// ChangeDisplayName godoc
//
//	@ID			ChangeDisplayName
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Update display name of a user
//	@Tags		Users
//	@Produce	json
//
//	@Param		username	path	string	true	"Username of user to update"
//	@Param		newFullName	query	string	true	"New full name of user"
//	@Success	200 {object}	structs.UserInfo
//	@Failure	400	{object}	structs.WeblensErrorInfo
//	@Failure	401	{object}	structs.WeblensErrorInfo
//	@Failure	404	{object}	structs.WeblensErrorInfo
//	@Router		/users/{username}/fullName [patch]
func ChangeDisplayName(ctx ctxservice.RequestContext) {
	username := ctx.Path("username")

	newName := ctx.Query("newFullName")
	if newName == "" {
		ctx.Status(http.StatusBadRequest)

		return
	}

	u, err := user_model.GetUserByUsername(ctx, username)
	if err != nil {
		ctx.Status(http.StatusNotFound)

		return
	}

	if u.Username != ctx.Requester.Username && !ctx.Requester.IsAdmin() {
		ctx.Status(http.StatusForbidden)

		return
	}

	err = u.UpdateDisplayName(ctx, newName)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to update user full name")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, reshape.UserToUserInfo(ctx, u))
}

// Delete godoc
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
func Delete(ctx ctxservice.RequestContext) {
	if !ctx.Requester.IsOwner() {
		ctx.Status(http.StatusForbidden)

		return
	}

	username := ctx.Path("username")
	if username == "" {
		ctx.Status(http.StatusBadRequest)

		return
	}

	u, err := user_model.GetUserByUsername(ctx, username)
	if err != nil {
		ctx.Status(http.StatusNotFound)

		return
	}

	err = u.Delete(ctx)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to delete user")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

var minSearchLength = 2

// Search godoc
//
//	@ID			SearchUsers
//
//	@Security	SessionAuth
//
//	@Summary	Search for users by username
//	@Tags		Users
//	@Produce	json
//
//	@Param		search	query		string						true	"Partial username to search for"
//	@Success	200		{array}		structs.UserInfo			"List of users"
//	@Failure	400		{object}	structs.WeblensErrorInfo	"Username autocomplete must contain at least 2 characters"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/users/search [get]
func Search(ctx ctxservice.RequestContext) {
	search := ctx.Query("search")
	if len(search) < minSearchLength {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Username autocomplete must contain at least 2 characters"))

		return
	}

	users, err := user_model.SearchByUsername(ctx, search)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to search users by username")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	usersInfo := []structs.UserInfo{}

	for _, user := range users {
		if user.Username == ctx.Requester.Username {
			continue
		}

		usersInfo = append(usersInfo, reshape.UserToUserInfo(ctx, user))
	}

	ctx.JSON(http.StatusOK, usersInfo)
}
