package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/wlerrors"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/services/auth"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/viccon/sturdyc"
)

// stubFileService implements file_model.Service with a fixed ID→file map.
// Only GetFileByID is implemented; all other methods panic if called.
type stubFileService struct {
	files map[string]*file_model.WeblensFileImpl
}

func (s *stubFileService) GetFileByID(_ context.Context, fileID string) (*file_model.WeblensFileImpl, error) {
	f, ok := s.files[fileID]
	if !ok {
		return nil, wlerrors.Errorf("%w: %s", file_model.ErrFileNotFound, fileID)
	}

	return f, nil
}

func (s *stubFileService) AddFile(_ context.Context, _ ...*file_model.WeblensFileImpl) error {
	panic("not implemented")
}

func (s *stubFileService) Size(_ string) int64 { panic("not implemented") }

func (s *stubFileService) GetFileByFilepath(_ context.Context, _ file_system.Filepath, _ ...bool) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) CreateFile(_ context.Context, _ *file_model.WeblensFileImpl, _ string, _ ...[]byte) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) CreateFolder(_ context.Context, _ *file_model.WeblensFileImpl, _ string) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) GetChildren(_ context.Context, _ *file_model.WeblensFileImpl) ([]*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) RecursiveEnsureChildrenLoaded(_ context.Context, _ *file_model.WeblensFileImpl) error {
	panic("not implemented")
}

func (s *stubFileService) CreateUserHome(_ context.Context, _ *user_model.User) error {
	panic("not implemented")
}

func (s *stubFileService) NewBackupRestoreFile(_ context.Context, _, _ string) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) InitBackupDirectory(_ context.Context, _ tower_model.Instance) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) MoveFiles(_ context.Context, _ []*file_model.WeblensFileImpl, _ *file_model.WeblensFileImpl) error {
	panic("not implemented")
}

func (s *stubFileService) RenameFile(_ context.Context, _ *file_model.WeblensFileImpl, _ string) error {
	panic("not implemented")
}

func (s *stubFileService) ReturnFilesFromTrash(_ context.Context, _ []*file_model.WeblensFileImpl) error {
	panic("not implemented")
}

func (s *stubFileService) DeleteFiles(_ context.Context, _ ...*file_model.WeblensFileImpl) error {
	panic("not implemented")
}

func (s *stubFileService) RestoreFiles(_ context.Context, _ []string, _ *file_model.WeblensFileImpl, _ time.Time) error {
	panic("not implemented")
}

func (s *stubFileService) GetMediaCacheByFilename(_ context.Context, _ string) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) GetFileByContentID(_ context.Context, _ string) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) NewCacheFile(_ string, _ string, _ int) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) DeleteCacheFile(_ *file_model.WeblensFileImpl) error {
	panic("not implemented")
}

func (s *stubFileService) NewZip(_ context.Context, _ string, _ *user_model.User) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) GetZip(_ context.Context, _ string) (*file_model.WeblensFileImpl, error) {
	panic("not implemented")
}

func (s *stubFileService) DeleteZips(_ context.Context) error {
	panic("not implemented")
}

// newTestFile creates a test file owned by the given username (via path convention).
func newTestFile(owner, name string) *file_model.WeblensFileImpl {
	fp := file_system.BuildFilePath(file_model.UsersTreeKey, owner+"/"+name)

	return file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:       fp,
		MemOnly:    true,
		GenerateID: true,
	})
}

// newTestRequestCtx creates a RequestContext backed by a stub file service and a recording response writer.
func newTestRequestCtx(requester *user_model.User, files map[string]*file_model.WeblensFileImpl, w http.ResponseWriter) ctxservice.RequestContext {
	svc := &stubFileService{files: files}

	logger := wlog.NewZeroLogger()
	basicCtx := ctxservice.NewBasicContext(context.Background(), logger)

	appCtx := ctxservice.AppContext{
		BasicContext: basicCtx,
		FileService:  svc,
		Cache:        make(map[string]*sturdyc.Client[any]),
		WG:           &sync.WaitGroup{},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	return ctxservice.RequestContext{
		AppContext:  appCtx,
		ReqCtx:     req.Context(),
		Req:        req,
		W:          w,
		Requester:  requester,
		IsLoggedIn: requester != nil,
	}
}

func TestCanUserAccessFileByID_HappyPath(t *testing.T) {
	owner := &user_model.User{Username: "alice", UserPerms: user_model.UserPermissionBasic}
	f := newTestFile("alice", "photo.jpg")

	ctx := newTestRequestCtx(owner, map[string]*file_model.WeblensFileImpl{f.ID(): f}, httptest.NewRecorder())

	got, err := auth.CanUserAccessFileByID(ctx, f.ID())
	require.NoError(t, err)
	assert.Equal(t, f.ID(), got.ID())
}

func TestCanUserAccessFileByID_NotFound(t *testing.T) {
	owner := &user_model.User{Username: "alice", UserPerms: user_model.UserPermissionBasic}

	ctx := newTestRequestCtx(owner, map[string]*file_model.WeblensFileImpl{}, httptest.NewRecorder())

	_, err := auth.CanUserAccessFileByID(ctx, "nonexistent-id")
	require.Error(t, err)

	code, _ := wlerrors.AsStatus(err, 0)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestCanUserAccessFileByID_PermissionDenied(t *testing.T) {
	other := &user_model.User{Username: "bob", UserPerms: user_model.UserPermissionBasic}
	f := newTestFile("alice", "secret.jpg")

	ctx := newTestRequestCtx(other, map[string]*file_model.WeblensFileImpl{f.ID(): f}, httptest.NewRecorder())

	_, err := auth.CanUserAccessFileByID(ctx, f.ID())
	require.Error(t, err)

	code, _ := wlerrors.AsStatus(err, 0)
	assert.Equal(t, http.StatusForbidden, code)
}

func TestRequireFileAccess_AllMustPassDenied(t *testing.T) {
	other := &user_model.User{Username: "bob", UserPerms: user_model.UserPermissionBasic}

	f1 := newTestFile("alice", "file1.jpg")
	f2 := newTestFile("alice", "file2.jpg")

	w := httptest.NewRecorder()
	ctx := newTestRequestCtx(other, map[string]*file_model.WeblensFileImpl{
		f1.ID(): f1,
		f2.ID(): f2,
	}, w)

	_, err := auth.RequireFileAccess(ctx, []string{f1.ID(), f2.ID()})
	require.Error(t, err)

	// ctx.Error should have been called exactly once - the recorder should have a non-200 status.
	assert.NotEqual(t, http.StatusOK, w.Code, "ctx.Error should have written an error response")

	// Only one error response should have been written (first failure stops the loop).
	// The body should contain exactly one JSON error object - verify it is valid JSON.
	assert.Greater(t, w.Body.Len(), 0, "response body should be non-empty")
}

func TestRequireAnyFileAccess_FirstPassesHappyPath(t *testing.T) {
	owner := &user_model.User{Username: "alice", UserPerms: user_model.UserPermissionBasic}

	f1 := newTestFile("alice", "file1.jpg")
	f2 := newTestFile("alice", "file2.jpg")

	w := httptest.NewRecorder()
	ctx := newTestRequestCtx(owner, map[string]*file_model.WeblensFileImpl{
		f1.ID(): f1,
		f2.ID(): f2,
	}, w)

	got, err := auth.RequireAnyFileAccess(ctx, []string{f1.ID(), f2.ID()})
	require.NoError(t, err)
	assert.Equal(t, f1.ID(), got.ID())

	// ctx.Error must NOT have been called.
	assert.Equal(t, http.StatusOK, w.Code, "ctx.Error should not have been called when first file passes")
}

func TestRequireAnyFileAccess_FirstFailsSecondPasses(t *testing.T) {
	bob := &user_model.User{Username: "bob", UserPerms: user_model.UserPermissionBasic}

	// f1 is owned by alice - bob will be denied.
	f1 := newTestFile("alice", "file1.jpg")
	// f2 is owned by bob - bob will be allowed.
	f2 := newTestFile("bob", "file2.jpg")

	w := httptest.NewRecorder()
	ctx := newTestRequestCtx(bob, map[string]*file_model.WeblensFileImpl{
		f1.ID(): f1,
		f2.ID(): f2,
	}, w)

	got, err := auth.RequireAnyFileAccess(ctx, []string{f1.ID(), f2.ID()})
	require.NoError(t, err, "should succeed because bob owns f2")
	assert.Equal(t, f2.ID(), got.ID())

	// ctx.Error must NOT have been called even though f1 was denied.
	assert.Equal(t, http.StatusOK, w.Code, "ctx.Error should not be called when any file passes")
}

func TestRequireAnyFileAccess_AllFail(t *testing.T) {
	bob := &user_model.User{Username: "bob", UserPerms: user_model.UserPermissionBasic}

	f1 := newTestFile("alice", "file1.jpg")
	f2 := newTestFile("alice", "file2.jpg")

	w := httptest.NewRecorder()
	ctx := newTestRequestCtx(bob, map[string]*file_model.WeblensFileImpl{
		f1.ID(): f1,
		f2.ID(): f2,
	}, w)

	_, err := auth.RequireAnyFileAccess(ctx, []string{f1.ID(), f2.ID()})
	require.Error(t, err)

	// ctx.Error should have been called exactly once after all IDs failed.
	assert.NotEqual(t, http.StatusOK, w.Code, "ctx.Error should have been called when all files fail")
}

func TestRequireAnyFileAccess_HardErrorBreaksLoop(t *testing.T) {
	bob := &user_model.User{Username: "bob", UserPerms: user_model.UserPermissionBasic}

	// f2 is owned by bob and would succeed if reached. But f1 is a missing ID, which
	// produces a 404 hard error. Any-of semantics should NOT swallow hard errors -
	// the loop must break on the 404 rather than fall through to f2.
	f2 := newTestFile("bob", "file2.jpg")

	w := httptest.NewRecorder()
	ctx := newTestRequestCtx(bob, map[string]*file_model.WeblensFileImpl{
		f2.ID(): f2,
	}, w)

	_, err := auth.RequireAnyFileAccess(ctx, []string{"missing-file-id", f2.ID()})
	require.Error(t, err)

	status, _ := wlerrors.AsStatus(err, 0)
	assert.Equal(t, http.StatusNotFound, status, "hard 404 error should surface immediately, not fall through")
}
