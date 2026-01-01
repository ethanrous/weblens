//go:build test

package share_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/share"
	"github.com/stretchr/testify/assert"
)

func TestNewPermissions(t *testing.T) {
	t.Run("creates default permissions", func(t *testing.T) {
		perms := share.NewPermissions()
		assert.True(t, perms.CanView)
		assert.True(t, perms.CanDownload)
		assert.False(t, perms.CanEdit)
		assert.False(t, perms.CanDelete)
	})
}

func TestNewFullPermissions(t *testing.T) {
	t.Run("creates full permissions", func(t *testing.T) {
		perms := share.NewFullPermissions()
		assert.True(t, perms.CanView)
		assert.True(t, perms.CanDownload)
		assert.True(t, perms.CanEdit)
		assert.True(t, perms.CanDelete)
	})
}

func TestPermissionsSetEdit(t *testing.T) {
	t.Run("sets edit permission to true", func(t *testing.T) {
		perms := share.NewPermissions()
		perms.SetEdit(true)
		assert.True(t, perms.CanEdit)
	})

	t.Run("sets edit permission to false", func(t *testing.T) {
		perms := share.NewFullPermissions()
		perms.SetEdit(false)
		assert.False(t, perms.CanEdit)
	})
}

func TestPermissionsSetDownload(t *testing.T) {
	t.Run("sets download permission to true", func(t *testing.T) {
		perms := &share.Permissions{}
		perms.SetDownload(true)
		assert.True(t, perms.CanDownload)
	})

	t.Run("sets download permission to false", func(t *testing.T) {
		perms := share.NewPermissions()
		perms.SetDownload(false)
		assert.False(t, perms.CanDownload)
	})
}

func TestPermissionsSetDelete(t *testing.T) {
	t.Run("sets delete permission to true", func(t *testing.T) {
		perms := share.NewPermissions()
		perms.SetDelete(true)
		assert.True(t, perms.CanDelete)
	})

	t.Run("sets delete permission to false", func(t *testing.T) {
		perms := share.NewFullPermissions()
		perms.SetDelete(false)
		assert.False(t, perms.CanDelete)
	})
}

func TestPermissionsAddPermission(t *testing.T) {
	tests := []struct {
		name       string
		permission share.Permission
		value      bool
		check      func(*share.Permissions) bool
	}{
		{"add edit true", share.SharePermissionEdit, true, func(p *share.Permissions) bool { return p.CanEdit }},
		{"add edit false", share.SharePermissionEdit, false, func(p *share.Permissions) bool { return !p.CanEdit }},
		{"add download true", share.SharePermissionDownload, true, func(p *share.Permissions) bool { return p.CanDownload }},
		{"add download false", share.SharePermissionDownload, false, func(p *share.Permissions) bool { return !p.CanDownload }},
		{"add delete true", share.SharePermissionDelete, true, func(p *share.Permissions) bool { return p.CanDelete }},
		{"add delete false", share.SharePermissionDelete, false, func(p *share.Permissions) bool { return !p.CanDelete }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := &share.Permissions{}
			perms.AddPermission(tt.permission, tt.value)
			assert.True(t, tt.check(perms))
		})
	}

	t.Run("unknown permission does nothing", func(t *testing.T) {
		perms := share.NewPermissions()
		original := *perms
		perms.AddPermission(share.Permission("unknown"), true)
		assert.Equal(t, original, *perms)
	})
}

func TestPermissionsRemovePermission(t *testing.T) {
	t.Run("removes edit permission", func(t *testing.T) {
		perms := share.NewFullPermissions()
		perms.RemovePermission(share.SharePermissionEdit)
		assert.False(t, perms.CanEdit)
	})

	t.Run("removes download permission", func(t *testing.T) {
		perms := share.NewFullPermissions()
		perms.RemovePermission(share.SharePermissionDownload)
		assert.False(t, perms.CanDownload)
	})

	t.Run("removes delete permission", func(t *testing.T) {
		perms := share.NewFullPermissions()
		perms.RemovePermission(share.SharePermissionDelete)
		assert.False(t, perms.CanDelete)
	})
}

func TestPermissionsHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		perms      *share.Permissions
		permission share.Permission
		expected   bool
	}{
		{"default has view", share.NewPermissions(), share.SharePermissionView, true},
		{"default has download", share.NewPermissions(), share.SharePermissionDownload, true},
		{"default no edit", share.NewPermissions(), share.SharePermissionEdit, false},
		{"default no delete", share.NewPermissions(), share.SharePermissionDelete, false},
		{"full has view", share.NewFullPermissions(), share.SharePermissionView, true},
		{"full has download", share.NewFullPermissions(), share.SharePermissionDownload, true},
		{"full has edit", share.NewFullPermissions(), share.SharePermissionEdit, true},
		{"full has delete", share.NewFullPermissions(), share.SharePermissionDelete, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.perms.HasPermission(tt.permission))
		})
	}

	t.Run("unknown permission returns false", func(t *testing.T) {
		perms := share.NewFullPermissions()
		assert.False(t, perms.HasPermission(share.Permission("unknown")))
	})
}

func TestPermissionConstants(t *testing.T) {
	t.Run("SharePermissionView", func(t *testing.T) {
		assert.Equal(t, share.Permission("view"), share.SharePermissionView)
	})

	t.Run("SharePermissionDownload", func(t *testing.T) {
		assert.Equal(t, share.Permission("download"), share.SharePermissionDownload)
	})

	t.Run("SharePermissionEdit", func(t *testing.T) {
		assert.Equal(t, share.Permission("edit"), share.SharePermissionEdit)
	})

	t.Run("SharePermissionDelete", func(t *testing.T) {
		assert.Equal(t, share.Permission("delete"), share.SharePermissionDelete)
	})
}
