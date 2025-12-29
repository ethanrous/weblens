package file_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
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
