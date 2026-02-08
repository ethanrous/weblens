package e2e_test

import (
	"net/http"
	"testing"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	// Setup a core server for the test
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "testuser",
		Password:     "TestPass123",
		FullName:     "Test User",
		AutoActivate: &autoActivate,
	}).Execute()
	assert.NoError(t, err)

	users, _, err := client.UsersAPI.GetUsers(t.Context()).Execute()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))
}

func TestLoginUser(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a user first
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "loginuser",
		Password:     "TestPass123",
		FullName:     "Login User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Test successful login
	userInfo, resp, err := client.UsersAPI.LoginUser(t.Context()).LoginParams(openapi.LoginBody{
		Username: "loginuser",
		Password: "TestPass123",
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "loginuser", userInfo.GetUsername())

	// Test login with wrong password
	_, resp, err = client.UsersAPI.LoginUser(t.Context()).LoginParams(openapi.LoginBody{
		Username: "loginuser",
		Password: "WrongPassword",
	}).Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCheckUserExists(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a user first
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "existsuser",
		Password:     "TestPass123",
		FullName:     "Exists User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Check existing user
	resp, err := client.UsersAPI.CheckExists(t.Context(), "existsuser").Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check non-existing user
	resp, err = client.UsersAPI.CheckExists(t.Context(), "nonexistentuser").Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetUser(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Get the admin user (created during setup)
	userInfo, resp, err := client.UsersAPI.GetUser(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "admin", userInfo.GetUsername())
}

func TestUpdatePassword(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a user first
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "pwduser",
		Password:     "OldPass123",
		FullName:     "Password User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Update password (as admin, no old password needed)
	oldPass := "OldPass123"
	resp, err := client.UsersAPI.UpdateUserPassword(t.Context(), "pwduser").PasswordUpdateParams(openapi.PasswordUpdateParams{
		OldPassword: &oldPass,
		NewPassword: "NewPass456",
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify login works with new password
	userInfo, _, err := client.UsersAPI.LoginUser(t.Context()).LoginParams(openapi.LoginBody{
		Username: "pwduser",
		Password: "NewPass456",
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, "pwduser", userInfo.GetUsername())
}

func TestSetUserAdmin(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a user first
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "adminuser",
		Password:     "TestPass123",
		FullName:     "Admin User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Set user as admin
	resp, err := client.UsersAPI.SetUserAdmin(t.Context(), "adminuser").SetAdmin(true).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify user is admin by checking users list
	// Permission level 2 = admin
	users, _, err := client.UsersAPI.GetUsers(t.Context()).Execute()
	require.NoError(t, err)

	for _, u := range users {
		if u.GetUsername() == "adminuser" {
			assert.GreaterOrEqual(t, u.GetPermissionLevel(), int32(2))
		}
	}

	// Remove admin status
	resp, err = client.UsersAPI.SetUserAdmin(t.Context(), "adminuser").SetAdmin(false).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestActivateUser(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create an inactive user
	autoActivate := false
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "inactiveuser",
		Password:     "TestPass123",
		FullName:     "Inactive User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Login should fail for inactive user
	_, resp, err := client.UsersAPI.LoginUser(t.Context()).LoginParams(openapi.LoginBody{
		Username: "inactiveuser",
		Password: "TestPass123",
	}).Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Activate user
	resp, err = client.UsersAPI.ActivateUser(t.Context(), "inactiveuser").SetActive(true).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Login should now succeed
	userInfo, _, err := client.UsersAPI.LoginUser(t.Context()).LoginParams(openapi.LoginBody{
		Username: "inactiveuser",
		Password: "TestPass123",
	}).Execute()
	require.NoError(t, err)
	assert.Equal(t, "inactiveuser", userInfo.GetUsername())

	// Deactivate user
	resp, err = client.UsersAPI.ActivateUser(t.Context(), "inactiveuser").SetActive(false).Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestChangeDisplayName(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a user first
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "displayuser",
		Password:     "TestPass123",
		FullName:     "Original Name",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Change display name
	userInfo, resp, err := client.UsersAPI.ChangeDisplayName(t.Context(), "displayuser").NewFullName("New Display Name").Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "New Display Name", userInfo.GetFullName())
}

func TestDeleteUser(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create a user first
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "deleteuser",
		Password:     "TestPass123",
		FullName:     "Delete User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Verify user exists
	users, _, err := client.UsersAPI.GetUsers(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, 2, len(users)) // admin + deleteuser

	// Delete user
	resp, err := client.UsersAPI.DeleteUser(t.Context(), "deleteuser").Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify user no longer exists
	users, _, err = client.UsersAPI.GetUsers(t.Context()).Execute()
	require.NoError(t, err)
	assert.Equal(t, 1, len(users)) // only admin
}

func TestSearchUsers(t *testing.T) {
	coreSetup, err := setupTestServer(t.Context(), t.Name(), config.Provider{InitRole: string(tower.RoleCore), GenerateAdminAPIToken: true})
	require.NoError(t, err, "Failed to start test server")

	client := getAPIClientFromConfig(coreSetup.cnf, coreSetup.token)

	// Create multiple users
	autoActivate := true
	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "searchuser1",
		Password:     "TestPass123",
		FullName:     "Search User One",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "searchuser2",
		Password:     "TestPass123",
		FullName:     "Search User Two",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	_, err = client.UsersAPI.CreateUser(t.Context()).NewUserParams(openapi.NewUserParams{
		Username:     "otheruser",
		Password:     "TestPass123",
		FullName:     "Other User",
		AutoActivate: &autoActivate,
	}).Execute()
	require.NoError(t, err)

	// Search for "search" - should find searchuser1 and searchuser2
	results, resp, err := client.UsersAPI.SearchUsers(t.Context()).Search("search").Execute()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, len(results))

	// Search with short query - should fail
	_, resp, err = client.UsersAPI.SearchUsers(t.Context()).Search("a").Execute()
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
