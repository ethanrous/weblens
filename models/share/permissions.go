package share

import (
	"fmt"

	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/rs/zerolog"
)

// Permission represents a specific type of access permission for file shares.
type Permission string

const (
	// SharePermissionViewMedia allows read access to the media content of the file share, if applicable.
	SharePermissionViewMedia Permission = "viewMedia"
	// SharePermissionView allows read access to the file share.
	SharePermissionView Permission = "view"
	// SharePermissionDownload allows download access to the file share.
	SharePermissionDownload Permission = "download"
	// SharePermissionEdit allows write access to the file share.
	SharePermissionEdit Permission = "edit"
	// SharePermissionDelete allows delete access to the file share.
	SharePermissionDelete Permission = "delete"
)

// Permissions represents the specific permissions a user can have on a file share.
type Permissions struct {
	CanViewMedia bool `bson:"canViewMedia"` // Indicates if the user can view media content of the share
	CanView      bool `bson:"canView"`      // Indicates if the user can view files in the share
	CanEdit      bool `bson:"canEdit"`
	CanDownload  bool `bson:"canDownload"`
	CanDelete    bool `bson:"canDelete"`
}

// NewPermissions creates a new Permissions instance with default values.
func NewPermissions() *Permissions {
	return &Permissions{
		CanViewMedia: true, // Default to allowing viewing media
		CanView:      true, // Default to allowing viewing files
		CanEdit:      false,
		CanDownload:  true, // Default to allowing downloads
		CanDelete:    false,
	}
}

// NewFullPermissions creates a new Permissions instance with all permissions enabled.
func NewFullPermissions() *Permissions {
	// Allow all permissions
	return &Permissions{
		CanViewMedia: true,
		CanView:      true,
		CanEdit:      true,
		CanDownload:  true,
		CanDelete:    true,
	}
}

// NewEmptyPermissions creates a new Permissions instance with all permissions disabled.
func NewEmptyPermissions() *Permissions {
	// Deny all permissions
	return &Permissions{
		CanViewMedia: false,
		CanView:      false,
		CanEdit:      false,
		CanDownload:  false,
		CanDelete:    false,
	}
}

// SetEdit sets the edit permission.
func (p *Permissions) SetEdit(canEdit bool) {
	p.CanEdit = canEdit
}

// SetDownload sets the download permission.
func (p *Permissions) SetDownload(canDownload bool) {
	p.CanDownload = canDownload
}

// SetDelete sets the delete permission.
func (p *Permissions) SetDelete(canDelete bool) {
	p.CanDelete = canDelete
}

// AddPermission dynamically adds a new permission to the Permissions struct.
func (p *Permissions) AddPermission(permission Permission, value bool) {
	switch permission {
	case SharePermissionEdit:
		p.CanEdit = value
	case SharePermissionDownload:
		p.CanDownload = value
	case SharePermissionDelete:
		p.CanDelete = value
	default:
	}
}

// RemovePermission dynamically removes a permission by setting it to false.
func (p *Permissions) RemovePermission(permission Permission) {
	p.AddPermission(permission, false)
}

// HasPermission checks if a specific permission is granted.
func (p *Permissions) HasPermission(permission Permission) bool {
	switch permission {
	case SharePermissionViewMedia:
		return p.CanViewMedia
	case SharePermissionView:
		return p.CanView
	case SharePermissionEdit:
		return p.CanEdit
	case SharePermissionDownload:
		return p.CanDownload
	case SharePermissionDelete:
		return p.CanDelete
	default:
		wlog.GlobalLogger().Warn().Msgf("Unknown permission: %s", permission)

		return false
	}
}

// MarshalZerologObject implements the zerolog LogObjectMarshaler interface
func (p *Permissions) MarshalZerologObject(e *zerolog.Event) {
	permsStr := fmt.Sprintf("view:%t,media:%t,edit:%t,download:%t,delete:%t", p.CanView, p.CanViewMedia, p.CanEdit, p.CanDownload, p.CanDelete)
	e.Str("permissions", permsStr)
}
