// Package services provides WebDAV filesystem implementation for Weblens.
package services

// import "golang.org/x/net/webdav"

// var _ webdav.FileSystem = (*WebdavFs)(nil)

// WebdavFs is a placeholder type for WebDAV filesystem functionality.
type WebdavFs struct{}

// func (w WebdavFs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
// 	unescapeName, err := url.QueryUnescape(name)
// 	if err != nil {
// 		return err
// 	}
//
// 	parent, err := w.WeblensFs.PathToFile(filepath.Dir(unescapeName))
// 	if err != nil {
// 		return err
// 	}
//
// 	// TODO: add event
// 	_, err = w.WeblensFs.CreateFile(parent, filepath.Base(unescapeName), nil)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (w WebdavFs) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
// 	unescapeName, err := url.QueryUnescape(name)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// filename := filepath.Base(unescapeName)
// 	// if strings.HasPrefix(filename, "._") {
// 	// 	filename = filename[2:]
// 	// 	unescapeName = filepath.Dir(unescapeName) + "/" + filename
// 	// }
// 	// if unescapeName == "." {
// 	// 	unescapeName = "/"
// 	// }
//
// 	f, err := w.WeblensFs.PathToFile(unescapeName)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return f, nil
// }
//
// func (w WebdavFs) RemoveAll(ctx context.Context, name string) error {
//
// 	panic("implement me")
// }
//
// func (w WebdavFs) Rename(ctx context.Context, oldName, newName string) error {
// 	unescapeOldName, err := url.QueryUnescape(oldName)
// 	if err != nil {
// 		return err
// 	}
// 	unescapeNewName, err := url.QueryUnescape(newName)
// 	if err != nil {
// 		return err
// 	}
//
// 	oldFile, err := w.WeblensFs.PathToFile(unescapeOldName)
// 	if err != nil {
// 		return err
// 	}
//
// 	newParent, err := w.WeblensFs.PathToFile(filepath.Dir(unescapeNewName))
// 	if err != nil {
// 		return err
// 	}
//
// 	err = w.WeblensFs.MoveFiles([]*fileTree.WeblensFileImpl{oldFile}, newParent, "USERS", w.Caster)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (w WebdavFs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
// 	unescapeName, err := url.QueryUnescape(name)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	filename := filepath.Base(unescapeName)
// 	if strings.HasPrefix(filename, "._") {
// 		filename = filename[2:]
// 		unescapeName = filepath.Dir(unescapeName) + "/" + filename
// 	}
//
// 	f, err := w.WeblensFs.PathToFile(unescapeName)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return f, nil
// }
