package fs_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFilePath(t *testing.T) {
	tests.Setup(t)

	tests := []struct {
		name     string
		root     string
		relPaths []string
		want     fs.Filepath
	}{
		{
			name:     "simple path",
			root:     "test",
			relPaths: []string{"foo", "bar.txt"},
			want:     fs.Filepath{RootAlias: "test", RelPath: "foo/bar.txt"},
		},
		{
			name:     "directory path",
			root:     "test",
			relPaths: []string{"foo", "bar/"},
			want:     fs.Filepath{RootAlias: "test", RelPath: "foo/bar/"},
		},
		{
			name:     "empty path",
			root:     "test",
			relPaths: []string{},
			want:     fs.Filepath{RootAlias: "test", RelPath: ""},
		},
		{
			name:     "single file",
			root:     "test",
			relPaths: []string{"file.txt"},
			want:     fs.Filepath{RootAlias: "test", RelPath: "file.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.BuildFilePath(tt.root, tt.relPaths...)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewFilePath(t *testing.T) {
	tests.Setup(t)

	// Setup test directory structure
	tmpDir := t.TempDir()
	err := fs.RegisterAbsolutePrefix("test", tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name         string
		rootAlias    string
		absolutePath string
		wantErr      bool
	}{
		{
			name:         "valid path",
			rootAlias:    "test",
			absolutePath: filepath.Join(tmpDir, "foo/bar.txt"),
			wantErr:      false,
		},
		{
			name:         "invalid root",
			rootAlias:    "invalid",
			absolutePath: filepath.Join(tmpDir, "foo/bar.txt"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fs.NewFilePath(tt.rootAlias, tt.absolutePath)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestParsePortable(t *testing.T) {
	tests.Setup(t)

	tests := []struct {
		name        string
		portableStr string
		want        fs.Filepath
		wantErr     bool
	}{
		{
			name:        "valid portable path",
			portableStr: "test:foo/bar.txt",
			want:        fs.Filepath{RootAlias: "test", RelPath: "foo/bar.txt"},
			wantErr:     false,
		},
		{
			name:        "invalid format - no colon",
			portableStr: "testfoo/bar.txt",
			wantErr:     true,
		},
		{
			name:        "empty path after colon",
			portableStr: "test:",
			want:        fs.Filepath{RootAlias: "test", RelPath: ""},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.ParsePortable(tt.portableStr)

			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBSONMarshaling(t *testing.T) {
	tests.Setup(t)

	tests := []struct {
		name     string
		filepath fs.Filepath
		wantErr  bool
	}{
		{
			name:     "marshal regular path",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/bar.txt"},
			wantErr:  false,
		},
		{
			name:     "marshal empty path",
			filepath: fs.Filepath{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			bsonType, data, err := tt.filepath.MarshalBSONValue()
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)

			// Test unmarshaling
			var fp fs.Filepath

			err = fp.UnmarshalBSONValue(bsonType, data)
			assert.NoError(t, err)
			assert.Equal(t, tt.filepath, fp)
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	tests.Setup(t)

	tests := []struct {
		name     string
		filepath fs.Filepath
		want     string
		err      error
	}{
		{
			name:     "marshal regular path",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/bar.txt"},
			want:     `"test:foo/bar.txt"`,
		},
		{
			name:     "marshal empty path",
			filepath: fs.Filepath{},
			want:     `""`,
			err:      fs.ErrInvalidPortablePath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filepath.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(got))

			// Test unmarshaling
			var fp fs.Filepath

			err = (&fp).UnmarshalJSON(got)

			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.filepath, fp)
		})
	}
}

func TestFilepathMethods(t *testing.T) {
	tests.Setup(t)

	tests := []struct {
		name     string
		filepath fs.Filepath
		checks   func(*testing.T, fs.Filepath)
	}{
		{
			name:     "IsZero check",
			filepath: fs.Filepath{},
			checks: func(t *testing.T, fp fs.Filepath) {
				assert.True(t, fp.IsZero())
				assert.True(t, fs.IsZeroFilepath(fp))
			},
		},
		{
			name:     "IsRoot check",
			filepath: fs.Filepath{RootAlias: "test", RelPath: ""},
			checks: func(t *testing.T, fp fs.Filepath) {
				assert.True(t, fp.IsRoot())
			},
		},
		{
			name:     "Dir path",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/bar/file.txt"},
			checks: func(t *testing.T, fp fs.Filepath) {
				dir := fp.Dir()
				assert.Equal(t, "foo/bar/", dir.RelPath)
			},
		},
		{
			name:     "Filename",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/bar/file.txt"},
			checks: func(t *testing.T, fp fs.Filepath) {
				assert.Equal(t, "file.txt", fp.Filename())
			},
		},
		{
			name:     "IsDir check",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/bar/"},
			checks: func(t *testing.T, fp fs.Filepath) {
				assert.True(t, fp.IsDir())
			},
		},
		{
			name:     "Child path",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/"},
			checks: func(t *testing.T, fp fs.Filepath) {
				child := fp.Child("bar.txt", false)
				assert.Equal(t, "foo/bar.txt", child.RelPath)
			},
		},
		{
			name:     "Extension",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/bar.txt"},
			checks: func(t *testing.T, fp fs.Filepath) {
				assert.Equal(t, ".txt", fp.Ext())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checks(t, tt.filepath)
		})
	}
}

func TestReplacePrefix(t *testing.T) {
	tests.Setup(t)

	tests := []struct {
		name      string
		path      fs.Filepath
		prefix    fs.Filepath
		newPrefix fs.Filepath
		want      fs.Filepath
		wantErr   bool
	}{
		{
			name:      "valid prefix replacement",
			path:      fs.Filepath{RootAlias: "old", RelPath: "prefix/path/file.txt"},
			prefix:    fs.Filepath{RootAlias: "old", RelPath: "prefix/"},
			newPrefix: fs.Filepath{RootAlias: "new", RelPath: "newprefix/"},
			want:      fs.Filepath{RootAlias: "new", RelPath: "newprefix/path/file.txt"},
			wantErr:   false,
		},
		{
			name:      "invalid prefix",
			path:      fs.Filepath{RootAlias: "old", RelPath: "path/file.txt"},
			prefix:    fs.Filepath{RootAlias: "old", RelPath: "prefix/"},
			newPrefix: fs.Filepath{RootAlias: "new", RelPath: "newprefix/"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.path.ReplacePrefix(tt.prefix, tt.newPrefix)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAbsolutePathRegistration(t *testing.T) {
	tests.Setup(t)

	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		alias    string
		path     string
		wantErr  bool
		checkDir bool
	}{
		{
			name:     "valid registration",
			alias:    "test1",
			path:     filepath.Join(tmpDir, "test1"),
			wantErr:  false,
			checkDir: true,
		},
		{
			name:    "invalid path - not absolute",
			alias:   "test2",
			path:    "relative/path",
			wantErr: true,
		},
		{
			name:     "path with no trailing slash",
			alias:    "test3",
			path:     filepath.Join(tmpDir, "test3"),
			wantErr:  false,
			checkDir: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.RegisterAbsolutePrefix(tt.alias, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.checkDir {
					_, err := os.Stat(tt.path)
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestConcurrentPathOperations(t *testing.T) {
	tests.Setup(t)

	tmpDir := t.TempDir()

	const numGoroutines = 10

	const numOperations = 100

	// Create channels for synchronization
	done := make(chan bool)
	errs := make(chan error, numGoroutines)

	// Start multiple goroutines that register and access paths concurrently
	for id := range numGoroutines {
		go func(id int) {
			for op := range numOperations {
				// Register a new path
				alias := fmt.Sprintf("test%d_%d", id, op)
				path := filepath.Join(tmpDir, alias)

				err := fs.RegisterAbsolutePrefix(alias, path)
				if err != nil {
					errs <- fmt.Errorf("failed to register path: %v", err)

					return
				}

				// Create a filepath using the registered alias
				fp := fs.BuildFilePath(alias, "test.txt")

				// Access the absolute path
				absPath := fp.ToAbsolute()
				if absPath == "" {
					errs <- fmt.Errorf("failed to get absolute path for %s", alias)

					return
				}
			}

			done <- true
		}(id)
	}

	// Wait for all goroutines to complete or error out
	for range numGoroutines {
		select {
		case err := <-errs:
			t.Errorf("goroutine error: %v", err)
		case <-done:
		}
	}
}

func TestToAbsolute(t *testing.T) {
	tests.Setup(t)

	tmpDir := t.TempDir()
	err := fs.RegisterAbsolutePrefix("test", tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name     string
		filepath fs.Filepath
		want     string
	}{
		{
			name:     "valid path",
			filepath: fs.Filepath{RootAlias: "test", RelPath: "foo/bar.txt"},
			want:     filepath.Join(tmpDir, "foo/bar.txt"),
		},
		{
			name:     "empty filepath",
			filepath: fs.Filepath{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filepath.ToAbsolute()
			// Normalize paths for comparison
			if got != "" {
				got = filepath.Clean(got)
			}

			if tt.want != "" {
				tt.want = filepath.Clean(tt.want)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
