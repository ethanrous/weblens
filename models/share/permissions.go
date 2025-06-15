package share

type Permission string

const (
	// SharePermissionRead allows read access to the file share.
	SharePermissionDownload Permission = "download"
	// SharePermissionWrite allows write access to the file share.
	SharePermissionEdit Permission = "edit"
	// SharePermissionDelete allows delete access to the file share.
	SharePermissionDelete Permission = "delete"
)

// Permissions represents the specific permissions a user can have on a file share.
type Permissions struct {
	CanView     bool `bson:"canView"` // Indicates if the user can view the share
	CanEdit     bool `bson:"canEdit"`
	CanDownload bool `bson:"canDownload"`
	CanDelete   bool `bson:"canDelete"`
}

// NewPermissions creates a new Permissions instance with default values.
func NewPermissions() *Permissions {
	return &Permissions{
		CanView:     true, // Default to allowing viewing
		CanEdit:     false,
		CanDownload: true, // Default to allowing downloads
		CanDelete:   false,
	}
}

func NewFullPermissions() *Permissions {
	// Allow all permissions
	return &Permissions{
		CanView:     true,
		CanEdit:     true,
		CanDownload: true,
		CanDelete:   true,
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
		// Handle unknown permissions if needed
	}
}

// RemovePermission dynamically removes a permission by setting it to false.
func (p *Permissions) RemovePermission(permission Permission) {
	p.AddPermission(permission, false)
}

// HasPermission checks if a specific permission is granted.
func (p *Permissions) HasPermission(permission Permission) bool {
	switch permission {
	case SharePermissionEdit:
		return p.CanEdit
	case SharePermissionDownload:
		return p.CanDownload
	case SharePermissionDelete:
		return p.CanDelete
	default:
		return false
	}
}
