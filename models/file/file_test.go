package file_test

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/file"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoot is the root alias used for testing.
const TestRoot = "test"

// testSetup creates a temporary directory for testing and returns a cleanup function.
func testSetup(t *testing.T) {
	tests.Setup(t)

	tempDir, err := os.MkdirTemp("", "weblens_test_*")
	require.NoError(t, err)

	err = file_system.RegisterAbsolutePrefix(TestRoot, tempDir)
	if err != nil {
		t.Fatalf("Failed to register test root: %v", err)
	}

	cleanup := func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}

	t.Cleanup(cleanup)
}

func TestWeblensFile_BasicOperations(t *testing.T) {
	testSetup(t)

	t.Run("CreateAndVerifyFile", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "test.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		data := []byte("test content")
		n, err := f.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)

		assert.True(t, f.Exists())

		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, data, content)
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "testdir/")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		err := f.CreateSelf()
		assert.NoError(t, err)

		assert.True(t, f.Exists())
		assert.True(t, f.IsDir())
	})
}

func TestWeblensFile_HierarchyOperations(t *testing.T) {
	testSetup(t)

	t.Run("ParentChildRelationships", func(t *testing.T) {
		parentPath := file_system.BuildFilePath(TestRoot, "parent/")
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path: parentPath,
		})
		err := parent.CreateSelf()
		require.NoError(t, err)

		childPath := file_system.BuildFilePath(TestRoot, "parent/child.txt")
		child := file.NewWeblensFile(file.NewFileOptions{
			Path: childPath,
		})

		err = child.SetParent(parent)
		assert.NoError(t, err)
		err = parent.AddChild(child)
		assert.NoError(t, err)

		assert.Equal(t, parent, child.GetParent())
		children := parent.GetChildren()
		assert.Contains(t, children, child)

		retrievedChild, err := parent.GetChild(child.Name())
		assert.NoError(t, err)
		assert.Equal(t, child, retrievedChild)
	})

	t.Run("RemoveChild", func(t *testing.T) {
		parentPath := file_system.BuildFilePath(TestRoot, "parent2/")
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path: parentPath,
		})
		err := parent.CreateSelf()
		require.NoError(t, err)

		childPath := file_system.BuildFilePath(TestRoot, "parent2/child.txt")
		child := file.NewWeblensFile(file.NewFileOptions{
			Path: childPath,
		})

		err = parent.AddChild(child)
		require.NoError(t, err)

		err = parent.RemoveChild(child.Name())
		assert.NoError(t, err)

		_, err = parent.GetChild(child.Name())
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrFileNotFound)
	})
}

func TestWeblensFile_ContentOperations(t *testing.T) {
	testSetup(t)

	t.Run("WriteAndRead", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "content.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		data := []byte("test content")
		n, err := f.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)

		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, data, content)

		reader, err := f.Readable()
		assert.NoError(t, err)

		buf := make([]byte, 4)

		n, err = reader.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, []byte("test"), buf)
	})

	t.Run("AppendContent", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "append.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		initial := []byte("initial ")
		_, err := f.Write(initial)
		require.NoError(t, err)

		appendData := []byte("appended")
		err = f.Append(appendData)
		assert.NoError(t, err)

		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, []byte("initial appended"), content)
	})

	t.Run("LargeFileOperations", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "large.bin")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		data := make([]byte, 10*1024*1024)
		_, err := rand.Read(data)
		require.NoError(t, err)

		n, err := f.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)

		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, data, content)
	})
}

func TestWeblensFile_ConcurrentOperations(t *testing.T) {
	testSetup(t)

	t.Run("ConcurrentReads", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "concurrent.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		data := []byte("test content for concurrent reading")
		_, err := f.Write(data)
		require.NoError(t, err)

		var wg sync.WaitGroup

		numGoroutines := 10
		wg.Add(numGoroutines)

		for range numGoroutines {
			go func() {
				defer wg.Done()

				content, err := f.ReadAll()
				assert.NoError(t, err)
				assert.Equal(t, data, content)
			}()
		}

		wg.Wait()
	})

	t.Run("ConcurrentWritesAndReads", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "concurrent_rw.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		err := f.CreateSelf()
		require.NoError(t, err)

		var wg sync.WaitGroup

		numWriters := 5
		numReaders := 5

		wg.Add(numWriters + numReaders)

		for i := range numWriters {
			go func(id int) {
				defer wg.Done()

				data := fmt.Appendf(nil, "writer_%d", id)

				err := f.Append(data)
				assert.NoError(t, err)
				time.Sleep(10 * time.Millisecond)
			}(i)
		}

		for range numReaders {
			go func() {
				defer wg.Done()

				_, err := f.ReadAll()
				assert.NoError(t, err)
				time.Sleep(10 * time.Millisecond)
			}()
		}

		wg.Wait()
	})
}

func TestWeblensFile_ErrorHandling(t *testing.T) {
	testSetup(t)

	t.Run("ReadNonExistentFile", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "nonexistent.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})

		_, err := f.ReadAll()
		assert.Error(t, err)
	})

	t.Run("WriteToDirectory", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "testdir/")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: path,
		})
		err := f.CreateSelf()
		require.NoError(t, err)

		_, err = f.Write([]byte("test"))
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrDirectoryNotAllowed)
	})

	t.Run("InvalidParentChild", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "parent/"),
		})
		child := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath("different", "path/child"),
		})

		err := child.SetParent(parent)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrNotChild)
	})
}

func TestWeblensFile_MemoryOperations(t *testing.T) {
	t.Run("MemOnlyFile", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:    file_system.BuildFilePath("memory", "memory.txt"),
			FileID:  "mem1",
			MemOnly: true,
		})

		data := []byte("memory only content")
		n, err := f.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)

		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, data, content)

		assert.False(t, f.Exists())
	})

	t.Run("MemOnlyWithBuffer", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:    file_system.BuildFilePath("memory", "buffer.txt"),
			FileID:  "mem2",
			MemOnly: true,
		})

		buf := bytes.NewBuffer(nil)
		data := []byte("buffer test")
		_, err := buf.Write(data)
		require.NoError(t, err)

		n, err := f.Write(buf.Bytes())
		assert.NoError(t, err)
		assert.Equal(t, buf.Len(), n)

		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, data, content)
	})
}

func TestWeblensFile_TreeOperations(t *testing.T) {
	testSetup(t)

	t.Run("RecursiveMap", func(t *testing.T) {
		root := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "root/"),
		})
		require.NoError(t, root.CreateSelf())

		files := make([]*file.WeblensFileImpl, 0)

		for i := range 3 {
			child := file.NewWeblensFile(file.NewFileOptions{
				Path: file_system.BuildFilePath(TestRoot, fmt.Sprintf("root/child_%d", i)),
			})
			require.NoError(t, child.SetParent(root))
			require.NoError(t, root.AddChild(child))
			files = append(files, child)
		}

		visited := make(map[string]bool)
		err := root.RecursiveMap(func(f *file.WeblensFileImpl) error {
			visited[f.Name()] = true

			return nil
		})

		assert.NoError(t, err)
		assert.True(t, visited["root"])

		for _, f := range files {
			assert.True(t, visited[f.Name()])
		}
	})

	t.Run("LeafMap", func(t *testing.T) {
		root := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "leaftest/"),
		})
		require.NoError(t, root.CreateSelf())

		subdir := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "leaftest/subdir/"),
		})
		require.NoError(t, subdir.CreateSelf())
		require.NoError(t, subdir.SetParent(root))
		require.NoError(t, root.AddChild(subdir))

		leafFiles := make([]*file.WeblensFileImpl, 0)

		for i := range 2 {
			leaf := file.NewWeblensFile(file.NewFileOptions{
				Path: file_system.BuildFilePath(TestRoot, fmt.Sprintf("leaftest/subdir/leaf_%d", i)),
			})
			require.NoError(t, leaf.SetParent(subdir))
			require.NoError(t, subdir.AddChild(leaf))
			leafFiles = append(leafFiles, leaf)
		}

		visited := make([]string, 0)
		err := root.LeafMap(func(f *file.WeblensFileImpl) error {
			visited = append(visited, f.Name())

			return nil
		})
		assert.NoError(t, err)

		expectedOrder := []string{
			"leaf_0",
			"leaf_1",
			"subdir",
			"leaftest",
		}
		assert.Equal(t, expectedOrder, visited, "Files should be visited in depth-first order")
	})
}

func TestWeblensFile_Identity(t *testing.T) {
	testSetup(t)

	t.Run("ID and SetID", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:   file_system.BuildFilePath(TestRoot, "identity.txt"),
			FileID: "initial-id",
		})

		assert.Equal(t, "initial-id", f.ID())

		f.SetID("new-id")
		assert.Equal(t, "new-id", f.ID())
	})

	t.Run("Name returns filename", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "subdir/myfile.txt"),
		})

		assert.Equal(t, "myfile.txt", f.Name())
	})

	t.Run("Name returns directory name", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "subdir/mydir/"),
		})

		assert.Equal(t, "mydir", f.Name())
	})

	t.Run("GetPortablePath and SetPortablePath", func(t *testing.T) {
		originalPath := file_system.BuildFilePath(TestRoot, "original.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: originalPath,
		})

		assert.Equal(t, originalPath, f.GetPortablePath())

		newPath := file_system.BuildFilePath(TestRoot, "renamed.txt")
		f.SetPortablePath(newPath)
		assert.Equal(t, newPath, f.GetPortablePath())
	})
}

func TestWeblensFile_FileInfo(t *testing.T) {
	testSetup(t)

	t.Run("Mode returns 0", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "mode.txt"),
		})

		assert.Equal(t, os.FileMode(0), f.Mode())
	})

	t.Run("Sys returns nil", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "sys.txt"),
		})

		assert.Nil(t, f.Sys())
	})

	// t.Run("ModTime returns modification time", func(t *testing.T) {
	// 	modTime := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	// 	f := file.NewWeblensFile(file.NewFileOptions{
	// 		Path: file_system.BuildFilePath(TestRoot, "modtime.txt"),
	// 		ModTime:  modTime,
	// 	})
	//
	// 	// For memory-only files that are past files, it should return the set mod time
	// 	f.SetPastFile(true)
	// 	assert.Equal(t, modTime, f.ModTime())
	// })

	t.Run("Stat returns file info", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:   file_system.BuildFilePath(TestRoot, "stat.txt"),
			FileID: "stat-id",
		})

		info, err := f.Stat()
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "stat.txt", info.Name())
	})

	t.Run("Size and SetSize", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "size.txt"),
			Size: 100,
		})

		assert.Equal(t, int64(100), f.Size())

		f.SetSize(200)
		assert.Equal(t, int64(200), f.Size())
	})

	// t.Run("Size for memory-only past file", func(t *testing.T) {
	// 	f := file.NewTestFile(file.TestFileOptions{
	// 		RootName:   TestRoot,
	// 		RelPath:    "pastsize.txt",
	// 		Size:       500,
	// 		IsPastFile: true,
	// 	})
	//
	// 	assert.Equal(t, int64(500), f.Size())
	// })

	t.Run("LoadStat for existing file", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "loadstat.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		data := []byte("test content for loadstat")
		_, err := f.Write(data)
		require.NoError(t, err)

		// Reset size to force reload
		f.SetSize(-1)

		newSize, err := f.LoadStat()
		assert.NoError(t, err)
		assert.Equal(t, int64(len(data)), newSize)
	})

	t.Run("LoadStat for non-existent file returns error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "nonexistent_loadstat.txt"),
		})

		_, err := f.LoadStat()
		assert.Error(t, err)
	})
}

func TestWeblensFile_IOOperations(t *testing.T) {
	testSetup(t)

	t.Run("Read reads into buffer", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "read.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		data := []byte("hello world")
		_, err := f.Write(data)
		require.NoError(t, err)

		buf := make([]byte, 5)
		n, err := f.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, []byte("hello"), buf)
	})

	t.Run("Read on memory-only file", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:    file_system.BuildFilePath("memory", "memread.txt"),
			FileID:  "memread-id",
			MemOnly: true,
		})

		data := []byte("memory content")
		_, err := f.Write(data)
		require.NoError(t, err)

		buf := make([]byte, 6)
		_, err = f.Read(buf)
		assert.NoError(t, err)
	})

	t.Run("Close returns nil", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "close.txt"),
		})

		err := f.Close()
		assert.NoError(t, err)
	})

	t.Run("Readdir returns children", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "readdir/"),
			CreateNow: true,
		})

		child1 := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "readdir/child1.txt"),
			CreateNow: true,
		})
		child2 := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "readdir/child2.txt"),
			CreateNow: true,
		})

		require.NoError(t, parent.AddChild(child1))
		require.NoError(t, parent.AddChild(child2))

		entries, err := parent.Readdir(0)
		assert.NoError(t, err)
		assert.Len(t, entries, 2)
	})

	t.Run("Readdir with count limit", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "readdir_limit/"),
			CreateNow: true,
		})

		for i := range 5 {
			child := file.NewWeblensFile(file.NewFileOptions{
				Path:      file_system.BuildFilePath(TestRoot, fmt.Sprintf("readdir_limit/child%d.txt", i)),
				CreateNow: true,
			})
			require.NoError(t, parent.AddChild(child))
		}

		entries, err := parent.Readdir(2)
		assert.NoError(t, err)
		assert.Len(t, entries, 2)
	})

	t.Run("Seek with SeekStart", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "seek.txt"),
			Size: 100,
		})

		pos, err := f.Seek(50, io.SeekStart)
		assert.NoError(t, err)
		assert.Equal(t, int64(50), pos)
	})

	t.Run("Seek with SeekCurrent", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "seek_current.txt"),
			Size: 100,
		})

		// First seek to position 30
		_, err := f.Seek(30, io.SeekStart)
		require.NoError(t, err)

		// Now seek 20 more from current
		pos, err := f.Seek(20, io.SeekCurrent)
		assert.NoError(t, err)
		assert.Equal(t, int64(50), pos)
	})

	t.Run("Seek with SeekEnd", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "seek_end.txt"),
			Size: 100,
		})

		pos, err := f.Seek(10, io.SeekEnd)
		assert.NoError(t, err)
		assert.Equal(t, int64(90), pos) // 100 - 10 = 90
	})

	t.Run("Writer returns write closer", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "writer.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		writer, err := f.Writer()
		assert.NoError(t, err)
		assert.NotNil(t, writer)

		n, err := writer.Write([]byte("written via writer"))
		assert.NoError(t, err)
		assert.Equal(t, 18, n)

		err = writer.Close()
		assert.NoError(t, err)

		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, []byte("written via writer"), content)
	})

	t.Run("Writer on directory returns error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "writer_dir/"),
			CreateNow: true,
		})

		_, err := f.Writer()
		assert.Error(t, err)
	})

	t.Run("Readable on directory returns error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "readable_dir/"),
			CreateNow: true,
		})

		_, err := f.Readable()
		assert.Error(t, err)
	})
}

func TestWeblensFile_StateOperations(t *testing.T) {
	testSetup(t)

	t.Run("SetMemOnly", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:   file_system.BuildFilePath(TestRoot, "memonly.txt"),
			FileID: "memonly-id",
		})

		// Initially not memory-only
		assert.True(t, f.Exists() == false) // File not created yet

		f.SetMemOnly(true)

		// Write should only go to buffer
		_, err := f.Write([]byte("memory content"))
		assert.NoError(t, err)

		// Should not exist on disk
		assert.False(t, f.Exists())
	})

	t.Run("WithLock executes function under lock", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "withlock.txt"),
			CreateNow: true,
		})

		executed := false
		err := f.WithLock(func() error {
			executed = true

			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("WithLock propagates error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "withlock_err.txt"),
			CreateNow: true,
		})

		expectedErr := fmt.Errorf("test error")
		err := f.WithLock(func() error {
			return expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("SetWatching marks file as watched", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "watching.txt"),
		})

		err := f.SetWatching()
		assert.NoError(t, err)
	})

	t.Run("IsReadOnly returns read-only status", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "readonly.txt"),
		})

		// Default should be not read-only
		assert.False(t, f.IsReadOnly())
	})

	t.Run("Freeze creates shallow copy", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "freeze.txt"),
			FileID:    "freeze-id",
			ContentID: "content-123",
			Size:      500,
		})

		frozen := f.Freeze()

		assert.Equal(t, f.ID(), frozen.ID())
		assert.Equal(t, f.GetContentID(), frozen.GetContentID())
		assert.Equal(t, f.Size(), frozen.Size())
		assert.Equal(t, f.GetPortablePath(), frozen.GetPortablePath())

		// Verify it's a different instance
		f.SetID("modified-id")
		assert.NotEqual(t, f.ID(), frozen.ID())
	})
}

func TestWeblensFile_ContentID(t *testing.T) {
	testSetup(t)

	t.Run("GetContentID and SetContentID", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "contentid.txt"),
			ContentID: "initial-content-id",
		})

		assert.Equal(t, "initial-content-id", f.GetContentID())

		f.SetContentID("new-content-id")
		assert.Equal(t, "new-content-id", f.GetContentID())
	})

	t.Run("GenerateContentID for file", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "generateid.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		data := []byte("content for hashing")
		_, err := f.Write(data)
		require.NoError(t, err)

		ctx := t.Context()
		contentID, err := file.GenerateContentID(ctx, f)
		assert.NoError(t, err)
		assert.NotEmpty(t, contentID)
		assert.Equal(t, 20, len(contentID))

		// Calling again should return same ID
		contentID2, err := file.GenerateContentID(ctx, f)
		assert.NoError(t, err)
		assert.Equal(t, contentID, contentID2)
	})

	t.Run("GenerateContentID for directory returns error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "contentid_dir/"),
			CreateNow: true,
		})

		ctx := t.Context()
		_, err := file.GenerateContentID(ctx, f)
		assert.Error(t, err)
	})

	t.Run("GenerateContentID for empty file returns error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "empty_contentid.txt"),
			CreateNow: true,
		})

		ctx := t.Context()
		_, err := file.GenerateContentID(ctx, f)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrEmptyFile)
	})
}

func TestWeblensFile_HierarchyExtended(t *testing.T) {
	testSetup(t)

	t.Run("ChildrenLoaded returns false for uninitialized", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "unloaded_dir/"),
		})

		assert.False(t, f.ChildrenLoaded())
	})

	t.Run("ChildrenLoaded returns true after InitChildren", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "loaded_dir/"),
			CreateNow: true,
		})

		f.InitChildren()
		assert.True(t, f.ChildrenLoaded())
	})

	t.Run("InitChildren initializes map", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "init_children_dir/"),
			CreateNow: true,
		})

		f.InitChildren()

		// Should now be able to add children
		child := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "init_children_dir/child.txt"),
		})

		err := f.AddChild(child)
		assert.NoError(t, err)
	})

	t.Run("BubbleMap traverses up tree", func(t *testing.T) {
		grandparent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "grandparent/"),
			CreateNow: true,
		})

		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "grandparent/parent/"),
			CreateNow: true,
		})
		require.NoError(t, parent.SetParent(grandparent))
		require.NoError(t, grandparent.AddChild(parent))

		child := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "grandparent/parent/child.txt"),
			CreateNow: true,
		})
		require.NoError(t, child.SetParent(parent))
		require.NoError(t, parent.AddChild(child))

		visited := make([]string, 0)
		err := child.BubbleMap(func(f *file.WeblensFileImpl) error {
			visited = append(visited, f.Name())

			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, []string{"child.txt", "parent", "grandparent"}, visited)
	})

	t.Run("BubbleMap on nil file returns nil", func(t *testing.T) {
		var f *file.WeblensFileImpl

		err := f.BubbleMap(func(_ *file.WeblensFileImpl) error {
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("BubbleMap propagates error", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "bubble_err_parent/"),
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "bubble_err_parent/child.txt"),
			CreateNow: true,
		})
		require.NoError(t, child.SetParent(parent))

		expectedErr := fmt.Errorf("bubble error")
		err := child.BubbleMap(func(f *file.WeblensFileImpl) error {
			if f.Name() == "bubble_err_parent" {
				return expectedErr
			}

			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("IsParentOf returns true for child", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "isparent/"),
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "isparent/child.txt"),
			CreateNow: true,
		})

		assert.True(t, parent.IsParentOf(child))
	})

	// t.Run("IsParentOf returns true for grandchild", func(t *testing.T) {
	// 	// Using test files that don't need to exist on disk
	// 	grandparent := file.NewTestFile(file.TestFileOptions{
	// 		RootName: TestRoot,
	// 		RelPath:  "isgrandparent/",
	// 		IsDir:    true,
	// 	})
	//
	// 	grandchild := file.NewTestFile(file.TestFileOptions{
	// 		RootName: TestRoot,
	// 		RelPath:  "isgrandparent/parent/child.txt",
	// 	})
	//
	// 	assert.True(t, grandparent.IsParentOf(grandchild))
	// })

	// t.Run("IsParentOf returns false for unrelated files", func(t *testing.T) {
	// 	// Using test files that don't need to exist on disk
	// 	f1 := file.NewTestFile(file.TestFileOptions{
	// 		RootName: TestRoot,
	// 		RelPath:  "unrelated1/",
	// 		IsDir:    true,
	// 	})
	//
	// 	f2 := file.NewTestFile(file.TestFileOptions{
	// 		RootName: TestRoot,
	// 		RelPath:  "unrelated2/file.txt",
	// 	})
	//
	// 	assert.False(t, f1.IsParentOf(f2))
	// })

	// t.Run("IsParentOf returns false for different roots", func(t *testing.T) {
	// 	f1 := file.NewTestFile(file.TestFileOptions{
	// 		RootName: "root1",
	// 		RelPath:  "parent/",
	// 		IsDir:    true,
	// 	})
	//
	// 	f2 := file.NewTestFile(file.TestFileOptions{
	// 		RootName: "root2",
	// 		RelPath:  "parent/child.txt",
	// 	})
	//
	// 	assert.False(t, f1.IsParentOf(f2))
	// })
}

func TestWeblensFile_Lifecycle(t *testing.T) {
	testSetup(t)

	t.Run("Remove deletes file", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "remove_file.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		_, err := f.Write([]byte("content"))
		require.NoError(t, err)
		assert.True(t, f.Exists())

		err = f.Remove()
		assert.NoError(t, err)
		assert.False(t, f.Exists())
	})

	t.Run("Remove deletes directory", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "remove_dir/")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		assert.True(t, f.Exists())

		err := f.Remove()
		assert.NoError(t, err)
		assert.False(t, f.Exists())
	})

	t.Run("Remove deletes directory with contents", func(t *testing.T) {
		parentPath := file_system.BuildFilePath(TestRoot, "remove_dir_contents/")
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      parentPath,
			CreateNow: true,
		})

		childPath := file_system.BuildFilePath(TestRoot, "remove_dir_contents/child.txt")
		child := file.NewWeblensFile(file.NewFileOptions{
			Path:      childPath,
			CreateNow: true,
		})
		_, err := child.Write([]byte("child content"))
		require.NoError(t, err)

		assert.True(t, parent.Exists())
		assert.True(t, child.Exists())

		err = parent.Remove()
		assert.NoError(t, err)
		assert.False(t, parent.Exists())
		assert.False(t, child.Exists())
	})

	t.Run("SetModifiedTime updates modification time", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "setmodtime.txt"),
		})

		newTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
		f.SetModifiedTime(newTime)

		// For past files, ModTime returns the set time directly
		f.SetPastFile(true)
		assert.Equal(t, newTime, f.ModTime())
	})

	// t.Run("ReplaceRoot changes root alias", func(t *testing.T) {
	// 	f := file.NewTestFile(file.TestFileOptions{
	// 		RootName: "oldroot",
	// 		RelPath:  "somefile.txt",
	// 	})
	//
	// 	assert.Equal(t, "oldroot", f.GetPortablePath().RootName())
	//
	// 	f.ReplaceRoot("newroot")
	// 	assert.Equal(t, "newroot", f.GetPortablePath().RootName())
	// })

	t.Run("SetPastFile and IsPastFile", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "pastfile.txt"),
		})

		assert.False(t, f.IsPastFile())

		f.SetPastFile(true)
		assert.True(t, f.IsPastFile())

		f.SetPastFile(false)
		assert.False(t, f.IsPastFile())
	})
}

func TestWeblensFile_JSONSerialization(t *testing.T) {
	testSetup(t)

	t.Run("MarshalJSON produces valid JSON", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "json_parent/"),
			FileID:    "parent-id",
			CreateNow: true,
		})

		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "json_parent/marshal.txt"),
			FileID:    "file-id",
			ContentID: "content-id",
			Size:      100,
			CreateNow: true,
		})
		require.NoError(t, f.SetParent(parent))

		data, err := f.MarshalJSON()
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify it's valid JSON by unmarshaling into a map
		var result map[string]any

		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)

		assert.Equal(t, float64(100), result["size"])
		assert.Equal(t, false, result["isDir"])
		assert.Equal(t, "content-id", result["contentID"])
		assert.Equal(t, "parent-id", result["parentID"])
	})

	t.Run("MarshalJSON for directory", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "json_dir/"),
			FileID:    "dir-id",
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path:   file_system.BuildFilePath(TestRoot, "json_dir/child.txt"),
			FileID: "child-id",
		})
		require.NoError(t, f.AddChild(child))

		data, err := f.MarshalJSON()
		assert.NoError(t, err)

		var result map[string]any

		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)

		assert.Equal(t, true, result["isDir"])
		childrenIDs := result["childrenIds"].([]any)
		assert.Len(t, childrenIDs, 1)
		assert.Equal(t, "child-id", childrenIDs[0])
	})

	t.Run("UnmarshalJSON restores file", func(t *testing.T) {
		// Create a JSON structure that matches what UnmarshalJSON expects
		jsonData := `{
			"portablePath": "test://unmarshal.txt",
			"size": 250,
			"isDir": false,
			"modifyTimestamp": 1704067200000,
			"contentID": "unmarshal-content",
			"childrenIDs": []
		}`

		restored := &file.WeblensFileImpl{}
		err := restored.UnmarshalJSON([]byte(jsonData))
		assert.NoError(t, err)

		assert.Equal(t, int64(250), restored.Size())
		assert.Equal(t, "unmarshal-content", restored.GetContentID())
		assert.False(t, restored.IsDir())
	})

	t.Run("UnmarshalJSON for directory", func(t *testing.T) {
		// Create a JSON structure that matches what UnmarshalJSON expects
		jsonData := `{
			"portablePath": "test://unmarshal_dir/",
			"size": 0,
			"isDir": true,
			"modifyTimestamp": 1704067200000,
			"contentID": "",
			"childrenIDs": ["child1", "child2"]
		}`

		restored := &file.WeblensFileImpl{}
		err := restored.UnmarshalJSON([]byte(jsonData))
		assert.NoError(t, err)

		assert.True(t, restored.IsDir())
	})

	t.Run("UnmarshalJSON with invalid JSON returns error", func(t *testing.T) {
		f := &file.WeblensFileImpl{}
		err := f.UnmarshalJSON([]byte("invalid json"))
		assert.Error(t, err)
	})
}

func TestWeblensFile_DirectorySize(t *testing.T) {
	testSetup(t)

	t.Run("Size calculates directory size from children", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "dirsize/"),
			CreateNow: true,
		})

		child1 := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "dirsize/child1.txt"),
			CreateNow: true,
		})
		_, err := child1.Write([]byte("12345"))
		require.NoError(t, err)
		require.NoError(t, child1.SetParent(parent))
		require.NoError(t, parent.AddChild(child1))

		child2 := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "dirsize/child2.txt"),
			CreateNow: true,
		})
		_, err = child2.Write([]byte("1234567890"))
		require.NoError(t, err)
		require.NoError(t, child2.SetParent(parent))
		require.NoError(t, parent.AddChild(child2))

		// Directory size should be sum of children
		size := parent.Size()
		assert.Equal(t, int64(15), size)
	})
}

func TestWeblensFile_AddChildErrors(t *testing.T) {
	testSetup(t)

	t.Run("AddChild to non-directory returns error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "notadir.txt"),
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "notadir.txt/child.txt"),
		})

		err := f.AddChild(child)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrDirectoryRequired)
	})

	t.Run("AddChild duplicate returns error", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "dupparent/"),
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "dupparent/child.txt"),
		})

		err := parent.AddChild(child)
		require.NoError(t, err)

		// Add same child again
		err = parent.AddChild(child)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrFileAlreadyExists)
	})
}

func TestWeblensFile_RemoveChildErrors(t *testing.T) {
	testSetup(t)

	t.Run("RemoveChild from empty directory returns error", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "empty_remove/"),
			CreateNow: true,
		})

		err := f.RemoveChild("nonexistent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrNoChildren)
	})

	t.Run("RemoveChild nonexistent returns error", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "remove_nonexistent/"),
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "remove_nonexistent/exists.txt"),
		})
		require.NoError(t, parent.AddChild(child))

		err := parent.RemoveChild("nonexistent.txt")
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrFileNotFound)
	})
}

func TestWeblensFile_GetChildErrors(t *testing.T) {
	testSetup(t)

	t.Run("GetChild with empty name returns error", func(t *testing.T) {
		parent := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "getchild_empty/"),
			CreateNow: true,
		})
		parent.InitChildren()

		_, err := parent.GetChild("")
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrFileNotFound)
	})
}

func TestWeblensFile_RecursiveMapError(t *testing.T) {
	testSetup(t)

	t.Run("RecursiveMap propagates error", func(t *testing.T) {
		root := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "recursive_err/"),
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "recursive_err/child.txt"),
		})
		require.NoError(t, root.AddChild(child))

		expectedErr := fmt.Errorf("recursive error")
		err := root.RecursiveMap(func(f *file.WeblensFileImpl) error {
			if f.Name() == "child.txt" {
				return expectedErr
			}

			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestWeblensFile_LeafMapError(t *testing.T) {
	testSetup(t)

	t.Run("LeafMap on nil returns error", func(t *testing.T) {
		var f *file.WeblensFileImpl

		err := f.LeafMap(func(_ *file.WeblensFileImpl) error {
			return nil
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrNilFile)
	})

	t.Run("LeafMap propagates error", func(t *testing.T) {
		root := file.NewWeblensFile(file.NewFileOptions{
			Path:      file_system.BuildFilePath(TestRoot, "leafmap_err/"),
			CreateNow: true,
		})

		child := file.NewWeblensFile(file.NewFileOptions{
			Path: file_system.BuildFilePath(TestRoot, "leafmap_err/child.txt"),
		})
		require.NoError(t, root.AddChild(child))

		expectedErr := fmt.Errorf("leaf error")
		err := root.LeafMap(func(f *file.WeblensFileImpl) error {
			if f.Name() == "child.txt" {
				return expectedErr
			}

			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestWeblensFile_WriteAt(t *testing.T) {
	testSetup(t)

	t.Run("BasicWriteAt", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "writeat.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		// Write initial content
		initialData := []byte("Hello World")
		n, err := f.Write(initialData)
		require.NoError(t, err)
		require.Equal(t, len(initialData), n)

		// Write at specific position
		newData := []byte("NEW")
		err = f.WriteAt(newData, 6) // Position after "Hello "
		assert.NoError(t, err)

		// Verify content
		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, []byte("Hello NEWld"), content)
	})

	t.Run("WriteAtZeroPosition", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "writeat_zero.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		// Write initial content
		initialData := []byte("Hello World")
		n, err := f.Write(initialData)
		require.NoError(t, err)
		require.Equal(t, len(initialData), n)

		// Write at beginning
		newData := []byte("Start")
		err = f.WriteAt(newData, 0)
		assert.NoError(t, err)

		// Verify content
		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, []byte("Start World"), content)
	})

	t.Run("WriteAtBeyondFileSize", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "writeat_beyond.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		// Write initial content
		initialData := []byte("Short")
		n, err := f.Write(initialData)
		require.NoError(t, err)
		require.Equal(t, len(initialData), n)

		// Write beyond current file size
		newData := []byte("Extended")
		err = f.WriteAt(newData, 10)
		assert.NoError(t, err)

		// Verify content
		content, err := f.ReadAll()
		assert.NoError(t, err)
		// Should have null bytes between "Short" and "Extended"
		expected := append([]byte("Short"), make([]byte, 5)...) // 5 null bytes
		expected = append(expected, []byte("Extended")...)
		assert.Equal(t, expected, content)
	})

	t.Run("ConcurrentWriteAt", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "writeat_concurrent.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})

		// Write initial content
		initialData := []byte("0000000000") // 10 zeros
		n, err := f.Write(initialData)
		require.NoError(t, err)
		require.Equal(t, len(initialData), n)

		var wg sync.WaitGroup

		numGoroutines := 5

		// Concurrently write at different positions
		wg.Add(numGoroutines)

		for i := range numGoroutines {
			go func(pos int) {
				defer wg.Done()

				data := fmt.Appendf(nil, "%d", pos)

				err := f.WriteAt(data, int64(pos))
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// Verify content
		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, 10, len(content))

		// Verify that first 5 positions contain digits 0-4
		for i := range numGoroutines {
			assert.Equal(t, byte('0'+i), content[i])
		}
	})

	t.Run("WriteAtToDirectory", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "writeat_dir/")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})
		require.NotNil(t, f)

		// Attempt to write to directory
		err := f.WriteAt([]byte("test"), 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, file.ErrDirectoryNotAllowed)
	})

	t.Run("MemOnlyWriteAt", func(t *testing.T) {
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:    file_system.BuildFilePath("memory", "writeat_mem.txt"),
			FileID:  "mem_writeat",
			MemOnly: true,
		})
		require.NotNil(t, f)

		// Write initial content
		initialData := []byte("Hello World")
		n, err := f.Write(initialData)
		require.NoError(t, err)
		require.Equal(t, len(initialData), n)

		// Write at specific position
		newData := []byte("NEW")
		err = f.WriteAt(newData, 6)
		assert.NoError(t, err)

		// Verify content
		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, []byte("Hello NEWld"), content)
		assert.False(t, f.Exists()) // Memory-only file should not exist on disk
	})

	t.Run("LargeWriteAt", func(t *testing.T) {
		path := file_system.BuildFilePath(TestRoot, "writeat_large.txt")
		f := file.NewWeblensFile(file.NewFileOptions{
			Path:      path,
			CreateNow: true,
		})
		require.NotNil(t, f)

		// Write initial large content
		initialSize := 1024 * 1024 // 1MB

		initialData := make([]byte, initialSize)
		for i := range initialData {
			initialData[i] = 'A'
		}

		n, err := f.Write(initialData)
		require.NoError(t, err)
		require.Equal(t, initialSize, n)

		// Write large chunk at middle
		writeSize := 512 * 1024 // 512KB

		writeData := make([]byte, writeSize)
		for i := range writeData {
			writeData[i] = 'B'
		}

		err = f.WriteAt(writeData, int64(initialSize/2))
		assert.NoError(t, err)

		// Verify content
		content, err := f.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, initialSize, len(content))

		// Verify first quarter is 'A's
		for i := 0; i < initialSize/4; i++ {
			assert.Equal(t, byte('A'), content[i])
		}

		// Verify middle half is 'B's
		for i := initialSize / 2; i < initialSize/2+writeSize; i++ {
			assert.Equal(t, byte('B'), content[i])
		}
	})
}
