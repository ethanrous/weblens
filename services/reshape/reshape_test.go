package reshape_test

// ============================================================================
// journal.go tests
// ============================================================================

// func TestFileActionToFileActionInfo(t *testing.T) {
// 	now := time.Now()
// 	filepath := fs.BuildFilePath("USERS", "testuser/file.txt")
// 	originPath := fs.BuildFilePath("USERS", "testuser/old/file.txt")
// 	destPath := fs.BuildFilePath("USERS", "testuser/new/file.txt")
//
// 	tests := []struct {
// 		name     string
// 		action   history.FileAction
// 		expected structs.FileActionInfo
// 	}{
// 		{
// 			name: "basic file create action",
// 			action: history.FileAction{
// 				ActionType: history.FileCreate,
// 				FileID:     "file123",
// 				Filepath:   filepath,
// 				EventID:    "event456",
// 				TowerID:    "tower789",
// 				Timestamp:  now,
// 				Size:       1024,
// 				ContentID:  "content123",
// 			},
// 			expected: structs.FileActionInfo{
// 				ActionType:      history.FileCreate,
// 				FileID:          "file123",
// 				Filepath:        filepath.ToPortable(),
// 				EventID:         "event456",
// 				TowerID:         "tower789",
// 				Timestamp:       now.UnixMilli(),
// 				Size:            1024,
// 				ContentID:       "content123",
// 				OriginPath:      "",
// 				DestinationPath: "",
// 			},
// 		},
// 		{
// 			name: "file move action with origin and destination",
// 			action: history.FileAction{
// 				ActionType:      history.FileMove,
// 				FileID:          "file456",
// 				Filepath:        filepath,
// 				OriginPath:      originPath,
// 				DestinationPath: destPath,
// 				EventID:         "event789",
// 				TowerID:         "tower123",
// 				Timestamp:       now,
// 				Size:            2048,
// 				ContentID:       "content456",
// 			},
// 			expected: structs.FileActionInfo{
// 				ActionType:      history.FileMove,
// 				FileID:          "file456",
// 				Filepath:        filepath.ToPortable(),
// 				OriginPath:      originPath.ToPortable(),
// 				DestinationPath: destPath.ToPortable(),
// 				EventID:         "event789",
// 				TowerID:         "tower123",
// 				Timestamp:       now.UnixMilli(),
// 				Size:            2048,
// 				ContentID:       "content456",
// 			},
// 		},
// 		{
// 			name: "file delete action",
// 			action: history.FileAction{
// 				ActionType: history.FileDelete,
// 				FileID:     "file789",
// 				Filepath:   filepath,
// 				EventID:    "event123",
// 				TowerID:    "tower456",
// 				Timestamp:  now,
// 				Size:       0,
// 			},
// 			expected: structs.FileActionInfo{
// 				ActionType:      history.FileDelete,
// 				FileID:          "file789",
// 				Filepath:        filepath.ToPortable(),
// 				EventID:         "event123",
// 				TowerID:         "tower456",
// 				Timestamp:       now.UnixMilli(),
// 				Size:            0,
// 				OriginPath:      "",
// 				DestinationPath: "",
// 			},
// 		},
// 		{
// 			name: "zero timestamp",
// 			action: history.FileAction{
// 				ActionType: history.FileCreate,
// 				FileID:     "file000",
// 				Filepath:   filepath,
// 				EventID:    "event000",
// 				TowerID:    "tower000",
// 				Timestamp:  time.Time{},
// 				Size:       512,
// 			},
// 			expected: structs.FileActionInfo{
// 				ActionType: history.FileCreate,
// 				FileID:     "file000",
// 				Filepath:   filepath.ToPortable(),
// 				EventID:    "event000",
// 				TowerID:    "tower000",
// 				Timestamp:  time.Time{}.UnixMilli(),
// 				Size:       512,
// 			},
// 		},
// 		{
// 			name: "empty paths",
// 			action: history.FileAction{
// 				ActionType: history.FileCreate,
// 				FileID:     "fileEmpty",
// 				Filepath:   fs.Filepath{},
// 				EventID:    "eventEmpty",
// 				TowerID:    "towerEmpty",
// 				Timestamp:  now,
// 				Size:       0,
// 			},
// 			expected: structs.FileActionInfo{
// 				ActionType: history.FileCreate,
// 				FileID:     "fileEmpty",
// 				Filepath:   "",
// 				EventID:    "eventEmpty",
// 				TowerID:    "towerEmpty",
// 				Timestamp:  now.UnixMilli(),
// 				Size:       0,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.FileActionToFileActionInfo(tt.action)
//
// 			if result.ActionType != tt.expected.ActionType {
// 				t.Errorf("ActionType = %v, want %v", result.ActionType, tt.expected.ActionType)
// 			}
// 			if result.FileID != tt.expected.FileID {
// 				t.Errorf("FileID = %v, want %v", result.FileID, tt.expected.FileID)
// 			}
// 			if result.Filepath != tt.expected.Filepath {
// 				t.Errorf("Filepath = %v, want %v", result.Filepath, tt.expected.Filepath)
// 			}
// 			if result.OriginPath != tt.expected.OriginPath {
// 				t.Errorf("OriginPath = %v, want %v", result.OriginPath, tt.expected.OriginPath)
// 			}
// 			if result.DestinationPath != tt.expected.DestinationPath {
// 				t.Errorf("DestinationPath = %v, want %v", result.DestinationPath, tt.expected.DestinationPath)
// 			}
// 			if result.EventID != tt.expected.EventID {
// 				t.Errorf("EventID = %v, want %v", result.EventID, tt.expected.EventID)
// 			}
// 			if result.TowerID != tt.expected.TowerID {
// 				t.Errorf("TowerID = %v, want %v", result.TowerID, tt.expected.TowerID)
// 			}
// 			if result.Timestamp != tt.expected.Timestamp {
// 				t.Errorf("Timestamp = %v, want %v", result.Timestamp, tt.expected.Timestamp)
// 			}
// 			if result.Size != tt.expected.Size {
// 				t.Errorf("Size = %v, want %v", result.Size, tt.expected.Size)
// 			}
// 			if result.ContentID != tt.expected.ContentID {
// 				t.Errorf("ContentID = %v, want %v", result.ContentID, tt.expected.ContentID)
// 			}
// 		})
// 	}
// }
//
// func TestFileActionInfoToFileAction(t *testing.T) {
// 	now := time.Now().Truncate(time.Millisecond)
//
// 	tests := []struct {
// 		name     string
// 		info     structs.FileActionInfo
// 		expected history.FileAction
// 	}{
// 		{
// 			name: "basic file create action",
// 			info: structs.FileActionInfo{
// 				ActionType: history.FileCreate,
// 				FileID:     "file123",
// 				Filepath:   "USERS:testuser/file.txt",
// 				EventID:    "event456",
// 				TowerID:    "tower789",
// 				Timestamp:  now.UnixMilli(),
// 				Size:       1024,
// 				ContentID:  "content123",
// 			},
// 			expected: history.FileAction{
// 				ActionType: history.FileCreate,
// 				FileID:     "file123",
// 				Filepath:   fs.BuildFilePath("USERS", "testuser/file.txt"),
// 				EventID:    "event456",
// 				TowerID:    "tower789",
// 				Timestamp:  now,
// 				Size:       1024,
// 				ContentID:  "content123",
// 			},
// 		},
// 		{
// 			name: "file move action with paths",
// 			info: structs.FileActionInfo{
// 				ActionType:      history.FileMove,
// 				FileID:          "file456",
// 				Filepath:        "USERS:testuser/file.txt",
// 				OriginPath:      "USERS:testuser/old/file.txt",
// 				DestinationPath: "USERS:testuser/new/file.txt",
// 				EventID:         "event789",
// 				TowerID:         "tower123",
// 				Timestamp:       now.UnixMilli(),
// 				Size:            2048,
// 				ContentID:       "content456",
// 			},
// 			expected: history.FileAction{
// 				ActionType:      history.FileMove,
// 				FileID:          "file456",
// 				Filepath:        fs.BuildFilePath("USERS", "testuser/file.txt"),
// 				OriginPath:      fs.BuildFilePath("USERS", "testuser/old/file.txt"),
// 				DestinationPath: fs.BuildFilePath("USERS", "testuser/new/file.txt"),
// 				EventID:         "event789",
// 				TowerID:         "tower123",
// 				Timestamp:       now,
// 				Size:            2048,
// 				ContentID:       "content456",
// 			},
// 		},
// 		{
// 			name: "zero timestamp",
// 			info: structs.FileActionInfo{
// 				ActionType: history.FileCreate,
// 				FileID:     "file000",
// 				Filepath:   "USERS:test.txt",
// 				EventID:    "event000",
// 				TowerID:    "tower000",
// 				Timestamp:  0,
// 				Size:       512,
// 			},
// 			expected: history.FileAction{
// 				ActionType: history.FileCreate,
// 				FileID:     "file000",
// 				Filepath:   fs.BuildFilePath("USERS", "test.txt"),
// 				EventID:    "event000",
// 				TowerID:    "tower000",
// 				Timestamp:  time.UnixMilli(0),
// 				Size:       512,
// 			},
// 		},
// 		{
// 			name: "negative timestamp",
// 			info: structs.FileActionInfo{
// 				ActionType: history.FileCreate,
// 				FileID:     "fileNeg",
// 				Filepath:   "USERS:test.txt",
// 				EventID:    "eventNeg",
// 				TowerID:    "towerNeg",
// 				Timestamp:  -1000,
// 				Size:       256,
// 			},
// 			expected: history.FileAction{
// 				ActionType: history.FileCreate,
// 				FileID:     "fileNeg",
// 				Filepath:   fs.BuildFilePath("USERS", "test.txt"),
// 				EventID:    "eventNeg",
// 				TowerID:    "towerNeg",
// 				Timestamp:  time.UnixMilli(-1000),
// 				Size:       256,
// 			},
// 		},
// 		{
// 			name: "empty paths - invalid portable path",
// 			info: structs.FileActionInfo{
// 				ActionType: history.FileCreate,
// 				FileID:     "fileEmpty",
// 				Filepath:   "",
// 				EventID:    "eventEmpty",
// 				TowerID:    "towerEmpty",
// 				Timestamp:  now.UnixMilli(),
// 				Size:       0,
// 			},
// 			expected: history.FileAction{
// 				ActionType: history.FileCreate,
// 				FileID:     "fileEmpty",
// 				Filepath:   fs.Filepath{},
// 				EventID:    "eventEmpty",
// 				TowerID:    "towerEmpty",
// 				Timestamp:  now,
// 				Size:       0,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.FileActionInfoToFileAction(tt.info)
//
// 			if result.ActionType != tt.expected.ActionType {
// 				t.Errorf("ActionType = %v, want %v", result.ActionType, tt.expected.ActionType)
// 			}
// 			if result.FileID != tt.expected.FileID {
// 				t.Errorf("FileID = %v, want %v", result.FileID, tt.expected.FileID)
// 			}
// 			if result.EventID != tt.expected.EventID {
// 				t.Errorf("EventID = %v, want %v", result.EventID, tt.expected.EventID)
// 			}
// 			if result.TowerID != tt.expected.TowerID {
// 				t.Errorf("TowerID = %v, want %v", result.TowerID, tt.expected.TowerID)
// 			}
// 			if result.Size != tt.expected.Size {
// 				t.Errorf("Size = %v, want %v", result.Size, tt.expected.Size)
// 			}
// 			if result.ContentID != tt.expected.ContentID {
// 				t.Errorf("ContentID = %v, want %v", result.ContentID, tt.expected.ContentID)
// 			}
// 			if !result.Timestamp.Equal(tt.expected.Timestamp) {
// 				t.Errorf("Timestamp = %v, want %v", result.Timestamp, tt.expected.Timestamp)
// 			}
// 			if result.Filepath.ToPortable() != tt.expected.Filepath.ToPortable() {
// 				t.Errorf("Filepath = %v, want %v", result.Filepath.ToPortable(), tt.expected.Filepath.ToPortable())
// 			}
// 			if result.OriginPath.ToPortable() != tt.expected.OriginPath.ToPortable() {
// 				t.Errorf("OriginPath = %v, want %v", result.OriginPath.ToPortable(), tt.expected.OriginPath.ToPortable())
// 			}
// 			if result.DestinationPath.ToPortable() != tt.expected.DestinationPath.ToPortable() {
// 				t.Errorf("DestinationPath = %v, want %v", result.DestinationPath.ToPortable(), tt.expected.DestinationPath.ToPortable())
// 			}
// 		})
// 	}
// }
//
// func TestFileActionRoundTrip(t *testing.T) {
// 	now := time.Now().Truncate(time.Millisecond)
// 	filepath := fs.BuildFilePath("USERS", "testuser/file.txt")
// 	originPath := fs.BuildFilePath("USERS", "testuser/old/file.txt")
// 	destPath := fs.BuildFilePath("USERS", "testuser/new/file.txt")
//
// 	original := history.FileAction{
// 		ActionType:      history.FileMove,
// 		FileID:          "roundtrip123",
// 		Filepath:        filepath,
// 		OriginPath:      originPath,
// 		DestinationPath: destPath,
// 		EventID:         "eventRT",
// 		TowerID:         "towerRT",
// 		Timestamp:       now,
// 		Size:            4096,
// 		ContentID:       "contentRT",
// 	}
//
// 	info := reshape.FileActionToFileActionInfo(original)
// 	result := reshape.FileActionInfoToFileAction(info)
//
// 	if result.ActionType != original.ActionType {
// 		t.Errorf("ActionType mismatch after round trip")
// 	}
// 	if result.FileID != original.FileID {
// 		t.Errorf("FileID mismatch after round trip")
// 	}
// 	if !result.Timestamp.Equal(original.Timestamp) {
// 		t.Errorf("Timestamp mismatch after round trip: got %v, want %v", result.Timestamp, original.Timestamp)
// 	}
// 	if result.Filepath.ToPortable() != original.Filepath.ToPortable() {
// 		t.Errorf("Filepath mismatch after round trip")
// 	}
// 	if result.OriginPath.ToPortable() != original.OriginPath.ToPortable() {
// 		t.Errorf("OriginPath mismatch after round trip")
// 	}
// 	if result.DestinationPath.ToPortable() != original.DestinationPath.ToPortable() {
// 		t.Errorf("DestinationPath mismatch after round trip")
// 	}
// }
//
// // ============================================================================
// // media.go tests
// // ============================================================================
//
// func TestMediaToMediaInfo(t *testing.T) {
// 	now := time.Now()
// 	mediaID := primitive.NewObjectID()
//
// 	tests := []struct {
// 		name     string
// 		media    *media_model.Media
// 		expected structs.MediaInfo
// 	}{
// 		{
// 			name: "basic media",
// 			media: &media_model.Media{
// 				MediaID:    mediaID,
// 				ContentID:  "contentABC",
// 				FileIDs:    []string{"file1", "file2"},
// 				CreateDate: now,
// 				Owner:      "testuser",
// 				Width:      1920,
// 				Height:     1080,
// 				MimeType:   "image/jpeg",
// 				Hidden:     false,
// 				Enabled:    true,
// 				LikedBy:    []string{"user1", "user2"},
// 			},
// 			expected: structs.MediaInfo{
// 				MediaID:    mediaID.Hex(),
// 				ContentID:  "contentABC",
// 				FileIDs:    []string{"file1", "file2"},
// 				CreateDate: now.UnixMilli(),
// 				Owner:      "testuser",
// 				Width:      1920,
// 				Height:     1080,
// 				MimeType:   "image/jpeg",
// 				Hidden:     false,
// 				Enabled:    true,
// 				LikedBy:    []string{"user1", "user2"},
// 				Imported:   false,
// 			},
// 		},
// 		{
// 			name: "video media with duration",
// 			media: &media_model.Media{
// 				MediaID:    mediaID,
// 				ContentID:  "videoContent",
// 				FileIDs:    []string{"video1"},
// 				CreateDate: now,
// 				Owner:      "videouser",
// 				Width:      3840,
// 				Height:     2160,
// 				MimeType:   "video/mp4",
// 				Duration:   120000,
// 				Hidden:     false,
// 				Enabled:    true,
// 			},
// 			expected: structs.MediaInfo{
// 				MediaID:    mediaID.Hex(),
// 				ContentID:  "videoContent",
// 				FileIDs:    []string{"video1"},
// 				CreateDate: now.UnixMilli(),
// 				Owner:      "videouser",
// 				Width:      3840,
// 				Height:     2160,
// 				MimeType:   "video/mp4",
// 				Duration:   120000,
// 				Hidden:     false,
// 				Enabled:    true,
// 				Imported:   false,
// 			},
// 		},
// 		{
// 			name: "hidden media",
// 			media: &media_model.Media{
// 				MediaID:    mediaID,
// 				ContentID:  "hiddenContent",
// 				FileIDs:    []string{"hidden1"},
// 				CreateDate: now,
// 				Owner:      "hiddenuser",
// 				Width:      800,
// 				Height:     600,
// 				MimeType:   "image/png",
// 				Hidden:     true,
// 				Enabled:    true,
// 			},
// 			expected: structs.MediaInfo{
// 				MediaID:    mediaID.Hex(),
// 				ContentID:  "hiddenContent",
// 				FileIDs:    []string{"hidden1"},
// 				CreateDate: now.UnixMilli(),
// 				Owner:      "hiddenuser",
// 				Width:      800,
// 				Height:     600,
// 				MimeType:   "image/png",
// 				Hidden:     true,
// 				Enabled:    true,
// 				Imported:   false,
// 			},
// 		},
// 		{
// 			name: "disabled media",
// 			media: &media_model.Media{
// 				MediaID:    mediaID,
// 				ContentID:  "disabledContent",
// 				FileIDs:    []string{},
// 				CreateDate: now,
// 				Owner:      "disableduser",
// 				Width:      640,
// 				Height:     480,
// 				MimeType:   "image/gif",
// 				Hidden:     false,
// 				Enabled:    false,
// 			},
// 			expected: structs.MediaInfo{
// 				MediaID:    mediaID.Hex(),
// 				ContentID:  "disabledContent",
// 				FileIDs:    []string{},
// 				CreateDate: now.UnixMilli(),
// 				Owner:      "disableduser",
// 				Width:      640,
// 				Height:     480,
// 				MimeType:   "image/gif",
// 				Hidden:     false,
// 				Enabled:    false,
// 				Imported:   false,
// 			},
// 		},
// 		{
// 			name: "media with location",
// 			media: &media_model.Media{
// 				MediaID:    mediaID,
// 				ContentID:  "geoContent",
// 				FileIDs:    []string{"geo1"},
// 				CreateDate: now,
// 				Owner:      "geouser",
// 				Width:      1024,
// 				Height:     768,
// 				MimeType:   "image/jpeg",
// 				Location:   [2]float64{37.7749, -122.4194},
// 				Hidden:     false,
// 				Enabled:    true,
// 			},
// 			expected: structs.MediaInfo{
// 				MediaID:    mediaID.Hex(),
// 				ContentID:  "geoContent",
// 				FileIDs:    []string{"geo1"},
// 				CreateDate: now.UnixMilli(),
// 				Owner:      "geouser",
// 				Width:      1024,
// 				Height:     768,
// 				MimeType:   "image/jpeg",
// 				Location:   [2]float64{37.7749, -122.4194},
// 				Hidden:     false,
// 				Enabled:    true,
// 				Imported:   false,
// 			},
// 		},
// 		{
// 			name: "multi-page media",
// 			media: &media_model.Media{
// 				MediaID:    mediaID,
// 				ContentID:  "pdfContent",
// 				FileIDs:    []string{"pdf1"},
// 				CreateDate: now,
// 				Owner:      "pdfuser",
// 				Width:      612,
// 				Height:     792,
// 				MimeType:   "application/pdf",
// 				PageCount:  10,
// 				Hidden:     false,
// 				Enabled:    true,
// 			},
// 			expected: structs.MediaInfo{
// 				MediaID:    mediaID.Hex(),
// 				ContentID:  "pdfContent",
// 				FileIDs:    []string{"pdf1"},
// 				CreateDate: now.UnixMilli(),
// 				Owner:      "pdfuser",
// 				Width:      612,
// 				Height:     792,
// 				MimeType:   "application/pdf",
// 				PageCount:  10,
// 				Hidden:     false,
// 				Enabled:    true,
// 				Imported:   false,
// 			},
// 		},
// 		{
// 			name: "media with nil recognition tags",
// 			media: &media_model.Media{
// 				MediaID:         mediaID,
// 				ContentID:       "noTagsContent",
// 				FileIDs:         []string{"notag1"},
// 				CreateDate:      now,
// 				Owner:           "notaguser",
// 				Width:           100,
// 				Height:          100,
// 				MimeType:        "image/webp",
// 				RecognitionTags: nil,
// 				Hidden:          false,
// 				Enabled:         true,
// 			},
// 			expected: structs.MediaInfo{
// 				MediaID:         mediaID.Hex(),
// 				ContentID:       "noTagsContent",
// 				FileIDs:         []string{"notag1"},
// 				CreateDate:      now.UnixMilli(),
// 				Owner:           "notaguser",
// 				Width:           100,
// 				Height:          100,
// 				MimeType:        "image/webp",
// 				RecognitionTags: nil,
// 				Hidden:          false,
// 				Enabled:         true,
// 				Imported:        false,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.MediaToMediaInfo(tt.media)
//
// 			if result.MediaID != tt.expected.MediaID {
// 				t.Errorf("MediaID = %v, want %v", result.MediaID, tt.expected.MediaID)
// 			}
// 			if result.ContentID != tt.expected.ContentID {
// 				t.Errorf("ContentID = %v, want %v", result.ContentID, tt.expected.ContentID)
// 			}
// 			if result.Owner != tt.expected.Owner {
// 				t.Errorf("Owner = %v, want %v", result.Owner, tt.expected.Owner)
// 			}
// 			if result.Width != tt.expected.Width {
// 				t.Errorf("Width = %v, want %v", result.Width, tt.expected.Width)
// 			}
// 			if result.Height != tt.expected.Height {
// 				t.Errorf("Height = %v, want %v", result.Height, tt.expected.Height)
// 			}
// 			if result.MimeType != tt.expected.MimeType {
// 				t.Errorf("MimeType = %v, want %v", result.MimeType, tt.expected.MimeType)
// 			}
// 			if result.Hidden != tt.expected.Hidden {
// 				t.Errorf("Hidden = %v, want %v", result.Hidden, tt.expected.Hidden)
// 			}
// 			if result.Enabled != tt.expected.Enabled {
// 				t.Errorf("Enabled = %v, want %v", result.Enabled, tt.expected.Enabled)
// 			}
// 			if result.Duration != tt.expected.Duration {
// 				t.Errorf("Duration = %v, want %v", result.Duration, tt.expected.Duration)
// 			}
// 			if result.PageCount != tt.expected.PageCount {
// 				t.Errorf("PageCount = %v, want %v", result.PageCount, tt.expected.PageCount)
// 			}
// 			if result.CreateDate != tt.expected.CreateDate {
// 				t.Errorf("CreateDate = %v, want %v", result.CreateDate, tt.expected.CreateDate)
// 			}
// 			if result.Location != tt.expected.Location {
// 				t.Errorf("Location = %v, want %v", result.Location, tt.expected.Location)
// 			}
// 		})
// 	}
// }
//
// func TestNewMediaBatchInfo(t *testing.T) {
// 	now := time.Now()
// 	mediaID1 := primitive.NewObjectID()
// 	mediaID2 := primitive.NewObjectID()
// 	mediaID3 := primitive.NewObjectID()
//
// 	tests := []struct {
// 		name          string
// 		media         []*media_model.Media
// 		expectedCount int
// 		checkEmpty    bool
// 	}{
// 		{
// 			name:          "empty slice",
// 			media:         []*media_model.Media{},
// 			expectedCount: 0,
// 			checkEmpty:    true,
// 		},
// 		{
// 			name:          "nil slice",
// 			media:         nil,
// 			expectedCount: 0,
// 			checkEmpty:    true,
// 		},
// 		{
// 			name: "single media",
// 			media: []*media_model.Media{
// 				{
// 					MediaID:    mediaID1,
// 					ContentID:  "single",
// 					CreateDate: now,
// 					Owner:      "user1",
// 					Width:      100,
// 					Height:     100,
// 					MimeType:   "image/jpeg",
// 					Enabled:    true,
// 				},
// 			},
// 			expectedCount: 1,
// 			checkEmpty:    false,
// 		},
// 		{
// 			name: "multiple media",
// 			media: []*media_model.Media{
// 				{
// 					MediaID:    mediaID1,
// 					ContentID:  "first",
// 					CreateDate: now,
// 					Owner:      "user1",
// 					Width:      100,
// 					Height:     100,
// 					MimeType:   "image/jpeg",
// 					Enabled:    true,
// 				},
// 				{
// 					MediaID:    mediaID2,
// 					ContentID:  "second",
// 					CreateDate: now,
// 					Owner:      "user2",
// 					Width:      200,
// 					Height:     200,
// 					MimeType:   "image/png",
// 					Enabled:    true,
// 				},
// 				{
// 					MediaID:    mediaID3,
// 					ContentID:  "third",
// 					CreateDate: now,
// 					Owner:      "user3",
// 					Width:      300,
// 					Height:     300,
// 					MimeType:   "video/mp4",
// 					Duration:   60000,
// 					Enabled:    true,
// 				},
// 			},
// 			expectedCount: 3,
// 			checkEmpty:    false,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.NewMediaBatchInfo(tt.media)
//
// 			if result.MediaCount != tt.expectedCount {
// 				t.Errorf("MediaCount = %v, want %v", result.MediaCount, tt.expectedCount)
// 			}
// 			if len(result.Media) != tt.expectedCount {
// 				t.Errorf("len(Media) = %v, want %v", len(result.Media), tt.expectedCount)
// 			}
//
// 			if tt.checkEmpty {
// 				if result.Media == nil {
// 					t.Errorf("Media should not be nil, should be empty slice")
// 				}
// 			}
//
// 			// Verify each media item is correctly converted
// 			for i, m := range tt.media {
// 				if i < len(result.Media) {
// 					if result.Media[i].ContentID != m.ContentID {
// 						t.Errorf("Media[%d].ContentID = %v, want %v", i, result.Media[i].ContentID, m.ContentID)
// 					}
// 				}
// 			}
// 		})
// 	}
// }
//
// func TestMediaTypeToMediaTypeInfo(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		mtype    media_model.MType
// 		expected structs.MediaTypeInfo
// 	}{
// 		{
// 			name: "jpeg type",
// 			mtype: media_model.MType{
// 				Mime:        "image/jpeg",
// 				Name:        "Jpeg",
// 				Extensions:  []string{"jpeg", "jpg", "JPG"},
// 				Displayable: true,
// 				Raw:         false,
// 				IsVideo:     false,
// 				ImgRecog:    true,
// 				MultiPage:   false,
// 			},
// 			expected: structs.MediaTypeInfo{
// 				Mime:        "image/jpeg",
// 				Name:        "Jpeg",
// 				Extensions:  []string{"jpeg", "jpg", "JPG"},
// 				Displayable: true,
// 				Raw:         false,
// 				Video:       false,
// 				ImgRecog:    true,
// 				MultiPage:   false,
// 			},
// 		},
// 		{
// 			name: "raw image type",
// 			mtype: media_model.MType{
// 				Mime:            "image/x-sony-arw",
// 				Name:            "Sony ARW",
// 				RawThumbExifKey: "PreviewImage",
// 				Extensions:      []string{"ARW"},
// 				Displayable:     true,
// 				Raw:             true,
// 				IsVideo:         false,
// 				ImgRecog:        true,
// 				MultiPage:       false,
// 			},
// 			expected: structs.MediaTypeInfo{
// 				Mime:            "image/x-sony-arw",
// 				Name:            "Sony ARW",
// 				RawThumbExifKey: "PreviewImage",
// 				Extensions:      []string{"ARW"},
// 				Displayable:     true,
// 				Raw:             true,
// 				Video:           false,
// 				ImgRecog:        true,
// 				MultiPage:       false,
// 			},
// 		},
// 		{
// 			name: "video type",
// 			mtype: media_model.MType{
// 				Mime:        "video/mp4",
// 				Name:        "MP4",
// 				Extensions:  []string{"MP4", "mp4", "MOV"},
// 				Displayable: true,
// 				Raw:         false,
// 				IsVideo:     true,
// 				ImgRecog:    false,
// 				MultiPage:   false,
// 			},
// 			expected: structs.MediaTypeInfo{
// 				Mime:        "video/mp4",
// 				Name:        "MP4",
// 				Extensions:  []string{"MP4", "mp4", "MOV"},
// 				Displayable: true,
// 				Raw:         false,
// 				Video:       true,
// 				ImgRecog:    false,
// 				MultiPage:   false,
// 			},
// 		},
// 		{
// 			name: "PDF multi-page type",
// 			mtype: media_model.MType{
// 				Mime:        "application/pdf",
// 				Name:        "PDF",
// 				Extensions:  []string{"pdf"},
// 				Displayable: true,
// 				Raw:         false,
// 				IsVideo:     false,
// 				ImgRecog:    false,
// 				MultiPage:   true,
// 			},
// 			expected: structs.MediaTypeInfo{
// 				Mime:        "application/pdf",
// 				Name:        "PDF",
// 				Extensions:  []string{"pdf"},
// 				Displayable: true,
// 				Raw:         false,
// 				Video:       false,
// 				ImgRecog:    false,
// 				MultiPage:   true,
// 			},
// 		},
// 		{
// 			name: "generic type",
// 			mtype: media_model.MType{
// 				Mime:        "generic",
// 				Name:        "File",
// 				Extensions:  []string{},
// 				Displayable: false,
// 				Raw:         false,
// 				IsVideo:     false,
// 				ImgRecog:    false,
// 				MultiPage:   false,
// 			},
// 			expected: structs.MediaTypeInfo{
// 				Mime:        "generic",
// 				Name:        "File",
// 				Extensions:  []string{},
// 				Displayable: false,
// 				Raw:         false,
// 				Video:       false,
// 				ImgRecog:    false,
// 				MultiPage:   false,
// 			},
// 		},
// 		{
// 			name: "nil extensions",
// 			mtype: media_model.MType{
// 				Mime:        "application/octet-stream",
// 				Name:        "Binary",
// 				Extensions:  nil,
// 				Displayable: false,
// 			},
// 			expected: structs.MediaTypeInfo{
// 				Mime:        "application/octet-stream",
// 				Name:        "Binary",
// 				Extensions:  nil,
// 				Displayable: false,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.MediaTypeToMediaTypeInfo(tt.mtype)
//
// 			if result.Mime != tt.expected.Mime {
// 				t.Errorf("Mime = %v, want %v", result.Mime, tt.expected.Mime)
// 			}
// 			if result.Name != tt.expected.Name {
// 				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
// 			}
// 			if result.RawThumbExifKey != tt.expected.RawThumbExifKey {
// 				t.Errorf("RawThumbExifKey = %v, want %v", result.RawThumbExifKey, tt.expected.RawThumbExifKey)
// 			}
// 			if result.Displayable != tt.expected.Displayable {
// 				t.Errorf("Displayable = %v, want %v", result.Displayable, tt.expected.Displayable)
// 			}
// 			if result.Raw != tt.expected.Raw {
// 				t.Errorf("Raw = %v, want %v", result.Raw, tt.expected.Raw)
// 			}
// 			if result.Video != tt.expected.Video {
// 				t.Errorf("Video = %v, want %v", result.Video, tt.expected.Video)
// 			}
// 			if result.ImgRecog != tt.expected.ImgRecog {
// 				t.Errorf("ImgRecog = %v, want %v", result.ImgRecog, tt.expected.ImgRecog)
// 			}
// 			if result.MultiPage != tt.expected.MultiPage {
// 				t.Errorf("MultiPage = %v, want %v", result.MultiPage, tt.expected.MultiPage)
// 			}
// 		})
// 	}
// }
//
// // ============================================================================
// // share.go tests
// // ============================================================================
//
// func TestPermissionsToPermissionsInfo(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		perms    map[string]*share_model.Permissions
// 		expected map[string]structs.PermissionsInfo
// 	}{
// 		{
// 			name:     "nil permissions",
// 			perms:    nil,
// 			expected: map[string]structs.PermissionsInfo{},
// 		},
// 		{
// 			name:     "empty permissions",
// 			perms:    map[string]*share_model.Permissions{},
// 			expected: map[string]structs.PermissionsInfo{},
// 		},
// 		{
// 			name: "single user permissions",
// 			perms: map[string]*share_model.Permissions{
// 				"user1": {CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 			},
// 			expected: map[string]structs.PermissionsInfo{
// 				"user1": {CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 			},
// 		},
// 		{
// 			name: "multiple user permissions",
// 			perms: map[string]*share_model.Permissions{
// 				"user1": {CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 				"user2": {CanView: true, CanEdit: true, CanDownload: true, CanDelete: true},
// 				"user3": {CanView: true, CanEdit: false, CanDownload: false, CanDelete: false},
// 			},
// 			expected: map[string]structs.PermissionsInfo{
// 				"user1": {CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 				"user2": {CanView: true, CanEdit: true, CanDownload: true, CanDelete: true},
// 				"user3": {CanView: true, CanEdit: false, CanDownload: false, CanDelete: false},
// 			},
// 		},
// 		{
// 			name: "all permissions false",
// 			perms: map[string]*share_model.Permissions{
// 				"restricted": {CanView: false, CanEdit: false, CanDownload: false, CanDelete: false},
// 			},
// 			expected: map[string]structs.PermissionsInfo{
// 				"restricted": {CanView: false, CanEdit: false, CanDownload: false, CanDelete: false},
// 			},
// 		},
// 		{
// 			name: "all permissions true",
// 			perms: map[string]*share_model.Permissions{
// 				"admin": {CanView: true, CanEdit: true, CanDownload: true, CanDelete: true},
// 			},
// 			expected: map[string]structs.PermissionsInfo{
// 				"admin": {CanView: true, CanEdit: true, CanDownload: true, CanDelete: true},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.PermissionsToPermissionsInfo(ctxservice.RequestContext{}, tt.perms)
//
// 			if len(result) != len(tt.expected) {
// 				t.Errorf("len(result) = %v, want %v", len(result), len(tt.expected))
// 				return
// 			}
//
// 			for k, expectedInfo := range tt.expected {
// 				resultInfo, ok := result[k]
// 				if !ok {
// 					t.Errorf("missing key %s in result", k)
// 					continue
// 				}
// 				if resultInfo.CanView != expectedInfo.CanView {
// 					t.Errorf("%s.CanView = %v, want %v", k, resultInfo.CanView, expectedInfo.CanView)
// 				}
// 				if resultInfo.CanEdit != expectedInfo.CanEdit {
// 					t.Errorf("%s.CanEdit = %v, want %v", k, resultInfo.CanEdit, expectedInfo.CanEdit)
// 				}
// 				if resultInfo.CanDownload != expectedInfo.CanDownload {
// 					t.Errorf("%s.CanDownload = %v, want %v", k, resultInfo.CanDownload, expectedInfo.CanDownload)
// 				}
// 				if resultInfo.CanDelete != expectedInfo.CanDelete {
// 					t.Errorf("%s.CanDelete = %v, want %v", k, resultInfo.CanDelete, expectedInfo.CanDelete)
// 				}
// 			}
// 		})
// 	}
// }
//
// func TestPermissionsParamsToPermissions(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		params   structs.PermissionsParams
// 		expected share_model.Permissions
// 	}{
// 		{
// 			name:     "all false",
// 			params:   structs.PermissionsParams{CanView: false, CanEdit: false, CanDownload: false, CanDelete: false},
// 			expected: share_model.Permissions{CanView: false, CanEdit: false, CanDownload: false, CanDelete: false},
// 		},
// 		{
// 			name:     "all true",
// 			params:   structs.PermissionsParams{CanView: true, CanEdit: true, CanDownload: true, CanDelete: true},
// 			expected: share_model.Permissions{CanView: true, CanEdit: true, CanDownload: true, CanDelete: true},
// 		},
// 		{
// 			name:     "view only",
// 			params:   structs.PermissionsParams{CanView: true, CanEdit: false, CanDownload: false, CanDelete: false},
// 			expected: share_model.Permissions{CanView: true, CanEdit: false, CanDownload: false, CanDelete: false},
// 		},
// 		{
// 			name:     "view and download",
// 			params:   structs.PermissionsParams{CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 			expected: share_model.Permissions{CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 		},
// 		{
// 			name:     "view edit and download",
// 			params:   structs.PermissionsParams{CanView: true, CanEdit: true, CanDownload: true, CanDelete: false},
// 			expected: share_model.Permissions{CanView: true, CanEdit: true, CanDownload: true, CanDelete: false},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result, err := reshape.PermissionsParamsToPermissions(ctxservice.RequestContext{}, tt.params)
// 			if err != nil {
// 				t.Errorf("unexpected error: %v", err)
// 				return
// 			}
//
// 			if result.CanView != tt.expected.CanView {
// 				t.Errorf("CanView = %v, want %v", result.CanView, tt.expected.CanView)
// 			}
// 			if result.CanEdit != tt.expected.CanEdit {
// 				t.Errorf("CanEdit = %v, want %v", result.CanEdit, tt.expected.CanEdit)
// 			}
// 			if result.CanDownload != tt.expected.CanDownload {
// 				t.Errorf("CanDownload = %v, want %v", result.CanDownload, tt.expected.CanDownload)
// 			}
// 			if result.CanDelete != tt.expected.CanDelete {
// 				t.Errorf("CanDelete = %v, want %v", result.CanDelete, tt.expected.CanDelete)
// 			}
// 		})
// 	}
// }
//
// func TestUnpackNewUserParams(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		params        structs.AddUserParams
// 		expectedUser  string
// 		expectedPerms share_model.Permissions
// 	}{
// 		{
// 			name: "basic user with default perms",
// 			params: structs.AddUserParams{
// 				Username: "testuser",
// 				PermissionsParams: structs.PermissionsParams{
// 					CanView:     true,
// 					CanEdit:     false,
// 					CanDownload: true,
// 					CanDelete:   false,
// 				},
// 			},
// 			expectedUser: "testuser",
// 			expectedPerms: share_model.Permissions{
// 				CanView:     true,
// 				CanEdit:     false,
// 				CanDownload: true,
// 				CanDelete:   false,
// 			},
// 		},
// 		{
// 			name: "user with full perms",
// 			params: structs.AddUserParams{
// 				Username: "adminuser",
// 				PermissionsParams: structs.PermissionsParams{
// 					CanView:     true,
// 					CanEdit:     true,
// 					CanDownload: true,
// 					CanDelete:   true,
// 				},
// 			},
// 			expectedUser: "adminuser",
// 			expectedPerms: share_model.Permissions{
// 				CanView:     true,
// 				CanEdit:     true,
// 				CanDownload: true,
// 				CanDelete:   true,
// 			},
// 		},
// 		{
// 			name: "empty username",
// 			params: structs.AddUserParams{
// 				Username: "",
// 				PermissionsParams: structs.PermissionsParams{
// 					CanView: true,
// 				},
// 			},
// 			expectedUser: "",
// 			expectedPerms: share_model.Permissions{
// 				CanView: true,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			username, perms, err := reshape.UnpackNewUserParams(ctxservice.RequestContext{}, tt.params)
// 			if err != nil {
// 				t.Errorf("unexpected error: %v", err)
// 				return
// 			}
//
// 			if username != tt.expectedUser {
// 				t.Errorf("username = %v, want %v", username, tt.expectedUser)
// 			}
// 			if perms.CanView != tt.expectedPerms.CanView {
// 				t.Errorf("perms.CanView = %v, want %v", perms.CanView, tt.expectedPerms.CanView)
// 			}
// 			if perms.CanEdit != tt.expectedPerms.CanEdit {
// 				t.Errorf("perms.CanEdit = %v, want %v", perms.CanEdit, tt.expectedPerms.CanEdit)
// 			}
// 			if perms.CanDownload != tt.expectedPerms.CanDownload {
// 				t.Errorf("perms.CanDownload = %v, want %v", perms.CanDownload, tt.expectedPerms.CanDownload)
// 			}
// 			if perms.CanDelete != tt.expectedPerms.CanDelete {
// 				t.Errorf("perms.CanDelete = %v, want %v", perms.CanDelete, tt.expectedPerms.CanDelete)
// 			}
// 		})
// 	}
// }
//
// // ============================================================================
// // token.go tests
// // ============================================================================
//
// func TestTokenToTokenInfo(t *testing.T) {
// 	now := time.Now()
// 	tokenID := primitive.NewObjectID()
//
// 	// Create a test token with known bytes
// 	var tokenBytes [32]byte
// 	for i := 0; i < 32; i++ {
// 		tokenBytes[i] = byte(i)
// 	}
//
// 	tests := []struct {
// 		name     string
// 		token    *auth_model.Token
// 		expected structs.TokenInfo
// 	}{
// 		{
// 			name: "basic token",
// 			token: &auth_model.Token{
// 				ID:          tokenID,
// 				CreatedTime: now,
// 				LastUsed:    now,
// 				Nickname:    "test-token",
// 				Owner:       "testuser",
// 				RemoteUsing: "",
// 				CreatedBy:   "tower123",
// 				Token:       tokenBytes,
// 			},
// 			expected: structs.TokenInfo{
// 				ID:          tokenID.Hex(),
// 				CreatedTime: now.UnixMilli(),
// 				LastUsed:    now.UnixMilli(),
// 				Nickname:    "test-token",
// 				Owner:       "testuser",
// 				RemoteUsing: "",
// 				CreatedBy:   "tower123",
// 				Token:       base64.StdEncoding.EncodeToString(tokenBytes[:]),
// 			},
// 		},
// 		{
// 			name: "token with remote using",
// 			token: &auth_model.Token{
// 				ID:          tokenID,
// 				CreatedTime: now,
// 				LastUsed:    now,
// 				Nickname:    "remote-token",
// 				Owner:       "remoteuser",
// 				RemoteUsing: "remote-tower",
// 				CreatedBy:   "core-tower",
// 				Token:       tokenBytes,
// 			},
// 			expected: structs.TokenInfo{
// 				ID:          tokenID.Hex(),
// 				CreatedTime: now.UnixMilli(),
// 				LastUsed:    now.UnixMilli(),
// 				Nickname:    "remote-token",
// 				Owner:       "remoteuser",
// 				RemoteUsing: "remote-tower",
// 				CreatedBy:   "core-tower",
// 				Token:       base64.StdEncoding.EncodeToString(tokenBytes[:]),
// 			},
// 		},
// 		{
// 			name: "token with zero time",
// 			token: &auth_model.Token{
// 				ID:          tokenID,
// 				CreatedTime: time.Time{},
// 				LastUsed:    time.Time{},
// 				Nickname:    "zero-time-token",
// 				Owner:       "user",
// 				Token:       tokenBytes,
// 			},
// 			expected: structs.TokenInfo{
// 				ID:          tokenID.Hex(),
// 				CreatedTime: time.Time{}.UnixMilli(),
// 				LastUsed:    time.Time{}.UnixMilli(),
// 				Nickname:    "zero-time-token",
// 				Owner:       "user",
// 				Token:       base64.StdEncoding.EncodeToString(tokenBytes[:]),
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.TokenToTokenInfo(context.Background(), tt.token)
//
// 			if result.ID != tt.expected.ID {
// 				t.Errorf("ID = %v, want %v", result.ID, tt.expected.ID)
// 			}
// 			if result.CreatedTime != tt.expected.CreatedTime {
// 				t.Errorf("CreatedTime = %v, want %v", result.CreatedTime, tt.expected.CreatedTime)
// 			}
// 			if result.LastUsed != tt.expected.LastUsed {
// 				t.Errorf("LastUsed = %v, want %v", result.LastUsed, tt.expected.LastUsed)
// 			}
// 			if result.Nickname != tt.expected.Nickname {
// 				t.Errorf("Nickname = %v, want %v", result.Nickname, tt.expected.Nickname)
// 			}
// 			if result.Owner != tt.expected.Owner {
// 				t.Errorf("Owner = %v, want %v", result.Owner, tt.expected.Owner)
// 			}
// 			if result.RemoteUsing != tt.expected.RemoteUsing {
// 				t.Errorf("RemoteUsing = %v, want %v", result.RemoteUsing, tt.expected.RemoteUsing)
// 			}
// 			if result.CreatedBy != tt.expected.CreatedBy {
// 				t.Errorf("CreatedBy = %v, want %v", result.CreatedBy, tt.expected.CreatedBy)
// 			}
// 			if result.Token != tt.expected.Token {
// 				t.Errorf("Token = %v, want %v", result.Token, tt.expected.Token)
// 			}
// 		})
// 	}
// }
//
// func TestTokenInfoToToken(t *testing.T) {
// 	now := time.Now().Truncate(time.Millisecond)
// 	tokenID := primitive.NewObjectID()
//
// 	var tokenBytes [32]byte
// 	for i := 0; i < 32; i++ {
// 		tokenBytes[i] = byte(i)
// 	}
// 	tokenStr := base64.StdEncoding.EncodeToString(tokenBytes[:])
//
// 	tests := []struct {
// 		name        string
// 		info        structs.TokenInfo
// 		expected    *auth_model.Token
// 		expectError bool
// 	}{
// 		{
// 			name: "basic token info",
// 			info: structs.TokenInfo{
// 				ID:          tokenID.Hex(),
// 				CreatedTime: now.UnixMilli(),
// 				LastUsed:    now.UnixMilli(),
// 				Nickname:    "test-token",
// 				Owner:       "testuser",
// 				RemoteUsing: "",
// 				CreatedBy:   "tower123",
// 				Token:       tokenStr,
// 			},
// 			expected: &auth_model.Token{
// 				ID:          tokenID,
// 				CreatedTime: now,
// 				LastUsed:    now,
// 				Nickname:    "test-token",
// 				Owner:       "testuser",
// 				RemoteUsing: "",
// 				CreatedBy:   "tower123",
// 				Token:       tokenBytes,
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "invalid token encoding",
// 			info: structs.TokenInfo{
// 				ID:    tokenID.Hex(),
// 				Token: "not-valid-base64!!!",
// 			},
// 			expected:    nil,
// 			expectError: true,
// 		},
// 		{
// 			name: "empty token ID",
// 			info: structs.TokenInfo{
// 				ID:          "",
// 				CreatedTime: now.UnixMilli(),
// 				LastUsed:    now.UnixMilli(),
// 				Nickname:    "empty-id-token",
// 				Owner:       "user",
// 				Token:       tokenStr,
// 			},
// 			expected: &auth_model.Token{
// 				ID:          primitive.NilObjectID,
// 				CreatedTime: now,
// 				LastUsed:    now,
// 				Nickname:    "empty-id-token",
// 				Owner:       "user",
// 				Token:       tokenBytes,
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "invalid token ID",
// 			info: structs.TokenInfo{
// 				ID:          "not-a-valid-object-id",
// 				CreatedTime: now.UnixMilli(),
// 				LastUsed:    now.UnixMilli(),
// 				Nickname:    "invalid-id-token",
// 				Owner:       "user",
// 				Token:       tokenStr,
// 			},
// 			expected: &auth_model.Token{
// 				ID:          primitive.NilObjectID,
// 				CreatedTime: now,
// 				LastUsed:    now,
// 				Nickname:    "invalid-id-token",
// 				Owner:       "user",
// 				Token:       tokenBytes,
// 			},
// 			expectError: false,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result, err := reshape.TokenInfoToToken(context.Background(), tt.info)
//
// 			if tt.expectError {
// 				if err == nil {
// 					t.Errorf("expected error but got none")
// 				}
// 				return
// 			}
//
// 			if err != nil {
// 				t.Errorf("unexpected error: %v", err)
// 				return
// 			}
//
// 			if result == nil {
// 				t.Errorf("result is nil")
// 				return
// 			}
//
// 			if !result.CreatedTime.Equal(tt.expected.CreatedTime) {
// 				t.Errorf("CreatedTime = %v, want %v", result.CreatedTime, tt.expected.CreatedTime)
// 			}
// 			if !result.LastUsed.Equal(tt.expected.LastUsed) {
// 				t.Errorf("LastUsed = %v, want %v", result.LastUsed, tt.expected.LastUsed)
// 			}
// 			if result.Nickname != tt.expected.Nickname {
// 				t.Errorf("Nickname = %v, want %v", result.Nickname, tt.expected.Nickname)
// 			}
// 			if result.Owner != tt.expected.Owner {
// 				t.Errorf("Owner = %v, want %v", result.Owner, tt.expected.Owner)
// 			}
// 			if result.RemoteUsing != tt.expected.RemoteUsing {
// 				t.Errorf("RemoteUsing = %v, want %v", result.RemoteUsing, tt.expected.RemoteUsing)
// 			}
// 			if result.CreatedBy != tt.expected.CreatedBy {
// 				t.Errorf("CreatedBy = %v, want %v", result.CreatedBy, tt.expected.CreatedBy)
// 			}
// 			if result.Token != tt.expected.Token {
// 				t.Errorf("Token = %v, want %v", result.Token, tt.expected.Token)
// 			}
// 		})
// 	}
// }
//
// func TestTokensToTokenInfos(t *testing.T) {
// 	now := time.Now()
// 	tokenID1 := primitive.NewObjectID()
// 	tokenID2 := primitive.NewObjectID()
//
// 	var tokenBytes1 [32]byte
// 	var tokenBytes2 [32]byte
// 	for i := 0; i < 32; i++ {
// 		tokenBytes1[i] = byte(i)
// 		tokenBytes2[i] = byte(32 - i)
// 	}
//
// 	tests := []struct {
// 		name     string
// 		tokens   []*auth_model.Token
// 		expected int
// 	}{
// 		{
// 			name:     "empty tokens",
// 			tokens:   []*auth_model.Token{},
// 			expected: 0,
// 		},
// 		{
// 			name: "single token",
// 			tokens: []*auth_model.Token{
// 				{
// 					ID:          tokenID1,
// 					CreatedTime: now,
// 					LastUsed:    now,
// 					Nickname:    "token1",
// 					Owner:       "user1",
// 					Token:       tokenBytes1,
// 				},
// 			},
// 			expected: 1,
// 		},
// 		{
// 			name: "multiple tokens",
// 			tokens: []*auth_model.Token{
// 				{
// 					ID:          tokenID1,
// 					CreatedTime: now,
// 					LastUsed:    now,
// 					Nickname:    "token1",
// 					Owner:       "user1",
// 					Token:       tokenBytes1,
// 				},
// 				{
// 					ID:          tokenID2,
// 					CreatedTime: now,
// 					LastUsed:    now,
// 					Nickname:    "token2",
// 					Owner:       "user2",
// 					Token:       tokenBytes2,
// 				},
// 			},
// 			expected: 2,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.TokensToTokenInfos(context.Background(), tt.tokens)
//
// 			if len(result) != tt.expected {
// 				t.Errorf("len(result) = %v, want %v", len(result), tt.expected)
// 			}
//
// 			for i, token := range tt.tokens {
// 				if i < len(result) {
// 					if result[i].ID != token.ID.Hex() {
// 						t.Errorf("result[%d].ID = %v, want %v", i, result[i].ID, token.ID.Hex())
// 					}
// 					if result[i].Nickname != token.Nickname {
// 						t.Errorf("result[%d].Nickname = %v, want %v", i, result[i].Nickname, token.Nickname)
// 					}
// 				}
// 			}
// 		})
// 	}
// }
//
// func TestTokenRoundTrip(t *testing.T) {
// 	now := time.Now().Truncate(time.Millisecond)
// 	tokenID := primitive.NewObjectID()
//
// 	var tokenBytes [32]byte
// 	for i := 0; i < 32; i++ {
// 		tokenBytes[i] = byte(i * 2)
// 	}
//
// 	original := &auth_model.Token{
// 		ID:          tokenID,
// 		CreatedTime: now,
// 		LastUsed:    now,
// 		Nickname:    "roundtrip-token",
// 		Owner:       "roundtripuser",
// 		RemoteUsing: "remote-123",
// 		CreatedBy:   "tower-456",
// 		Token:       tokenBytes,
// 	}
//
// 	info := reshape.TokenToTokenInfo(context.Background(), original)
// 	result, err := reshape.TokenInfoToToken(context.Background(), info)
//
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
//
// 	if result.ID != original.ID {
// 		t.Errorf("ID mismatch after round trip")
// 	}
// 	if !result.CreatedTime.Equal(original.CreatedTime) {
// 		t.Errorf("CreatedTime mismatch after round trip")
// 	}
// 	if !result.LastUsed.Equal(original.LastUsed) {
// 		t.Errorf("LastUsed mismatch after round trip")
// 	}
// 	if result.Nickname != original.Nickname {
// 		t.Errorf("Nickname mismatch after round trip")
// 	}
// 	if result.Owner != original.Owner {
// 		t.Errorf("Owner mismatch after round trip")
// 	}
// 	if result.RemoteUsing != original.RemoteUsing {
// 		t.Errorf("RemoteUsing mismatch after round trip")
// 	}
// 	if result.CreatedBy != original.CreatedBy {
// 		t.Errorf("CreatedBy mismatch after round trip")
// 	}
// 	if result.Token != original.Token {
// 		t.Errorf("Token mismatch after round trip")
// 	}
// }
//
// // ============================================================================
// // tower.go tests
// // ============================================================================
//
// func TestTowerInfoToTower(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		info     structs.TowerInfo
// 		expected *tower_model.Instance
// 	}{
// 		{
// 			name: "core tower",
// 			info: structs.TowerInfo{
// 				ID:           "tower123",
// 				Name:         "Core Tower",
// 				Role:         "core",
// 				Address:      "https://core.example.com",
// 				IsThisServer: true,
// 				LastBackup:   1234567890000,
// 			},
// 			expected: &tower_model.Instance{
// 				TowerID:     "tower123",
// 				Name:        "Core Tower",
// 				Role:        tower_model.RoleCore,
// 				IsThisTower: true,
// 				Address:     "https://core.example.com",
// 				LastBackup:  1234567890000,
// 			},
// 		},
// 		{
// 			name: "backup tower",
// 			info: structs.TowerInfo{
// 				ID:           "tower456",
// 				Name:         "Backup Tower",
// 				Role:         "backup",
// 				Address:      "https://backup.example.com",
// 				IsThisServer: false,
// 				LastBackup:   1234567890000,
// 			},
// 			expected: &tower_model.Instance{
// 				TowerID:     "tower456",
// 				Name:        "Backup Tower",
// 				Role:        tower_model.RoleBackup,
// 				IsThisTower: false,
// 				Address:     "https://backup.example.com",
// 				LastBackup:  1234567890000,
// 			},
// 		},
// 		{
// 			name: "init tower",
// 			info: structs.TowerInfo{
// 				ID:           "tower789",
// 				Name:         "Init Tower",
// 				Role:         "init",
// 				Address:      "",
// 				IsThisServer: true,
// 				LastBackup:   0,
// 			},
// 			expected: &tower_model.Instance{
// 				TowerID:     "tower789",
// 				Name:        "Init Tower",
// 				Role:        tower_model.RoleInit,
// 				IsThisTower: true,
// 				Address:     "",
// 				LastBackup:  0,
// 			},
// 		},
// 		{
// 			name: "empty values",
// 			info: structs.TowerInfo{
// 				ID:   "",
// 				Name: "",
// 				Role: "",
// 			},
// 			expected: &tower_model.Instance{
// 				TowerID: "",
// 				Name:    "",
// 				Role:    tower_model.Role(""),
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.TowerInfoToTower(tt.info)
//
// 			if result.TowerID != tt.expected.TowerID {
// 				t.Errorf("TowerID = %v, want %v", result.TowerID, tt.expected.TowerID)
// 			}
// 			if result.Name != tt.expected.Name {
// 				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
// 			}
// 			if result.Role != tt.expected.Role {
// 				t.Errorf("Role = %v, want %v", result.Role, tt.expected.Role)
// 			}
// 			if result.IsThisTower != tt.expected.IsThisTower {
// 				t.Errorf("IsThisTower = %v, want %v", result.IsThisTower, tt.expected.IsThisTower)
// 			}
// 			if result.Address != tt.expected.Address {
// 				t.Errorf("Address = %v, want %v", result.Address, tt.expected.Address)
// 			}
// 			if result.LastBackup != tt.expected.LastBackup {
// 				t.Errorf("LastBackup = %v, want %v", result.LastBackup, tt.expected.LastBackup)
// 			}
// 		})
// 	}
// }
//
// func TestAPITowerInfoToTower(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		info     openapi.TowerInfo
// 		expected tower_model.Instance
// 	}{
// 		{
// 			name: "core tower from API",
// 			info: openapi.TowerInfo{
// 				Id:          "api-tower-123",
// 				Name:        "API Core Tower",
// 				Role:        "core",
// 				CoreAddress: "https://api.example.com",
// 				LastBackup:  1234567890000,
// 			},
// 			expected: tower_model.Instance{
// 				TowerID:     "api-tower-123",
// 				Name:        "API Core Tower",
// 				Role:        tower_model.RoleCore,
// 				IsThisTower: false,
// 				Address:     "https://api.example.com",
// 				LastBackup:  1234567890000,
// 			},
// 		},
// 		{
// 			name: "backup tower from API",
// 			info: openapi.TowerInfo{
// 				Id:          "api-tower-456",
// 				Name:        "API Backup Tower",
// 				Role:        "backup",
// 				CoreAddress: "https://backup.example.com",
// 				LastBackup:  9876543210000,
// 			},
// 			expected: tower_model.Instance{
// 				TowerID:     "api-tower-456",
// 				Name:        "API Backup Tower",
// 				Role:        tower_model.RoleBackup,
// 				IsThisTower: false,
// 				Address:     "https://backup.example.com",
// 				LastBackup:  9876543210000,
// 			},
// 		},
// 		{
// 			name: "empty API tower",
// 			info: openapi.TowerInfo{},
// 			expected: tower_model.Instance{
// 				IsThisTower: false,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.APITowerInfoToTower(tt.info)
//
// 			if result.TowerID != tt.expected.TowerID {
// 				t.Errorf("TowerID = %v, want %v", result.TowerID, tt.expected.TowerID)
// 			}
// 			if result.Name != tt.expected.Name {
// 				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
// 			}
// 			if result.Role != tt.expected.Role {
// 				t.Errorf("Role = %v, want %v", result.Role, tt.expected.Role)
// 			}
// 			if result.IsThisTower != tt.expected.IsThisTower {
// 				t.Errorf("IsThisTower = %v, want %v", result.IsThisTower, tt.expected.IsThisTower)
// 			}
// 			if result.Address != tt.expected.Address {
// 				t.Errorf("Address = %v, want %v", result.Address, tt.expected.Address)
// 			}
// 			if result.LastBackup != tt.expected.LastBackup {
// 				t.Errorf("LastBackup = %v, want %v", result.LastBackup, tt.expected.LastBackup)
// 			}
// 		})
// 	}
// }
//
// // ============================================================================
// // user.go tests
// // ============================================================================
//
// func TestUserInfoArchiveToUser(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		info     structs.UserInfoArchive
// 		expected *user_model.User
// 	}{
// 		{
// 			name: "basic user",
// 			info: structs.UserInfoArchive{
// 				UserInfo: structs.UserInfo{
// 					Username:        "testuser",
// 					FullName:        "Test User",
// 					HomeID:          "home123",
// 					TrashID:         "trash123",
// 					PermissionLevel: int(user_model.UserPermissionBasic),
// 					Activated:       true,
// 				},
// 				Password: "hashed_password",
// 			},
// 			expected: &user_model.User{
// 				Username:  "testuser",
// 				Password:  "hashed_password",
// 				Activated: true,
// 				UserPerms: user_model.UserPermissionBasic,
// 				HomeID:    "home123",
// 				TrashID:   "trash123",
// 			},
// 		},
// 		{
// 			name: "admin user",
// 			info: structs.UserInfoArchive{
// 				UserInfo: structs.UserInfo{
// 					Username:        "adminuser",
// 					FullName:        "Admin User",
// 					HomeID:          "home-admin",
// 					TrashID:         "trash-admin",
// 					PermissionLevel: int(user_model.UserPermissionAdmin),
// 					Activated:       true,
// 				},
// 				Password: "admin_hashed_password",
// 			},
// 			expected: &user_model.User{
// 				Username:  "adminuser",
// 				Password:  "admin_hashed_password",
// 				Activated: true,
// 				UserPerms: user_model.UserPermissionAdmin,
// 				HomeID:    "home-admin",
// 				TrashID:   "trash-admin",
// 			},
// 		},
// 		{
// 			name: "owner user",
// 			info: structs.UserInfoArchive{
// 				UserInfo: structs.UserInfo{
// 					Username:        "owneruser",
// 					FullName:        "Owner User",
// 					HomeID:          "home-owner",
// 					TrashID:         "trash-owner",
// 					PermissionLevel: int(user_model.UserPermissionOwner),
// 					Activated:       true,
// 				},
// 				Password: "owner_hashed_password",
// 			},
// 			expected: &user_model.User{
// 				Username:  "owneruser",
// 				Password:  "owner_hashed_password",
// 				Activated: true,
// 				UserPerms: user_model.UserPermissionOwner,
// 				HomeID:    "home-owner",
// 				TrashID:   "trash-owner",
// 			},
// 		},
// 		{
// 			name: "inactive user",
// 			info: structs.UserInfoArchive{
// 				UserInfo: structs.UserInfo{
// 					Username:        "inactiveuser",
// 					FullName:        "Inactive User",
// 					HomeID:          "home-inactive",
// 					TrashID:         "trash-inactive",
// 					PermissionLevel: int(user_model.UserPermissionBasic),
// 					Activated:       false,
// 				},
// 				Password: "inactive_password",
// 			},
// 			expected: &user_model.User{
// 				Username:  "inactiveuser",
// 				Password:  "inactive_password",
// 				Activated: false,
// 				UserPerms: user_model.UserPermissionBasic,
// 				HomeID:    "home-inactive",
// 				TrashID:   "trash-inactive",
// 			},
// 		},
// 		{
// 			name: "empty user info",
// 			info: structs.UserInfoArchive{
// 				UserInfo: structs.UserInfo{},
// 				Password: "",
// 			},
// 			expected: &user_model.User{
// 				Username:  "",
// 				Password:  "",
// 				Activated: false,
// 				UserPerms: user_model.UserPermissionPublic,
// 				HomeID:    "",
// 				TrashID:   "",
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.UserInfoArchiveToUser(tt.info)
//
// 			if result.Username != tt.expected.Username {
// 				t.Errorf("Username = %v, want %v", result.Username, tt.expected.Username)
// 			}
// 			if result.Password != tt.expected.Password {
// 				t.Errorf("Password = %v, want %v", result.Password, tt.expected.Password)
// 			}
// 			if result.Activated != tt.expected.Activated {
// 				t.Errorf("Activated = %v, want %v", result.Activated, tt.expected.Activated)
// 			}
// 			if result.UserPerms != tt.expected.UserPerms {
// 				t.Errorf("UserPerms = %v, want %v", result.UserPerms, tt.expected.UserPerms)
// 			}
// 			if result.HomeID != tt.expected.HomeID {
// 				t.Errorf("HomeID = %v, want %v", result.HomeID, tt.expected.HomeID)
// 			}
// 			if result.TrashID != tt.expected.TrashID {
// 				t.Errorf("TrashID = %v, want %v", result.TrashID, tt.expected.TrashID)
// 			}
// 		})
// 	}
// }
//
// // ============================================================================
// // websocket.go tests
// // ============================================================================
//
// func TestGetSubscribeInfo(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		msg      websocket_mod.WsResponseInfo
// 		expected websocket_mod.SubscriptionInfo
// 	}{
// 		{
// 			name: "basic subscription",
// 			msg: websocket_mod.WsResponseInfo{
// 				SentTime:      1234567890000,
// 				BroadcastType: websocket_mod.SubscriptionType("folder"),
// 				SubscribeKey:  "folder-123",
// 				Content:       websocket_mod.WsData{"shareID": "share-456"},
// 			},
// 			expected: websocket_mod.SubscriptionInfo{
// 				When:           time.UnixMilli(1234567890000),
// 				Type:           websocket_mod.SubscriptionType("folder"),
// 				SubscriptionID: "folder-123",
// 				ShareID:        "share-456",
// 			},
// 		},
// 		{
// 			name: "subscription without share",
// 			msg: websocket_mod.WsResponseInfo{
// 				SentTime:      9876543210000,
// 				BroadcastType: websocket_mod.SubscriptionType("media"),
// 				SubscribeKey:  "media-789",
// 				Content:       websocket_mod.WsData{},
// 			},
// 			expected: websocket_mod.SubscriptionInfo{
// 				When:           time.UnixMilli(9876543210000),
// 				Type:           websocket_mod.SubscriptionType("media"),
// 				SubscriptionID: "media-789",
// 				ShareID:        "",
// 			},
// 		},
// 		{
// 			name: "subscription with nil content",
// 			msg: websocket_mod.WsResponseInfo{
// 				SentTime:      1000000000000,
// 				BroadcastType: websocket_mod.SubscriptionType("task"),
// 				SubscribeKey:  "task-000",
// 				Content:       nil,
// 			},
// 			expected: websocket_mod.SubscriptionInfo{
// 				When:           time.UnixMilli(1000000000000),
// 				Type:           websocket_mod.SubscriptionType("task"),
// 				SubscriptionID: "task-000",
// 				ShareID:        "",
// 			},
// 		},
// 		{
// 			name: "zero time subscription",
// 			msg: websocket_mod.WsResponseInfo{
// 				SentTime:      0,
// 				BroadcastType: websocket_mod.SubscriptionType("test"),
// 				SubscribeKey:  "test-key",
// 				Content:       websocket_mod.WsData{"shareID": "test-share"},
// 			},
// 			expected: websocket_mod.SubscriptionInfo{
// 				When:           time.UnixMilli(0),
// 				Type:           websocket_mod.SubscriptionType("test"),
// 				SubscriptionID: "test-key",
// 				ShareID:        "test-share",
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.GetSubscribeInfo(tt.msg)
//
// 			if !result.When.Equal(tt.expected.When) {
// 				t.Errorf("When = %v, want %v", result.When, tt.expected.When)
// 			}
// 			if result.Type != tt.expected.Type {
// 				t.Errorf("Type = %v, want %v", result.Type, tt.expected.Type)
// 			}
// 			if result.SubscriptionID != tt.expected.SubscriptionID {
// 				t.Errorf("SubscriptionID = %v, want %v", result.SubscriptionID, tt.expected.SubscriptionID)
// 			}
// 			if result.ShareID != tt.expected.ShareID {
// 				t.Errorf("ShareID = %v, want %v", result.ShareID, tt.expected.ShareID)
// 			}
// 		})
// 	}
// }
//
// func TestGetCancelInfo(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		msg      websocket_mod.WsResponseInfo
// 		expected websocket_mod.CancelInfo
// 	}{
// 		{
// 			name: "basic cancel",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{"taskID": "task-123"},
// 			},
// 			expected: websocket_mod.CancelInfo{
// 				TaskID: "task-123",
// 			},
// 		},
// 		{
// 			name: "cancel with other fields",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{
// 					"taskID": "task-456",
// 					"other":  "value",
// 					"count":  42,
// 				},
// 			},
// 			expected: websocket_mod.CancelInfo{
// 				TaskID: "task-456",
// 			},
// 		},
// 		{
// 			name: "cancel with empty task ID",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{"taskID": ""},
// 			},
// 			expected: websocket_mod.CancelInfo{
// 				TaskID: "",
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.GetCancelInfo(tt.msg)
//
// 			if result.TaskID != tt.expected.TaskID {
// 				t.Errorf("TaskID = %v, want %v", result.TaskID, tt.expected.TaskID)
// 			}
// 		})
// 	}
// }
//
// func TestGetScanInfo(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		msg      websocket_mod.WsResponseInfo
// 		expected websocket_mod.ScanInfo
// 	}{
// 		{
// 			name: "scan with folderID",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{
// 					"folderID": "folder-123",
// 					"shareID":  "share-456",
// 				},
// 			},
// 			expected: websocket_mod.ScanInfo{
// 				FileID:  "folder-123",
// 				ShareID: "share-456",
// 			},
// 		},
// 		{
// 			name: "scan with fileID fallback",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{
// 					"fileID":  "file-789",
// 					"shareID": "share-000",
// 				},
// 			},
// 			expected: websocket_mod.ScanInfo{
// 				FileID:  "file-789",
// 				ShareID: "share-000",
// 			},
// 		},
// 		{
// 			name: "folderID takes precedence over fileID",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{
// 					"folderID": "folder-abc",
// 					"fileID":   "file-xyz",
// 					"shareID":  "share-123",
// 				},
// 			},
// 			expected: websocket_mod.ScanInfo{
// 				FileID:  "folder-abc",
// 				ShareID: "share-123",
// 			},
// 		},
// 		{
// 			name: "scan without shareID",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{
// 					"folderID": "folder-only",
// 				},
// 			},
// 			expected: websocket_mod.ScanInfo{
// 				FileID:  "folder-only",
// 				ShareID: "",
// 			},
// 		},
// 		{
// 			name: "scan with empty content",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: websocket_mod.WsData{},
// 			},
// 			expected: websocket_mod.ScanInfo{
// 				FileID:  "",
// 				ShareID: "",
// 			},
// 		},
// 		{
// 			name: "scan with nil content",
// 			msg: websocket_mod.WsResponseInfo{
// 				Content: nil,
// 			},
// 			expected: websocket_mod.ScanInfo{
// 				FileID:  "",
// 				ShareID: "",
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.GetScanInfo(tt.msg)
//
// 			if result.FileID != tt.expected.FileID {
// 				t.Errorf("FileID = %v, want %v", result.FileID, tt.expected.FileID)
// 			}
// 			if result.ShareID != tt.expected.ShareID {
// 				t.Errorf("ShareID = %v, want %v", result.ShareID, tt.expected.ShareID)
// 			}
// 		})
// 	}
// }
//
// // ============================================================================
// // Edge case tests
// // ============================================================================
//
// func TestMediaToMediaInfo_WithRecognitionTags(t *testing.T) {
// 	mediaID := primitive.NewObjectID()
// 	now := time.Now()
//
// 	media := &media_model.Media{
// 		MediaID:    mediaID,
// 		ContentID:  "taggedContent",
// 		FileIDs:    []string{"file1"},
// 		CreateDate: now,
// 		Owner:      "taguser",
// 		Width:      1920,
// 		Height:     1080,
// 		MimeType:   "image/jpeg",
// 		Enabled:    true,
// 	}
// 	media.SetRecognitionTags([]string{"cat", "dog", "person"})
//
// 	result := reshape.MediaToMediaInfo(media)
//
// 	if len(result.RecognitionTags) != 3 {
// 		t.Errorf("len(RecognitionTags) = %v, want 3", len(result.RecognitionTags))
// 	}
//
// 	expectedTags := []string{"cat", "dog", "person"}
// 	for i, tag := range expectedTags {
// 		if i < len(result.RecognitionTags) && result.RecognitionTags[i] != tag {
// 			t.Errorf("RecognitionTags[%d] = %v, want %v", i, result.RecognitionTags[i], tag)
// 		}
// 	}
// }
//
// func TestTokenWithShortTokenBytes(t *testing.T) {
// 	now := time.Now().Truncate(time.Millisecond)
// 	tokenID := primitive.NewObjectID()
//
// 	// Create a shorter token string (less than 32 bytes after decode)
// 	shortTokenBytes := []byte{1, 2, 3, 4, 5}
// 	shortTokenStr := base64.StdEncoding.EncodeToString(shortTokenBytes)
//
// 	info := structs.TokenInfo{
// 		ID:          tokenID.Hex(),
// 		CreatedTime: now.UnixMilli(),
// 		LastUsed:    now.UnixMilli(),
// 		Nickname:    "short-token",
// 		Owner:       "user",
// 		Token:       shortTokenStr,
// 	}
//
// 	result, err := reshape.TokenInfoToToken(context.Background(), info)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
//
// 	// The first 5 bytes should match
// 	for i := 0; i < 5; i++ {
// 		if result.Token[i] != shortTokenBytes[i] {
// 			t.Errorf("Token[%d] = %v, want %v", i, result.Token[i], shortTokenBytes[i])
// 		}
// 	}
//
// 	// The rest should be zero
// 	for i := 5; i < 32; i++ {
// 		if result.Token[i] != 0 {
// 			t.Errorf("Token[%d] = %v, want 0", i, result.Token[i])
// 		}
// 	}
// }
//
// func TestFileActionWithLargeSize(t *testing.T) {
// 	now := time.Now().Truncate(time.Millisecond)
// 	filepath := fs.BuildFilePath("USERS", "testuser/largefile.bin")
//
// 	// Test with max int64 size
// 	action := history.FileAction{
// 		ActionType: history.FileCreate,
// 		FileID:     "largefile",
// 		Filepath:   filepath,
// 		EventID:    "event-large",
// 		TowerID:    "tower-large",
// 		Timestamp:  now,
// 		Size:       int64(1<<62 - 1), // Very large size
// 	}
//
// 	info := reshape.FileActionToFileActionInfo(action)
//
// 	if info.Size != action.Size {
// 		t.Errorf("Size mismatch for large file: got %v, want %v", info.Size, action.Size)
// 	}
//
// 	// Round trip test
// 	result := reshape.FileActionInfoToFileAction(info)
//
// 	if result.Size != action.Size {
// 		t.Errorf("Size mismatch after round trip: got %v, want %v", result.Size, action.Size)
// 	}
// }
//
// func TestNewMediaBatchInfoPreservesOrder(t *testing.T) {
// 	now := time.Now()
//
// 	media := []*media_model.Media{
// 		{MediaID: primitive.NewObjectID(), ContentID: "first", CreateDate: now, Owner: "user1", Enabled: true},
// 		{MediaID: primitive.NewObjectID(), ContentID: "second", CreateDate: now, Owner: "user2", Enabled: true},
// 		{MediaID: primitive.NewObjectID(), ContentID: "third", CreateDate: now, Owner: "user3", Enabled: true},
// 	}
//
// 	result := reshape.NewMediaBatchInfo(media)
//
// 	if len(result.Media) != 3 {
// 		t.Fatalf("expected 3 media items, got %d", len(result.Media))
// 	}
//
// 	expectedOrder := []string{"first", "second", "third"}
// 	for i, expected := range expectedOrder {
// 		if result.Media[i].ContentID != expected {
// 			t.Errorf("Media[%d].ContentID = %v, want %v", i, result.Media[i].ContentID, expected)
// 		}
// 	}
// }
//
// // ============================================================================
// // task.go tests
// // ============================================================================
//
// func TestTaskToTaskInfo(t *testing.T) {
// 	now := time.Now()
//
// 	tests := []struct {
// 		name     string
// 		task     *task_model.Task
// 		expected structs.TaskInfo
// 	}{
// 		{
// 			name: "basic completed task",
// 			task: task_model.NewTestTask(
// 				"task-123",
// 				"scan_directory",
// 				1,
// 				task_mod.TaskSuccess,
// 				task_mod.Result{"filesScanned": 100},
// 				now,
// 			),
// 			expected: structs.TaskInfo{
// 				TaskID:    "task-123",
// 				JobName:   "scan_directory",
// 				Progress:  0,
// 				Status:    task_mod.TaskSuccess,
// 				Completed: true,
// 				WorkerID:  1,
// 				Result:    task_mod.Result{"filesScanned": 100},
// 				StartTime: now,
// 			},
// 		},
// 		{
// 			name: "failed task",
// 			task: task_model.NewTestTask(
// 				"task-456",
// 				"process_media",
// 				2,
// 				task_mod.TaskError,
// 				task_mod.Result{"error": "file not found"},
// 				now,
// 			),
// 			expected: structs.TaskInfo{
// 				TaskID:    "task-456",
// 				JobName:   "process_media",
// 				Progress:  0,
// 				Status:    task_mod.TaskError,
// 				Completed: true,
// 				WorkerID:  2,
// 				Result:    task_mod.Result{"error": "file not found"},
// 				StartTime: now,
// 			},
// 		},
// 		{
// 			name: "cancelled task",
// 			task: task_model.NewTestTask(
// 				"task-789",
// 				"backup_files",
// 				3,
// 				task_mod.TaskCanceled,
// 				nil,
// 				now,
// 			),
// 			expected: structs.TaskInfo{
// 				TaskID:    "task-789",
// 				JobName:   "backup_files",
// 				Progress:  0,
// 				Status:    task_mod.TaskCanceled,
// 				Completed: true,
// 				WorkerID:  3,
// 				Result:    task_mod.Result{},
// 				StartTime: now,
// 			},
// 		},
// 		{
// 			name: "task with zero start time",
// 			task: task_model.NewTestTask(
// 				"task-000",
// 				"quick_task",
// 				0,
// 				task_mod.TaskSuccess,
// 				nil,
// 				time.Time{},
// 			),
// 			expected: structs.TaskInfo{
// 				TaskID:    "task-000",
// 				JobName:   "quick_task",
// 				Progress:  0,
// 				Status:    task_mod.TaskSuccess,
// 				Completed: true,
// 				WorkerID:  0,
// 				Result:    task_mod.Result{},
// 				StartTime: time.Time{},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.TaskToTaskInfo(tt.task)
//
// 			if result.TaskID != tt.expected.TaskID {
// 				t.Errorf("TaskID = %v, want %v", result.TaskID, tt.expected.TaskID)
// 			}
// 			if result.JobName != tt.expected.JobName {
// 				t.Errorf("JobName = %v, want %v", result.JobName, tt.expected.JobName)
// 			}
// 			if result.Progress != tt.expected.Progress {
// 				t.Errorf("Progress = %v, want %v", result.Progress, tt.expected.Progress)
// 			}
// 			if result.Status != tt.expected.Status {
// 				t.Errorf("Status = %v, want %v", result.Status, tt.expected.Status)
// 			}
// 			if result.Completed != tt.expected.Completed {
// 				t.Errorf("Completed = %v, want %v", result.Completed, tt.expected.Completed)
// 			}
// 			if result.WorkerID != tt.expected.WorkerID {
// 				t.Errorf("WorkerID = %v, want %v", result.WorkerID, tt.expected.WorkerID)
// 			}
// 			if !result.StartTime.Equal(tt.expected.StartTime) {
// 				t.Errorf("StartTime = %v, want %v", result.StartTime, tt.expected.StartTime)
// 			}
// 			// Check result - both should be nil or both should be maps
// 			if result.Result == nil && tt.expected.Result != nil {
// 				t.Errorf("Result = nil, want %v", tt.expected.Result)
// 			}
// 			if result.Result != nil && tt.expected.Result == nil {
// 				t.Errorf("Result = %v, want nil", result.Result)
// 			}
// 		})
// 	}
// }
//
// func TestTasksToTaskInfos(t *testing.T) {
// 	now := time.Now()
//
// 	tests := []struct {
// 		name          string
// 		tasks         []*task_model.Task
// 		expectedCount int
// 	}{
// 		{
// 			name:          "empty tasks",
// 			tasks:         []*task_model.Task{},
// 			expectedCount: 0,
// 		},
// 		{
// 			name: "single task",
// 			tasks: []*task_model.Task{
// 				task_model.NewTestTask("task-1", "job1", 1, task_mod.TaskSuccess, nil, now),
// 			},
// 			expectedCount: 1,
// 		},
// 		{
// 			name: "multiple tasks",
// 			tasks: []*task_model.Task{
// 				task_model.NewTestTask("task-1", "job1", 1, task_mod.TaskSuccess, nil, now),
// 				task_model.NewTestTask("task-2", "job2", 2, task_mod.TaskError, nil, now),
// 				task_model.NewTestTask("task-3", "job3", 3, task_mod.TaskCanceled, nil, now),
// 			},
// 			expectedCount: 3,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := reshape.TasksToTaskInfos(tt.tasks)
//
// 			if len(result) != tt.expectedCount {
// 				t.Errorf("len(result) = %v, want %v", len(result), tt.expectedCount)
// 			}
//
// 			// Verify each task is correctly converted
// 			for i, taskInfo := range result {
// 				if taskInfo.TaskID != tt.tasks[i].ID() {
// 					t.Errorf("result[%d].TaskID = %v, want %v", i, taskInfo.TaskID, tt.tasks[i].ID())
// 				}
// 				if taskInfo.JobName != tt.tasks[i].JobName() {
// 					t.Errorf("result[%d].JobName = %v, want %v", i, taskInfo.JobName, tt.tasks[i].JobName())
// 				}
// 			}
// 		})
// 	}
// }

// ============================================================================
// user.go tests (with AppContext mock)
// ============================================================================

// func TestUserToUserInfo(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		user        *user_model.User
// 		onlineUsers []string
// 		expected    structs.UserInfo
// 	}{
// 		{
// 			name: "basic user offline",
// 			user: &user_model.User{
// 				Username:    "testuser",
// 				DisplayName: "Test User",
// 				HomeID:      "home123",
// 				TrashID:     "trash123",
// 				UserPerms:   user_model.UserPermissionBasic,
// 				Activated:   true,
// 			},
// 			onlineUsers: []string{},
// 			expected: structs.UserInfo{
// 				Username:        "testuser",
// 				FullName:        "Test User",
// 				HomeID:          "home123",
// 				TrashID:         "trash123",
// 				PermissionLevel: int(user_model.UserPermissionBasic),
// 				Activated:       true,
// 				IsOnline:        false,
// 			},
// 		},
// 		{
// 			name: "admin user online",
// 			user: &user_model.User{
// 				Username:    "adminuser",
// 				DisplayName: "Admin User",
// 				HomeID:      "adminHome",
// 				TrashID:     "adminTrash",
// 				UserPerms:   user_model.UserPermissionAdmin,
// 				Activated:   true,
// 			},
// 			onlineUsers: []string{"adminuser"},
// 			expected: structs.UserInfo{
// 				Username:        "adminuser",
// 				FullName:        "Admin User",
// 				HomeID:          "adminHome",
// 				TrashID:         "adminTrash",
// 				PermissionLevel: int(user_model.UserPermissionAdmin),
// 				Activated:       true,
// 				IsOnline:        true,
// 			},
// 		},
// 		{
// 			name: "inactive user",
// 			user: &user_model.User{
// 				Username:    "inactiveuser",
// 				DisplayName: "Inactive User",
// 				HomeID:      "",
// 				TrashID:     "",
// 				UserPerms:   user_model.UserPermissionBasic,
// 				Activated:   false,
// 			},
// 			onlineUsers: []string{},
// 			expected: structs.UserInfo{
// 				Username:        "inactiveuser",
// 				FullName:        "Inactive User",
// 				HomeID:          "",
// 				TrashID:         "",
// 				PermissionLevel: int(user_model.UserPermissionBasic),
// 				Activated:       false,
// 				IsOnline:        false,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var opts []testAppContextOption
// 			for _, u := range tt.onlineUsers {
// 				opts = append(opts, withOnlineUser(u))
// 			}
//
// 			ctx := newTestAppContext(t, opts...)
// 			result := reshape.UserToUserInfo(ctx, tt.user)
//
// 			if result.Username != tt.expected.Username {
// 				t.Errorf("Username = %v, want %v", result.Username, tt.expected.Username)
// 			}
// 			if result.FullName != tt.expected.FullName {
// 				t.Errorf("FullName = %v, want %v", result.FullName, tt.expected.FullName)
// 			}
// 			if result.HomeID != tt.expected.HomeID {
// 				t.Errorf("HomeID = %v, want %v", result.HomeID, tt.expected.HomeID)
// 			}
// 			if result.TrashID != tt.expected.TrashID {
// 				t.Errorf("TrashID = %v, want %v", result.TrashID, tt.expected.TrashID)
// 			}
// 			if result.PermissionLevel != tt.expected.PermissionLevel {
// 				t.Errorf("PermissionLevel = %v, want %v", result.PermissionLevel, tt.expected.PermissionLevel)
// 			}
// 			if result.Activated != tt.expected.Activated {
// 				t.Errorf("Activated = %v, want %v", result.Activated, tt.expected.Activated)
// 			}
// 			if result.IsOnline != tt.expected.IsOnline {
// 				t.Errorf("IsOnline = %v, want %v", result.IsOnline, tt.expected.IsOnline)
// 			}
// 		})
// 	}
// }

// func TestUserToUserInfoArchive(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		user        *user_model.User
// 		onlineUsers []string
// 		expected    structs.UserInfoArchive
// 	}{
// 		{
// 			name: "basic user with password",
// 			user: &user_model.User{
// 				Username:    "testuser",
// 				DisplayName: "Test User",
// 				Password:    "hashedpassword123",
// 				HomeID:      "home123",
// 				TrashID:     "trash123",
// 				UserPerms:   user_model.UserPermissionBasic,
// 				Activated:   true,
// 			},
// 			onlineUsers: []string{"testuser"},
// 			expected: structs.UserInfoArchive{
// 				UserInfo: structs.UserInfo{
// 					Username:        "testuser",
// 					FullName:        "Test User",
// 					HomeID:          "home123",
// 					TrashID:         "trash123",
// 					PermissionLevel: int(user_model.UserPermissionBasic),
// 					Activated:       true,
// 					IsOnline:        true,
// 				},
// 				Password: "hashedpassword123",
// 			},
// 		},
// 		{
// 			name:        "nil user returns empty archive",
// 			user:        nil,
// 			onlineUsers: []string{},
// 			expected:    structs.UserInfoArchive{},
// 		},
// 		{
// 			name: "system user returns empty archive",
// 			user: &user_model.User{
// 				Username:  "systemuser",
// 				UserPerms: user_model.UserPermissionSystem,
// 			},
// 			onlineUsers: []string{},
// 			expected:    structs.UserInfoArchive{},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var opts []testAppContextOption
// 			for _, u := range tt.onlineUsers {
// 				opts = append(opts, withOnlineUser(u))
// 			}
//
// 			ctx := newTestAppContext(t, opts...)
// 			result := reshape.UserToUserInfoArchive(ctx, tt.user)
//
// 			if result.Username != tt.expected.Username {
// 				t.Errorf("Username = %v, want %v", result.Username, tt.expected.Username)
// 			}
// 			if result.FullName != tt.expected.FullName {
// 				t.Errorf("FullName = %v, want %v", result.FullName, tt.expected.FullName)
// 			}
// 			if result.Password != tt.expected.Password {
// 				t.Errorf("Password = %v, want %v", result.Password, tt.expected.Password)
// 			}
// 			if result.PermissionLevel != tt.expected.PermissionLevel {
// 				t.Errorf("PermissionLevel = %v, want %v", result.PermissionLevel, tt.expected.PermissionLevel)
// 			}
// 			if result.Activated != tt.expected.Activated {
// 				t.Errorf("Activated = %v, want %v", result.Activated, tt.expected.Activated)
// 			}
// 			if result.IsOnline != tt.expected.IsOnline {
// 				t.Errorf("IsOnline = %v, want %v", result.IsOnline, tt.expected.IsOnline)
// 			}
// 		})
// 	}
// }

// ============================================================================
// tower.go tests (with AppContext mock)
// ============================================================================

// func TestTowerToTowerInfo(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		tower        tower_model.Instance
// 		onlineTowers []string
// 		fileService  func() *mockFileService
// 		expected     structs.TowerInfo
// 	}{
// 		{
// 			name: "local tower (this server) is always online",
// 			tower: tower_model.Instance{
// 				TowerID:     "local-tower-123",
// 				Name:        "Local Tower",
// 				Role:        tower_model.RoleCore,
// 				Address:     "https://localhost:8080",
// 				LastBackup:  1234567890000,
// 				IsThisTower: true,
// 			},
// 			onlineTowers: []string{},
// 			fileService:  nil,
// 			expected: structs.TowerInfo{
// 				ID:           "local-tower-123",
// 				Name:         "Local Tower",
// 				Role:         "core",
// 				Address:      "https://localhost:8080",
// 				LastBackup:   1234567890000,
// 				IsThisServer: true,
// 				Started:      true,
// 				BackupSize:   0,
// 				ReportedRole: "core",
// 				Online:       true,
// 			},
// 		},
// 		{
// 			name: "remote tower online",
// 			tower: tower_model.Instance{
// 				TowerID:     "remote-tower-456",
// 				Name:        "Remote Tower",
// 				Role:        tower_model.RoleBackup,
// 				Address:     "https://remote.example.com",
// 				LastBackup:  9876543210000,
// 				IsThisTower: false,
// 			},
// 			onlineTowers: []string{"remote-tower-456"},
// 			fileService:  nil,
// 			expected: structs.TowerInfo{
// 				ID:           "remote-tower-456",
// 				Name:         "Remote Tower",
// 				Role:         "backup",
// 				Address:      "https://remote.example.com",
// 				LastBackup:   9876543210000,
// 				IsThisServer: false,
// 				Started:      true,
// 				BackupSize:   0,
// 				ReportedRole: "backup",
// 				Online:       true,
// 			},
// 		},
// 		{
// 			name: "remote tower offline",
// 			tower: tower_model.Instance{
// 				TowerID:     "offline-tower-789",
// 				Name:        "Offline Tower",
// 				Role:        tower_model.RoleBackup,
// 				Address:     "https://offline.example.com",
// 				LastBackup:  0,
// 				IsThisTower: false,
// 			},
// 			onlineTowers: []string{},
// 			fileService:  nil,
// 			expected: structs.TowerInfo{
// 				ID:           "offline-tower-789",
// 				Name:         "Offline Tower",
// 				Role:         "backup",
// 				Address:      "https://offline.example.com",
// 				LastBackup:   0,
// 				IsThisServer: false,
// 				Started:      true,
// 				BackupSize:   0,
// 				ReportedRole: "backup",
// 				Online:       false,
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var opts []testAppContextOption
// 			for _, tower := range tt.onlineTowers {
// 				opts = append(opts, withOnlineTower(tower))
// 			}
//
// 			// Add file service if provided
// 			if tt.fileService != nil {
// 				opts = append(opts, withFileService(tt.fileService()))
// 			} else {
// 				opts = append(opts, withFileService(newMockFileService()))
// 			}
//
// 			ctx := newTestAppContext(t, opts...)
// 			result := reshape.TowerToTowerInfo(ctx, tt.tower)
//
// 			if result.ID != tt.expected.ID {
// 				t.Errorf("ID = %v, want %v", result.ID, tt.expected.ID)
// 			}
// 			if result.Name != tt.expected.Name {
// 				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
// 			}
// 			if result.Role != tt.expected.Role {
// 				t.Errorf("Role = %v, want %v", result.Role, tt.expected.Role)
// 			}
// 			if result.Address != tt.expected.Address {
// 				t.Errorf("Address = %v, want %v", result.Address, tt.expected.Address)
// 			}
// 			if result.LastBackup != tt.expected.LastBackup {
// 				t.Errorf("LastBackup = %v, want %v", result.LastBackup, tt.expected.LastBackup)
// 			}
// 			if result.IsThisServer != tt.expected.IsThisServer {
// 				t.Errorf("IsThisServer = %v, want %v", result.IsThisServer, tt.expected.IsThisServer)
// 			}
// 			if result.Started != tt.expected.Started {
// 				t.Errorf("Started = %v, want %v", result.Started, tt.expected.Started)
// 			}
// 			if result.Online != tt.expected.Online {
// 				t.Errorf("Online = %v, want %v", result.Online, tt.expected.Online)
// 			}
// 			if result.ReportedRole != tt.expected.ReportedRole {
// 				t.Errorf("ReportedRole = %v, want %v", result.ReportedRole, tt.expected.ReportedRole)
// 			}
// 		})
// 	}
// }

// ============================================================================
// share.go integration tests (requires DB)
// ============================================================================

// func TestShareToShareInfo(t *testing.T) {
// 	now := time.Now()
// 	shareID := primitive.NewObjectID()
//
// 	tests := []struct {
// 		name                  string
// 		share                 *share_model.FileShare
// 		accessorUsers         []string // Usernames to create in DB
// 		onlineUsers           []string // Users marked as online
// 		expectedAccessorCount int
// 	}{
// 		{
// 			name: "share with no accessors",
// 			share: &share_model.FileShare{
// 				ShareID:     shareID,
// 				FileID:      "file-123",
// 				ShareName:   "test-share",
// 				Owner:       "shareowner",
// 				Accessors:   []string{},
// 				Permissions: map[string]*share_model.Permissions{},
// 				Public:      false,
// 				Wormhole:    false,
// 				Enabled:     true,
// 				Expires:     now.Add(24 * time.Hour),
// 				Updated:     now,
// 			},
// 			accessorUsers:         []string{},
// 			onlineUsers:           []string{},
// 			expectedAccessorCount: 0,
// 		},
// 		{
// 			name: "share with single accessor",
// 			share: &share_model.FileShare{
// 				ShareID:   shareID,
// 				FileID:    "file-456",
// 				ShareName: "shared-folder",
// 				Owner:     "owner1",
// 				Accessors: []string{"accessor1"},
// 				Permissions: map[string]*share_model.Permissions{
// 					"accessor1": {CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 				},
// 				Public:   false,
// 				Wormhole: false,
// 				Enabled:  true,
// 				Expires:  now.Add(7 * 24 * time.Hour),
// 				Updated:  now,
// 			},
// 			accessorUsers:         []string{"accessor1"},
// 			onlineUsers:           []string{"accessor1"},
// 			expectedAccessorCount: 1,
// 		},
// 		{
// 			name: "share with multiple accessors",
// 			share: &share_model.FileShare{
// 				ShareID:   shareID,
// 				FileID:    "file-789",
// 				ShareName: "team-share",
// 				Owner:     "teamowner",
// 				Accessors: []string{"user1", "user2", "user3"},
// 				Permissions: map[string]*share_model.Permissions{
// 					"user1": {CanView: true, CanEdit: true, CanDownload: true, CanDelete: true},
// 					"user2": {CanView: true, CanEdit: false, CanDownload: true, CanDelete: false},
// 					"user3": {CanView: true, CanEdit: false, CanDownload: false, CanDelete: false},
// 				},
// 				Public:   false,
// 				Wormhole: false,
// 				Enabled:  true,
// 				Expires:  now.Add(30 * 24 * time.Hour),
// 				Updated:  now,
// 			},
// 			accessorUsers:         []string{"user1", "user2", "user3"},
// 			onlineUsers:           []string{"user1", "user3"},
// 			expectedAccessorCount: 3,
// 		},
// 		{
// 			name: "public share",
// 			share: &share_model.FileShare{
// 				ShareID:     shareID,
// 				FileID:      "public-file",
// 				ShareName:   "public-share",
// 				Owner:       "publicowner",
// 				Accessors:   []string{},
// 				Permissions: map[string]*share_model.Permissions{},
// 				Public:      true,
// 				Wormhole:    false,
// 				Enabled:     true,
// 				Expires:     now.Add(365 * 24 * time.Hour),
// 				Updated:     now,
// 			},
// 			accessorUsers:         []string{},
// 			onlineUsers:           []string{},
// 			expectedAccessorCount: 0,
// 		},
// 		{
// 			name: "wormhole share",
// 			share: &share_model.FileShare{
// 				ShareID:     shareID,
// 				FileID:      "wormhole-file",
// 				ShareName:   "wormhole-share",
// 				Owner:       "wormholeowner",
// 				Accessors:   []string{},
// 				Permissions: map[string]*share_model.Permissions{},
// 				Public:      false,
// 				Wormhole:    true,
// 				Enabled:     true,
// 				Expires:     now.Add(1 * time.Hour),
// 				Updated:     now,
// 			},
// 			accessorUsers:         []string{},
// 			onlineUsers:           []string{},
// 			expectedAccessorCount: 0,
// 		},
// 		{
// 			name: "share with zero ObjectID",
// 			share: &share_model.FileShare{
// 				ShareID:     primitive.NilObjectID,
// 				FileID:      "zero-id-file",
// 				ShareName:   "zero-share",
// 				Owner:       "zeroowner",
// 				Accessors:   []string{},
// 				Permissions: map[string]*share_model.Permissions{},
// 				Public:      false,
// 				Enabled:     true,
// 				Expires:     now,
// 				Updated:     now,
// 			},
// 			accessorUsers:         []string{},
// 			onlineUsers:           []string{},
// 			expectedAccessorCount: 0,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Build options for online users
// 			var opts []testAppContextOption
// 			for _, u := range tt.onlineUsers {
// 				opts = append(opts, withOnlineUser(u))
// 			}
//
// 			// Create RequestContext with DB for tests that need accessor lookup
// 			reqCtx := newTestRequestContextWithDB(t, opts...)
//
// 			// Create test users in DB
// 			for _, username := range tt.accessorUsers {
// 				createTestUser(reqCtx, t, username)
// 			}
//
// 			// Call the function under test
// 			result := reshape.ShareToShareInfo(reqCtx, tt.share)
//
// 			// Verify basic fields
// 			if tt.share.ShareID.IsZero() {
// 				if result.ShareID != "" {
// 					t.Errorf("ShareID = %v, want empty for zero ObjectID", result.ShareID)
// 				}
// 			} else {
// 				if result.ShareID != tt.share.ShareID.Hex() {
// 					t.Errorf("ShareID = %v, want %v", result.ShareID, tt.share.ShareID.Hex())
// 				}
// 			}
//
// 			if result.FileID != tt.share.FileID {
// 				t.Errorf("FileID = %v, want %v", result.FileID, tt.share.FileID)
// 			}
//
// 			if result.ShareName != tt.share.ShareName {
// 				t.Errorf("ShareName = %v, want %v", result.ShareName, tt.share.ShareName)
// 			}
//
// 			if result.Owner != tt.share.Owner {
// 				t.Errorf("Owner = %v, want %v", result.Owner, tt.share.Owner)
// 			}
//
// 			if len(result.Accessors) != tt.expectedAccessorCount {
// 				t.Errorf("len(Accessors) = %v, want %v", len(result.Accessors), tt.expectedAccessorCount)
// 			}
//
// 			if result.Public != tt.share.Public {
// 				t.Errorf("Public = %v, want %v", result.Public, tt.share.Public)
// 			}
//
// 			if result.Wormhole != tt.share.Wormhole {
// 				t.Errorf("Wormhole = %v, want %v", result.Wormhole, tt.share.Wormhole)
// 			}
//
// 			if result.Enabled != tt.share.Enabled {
// 				t.Errorf("Enabled = %v, want %v", result.Enabled, tt.share.Enabled)
// 			}
//
// 			if result.Expires != tt.share.Expires.UnixMilli() {
// 				t.Errorf("Expires = %v, want %v", result.Expires, tt.share.Expires.UnixMilli())
// 			}
//
// 			if result.Updated != tt.share.Updated.UnixMilli() {
// 				t.Errorf("Updated = %v, want %v", result.Updated, tt.share.Updated.UnixMilli())
// 			}
//
// 			// Verify permissions were converted
// 			if len(result.Permissions) != len(tt.share.Permissions) {
// 				t.Errorf("len(Permissions) = %v, want %v", len(result.Permissions), len(tt.share.Permissions))
// 			}
// 		})
// 	}
// }
//
// // ============================================================================
// // file.go tests (integration - requires MongoDB)
// // ============================================================================
//
// func TestWeblensFileToFileInfo(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		setup    func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl
// 		opts     []reshape.FileInfoOptions
// 		validate func(t *testing.T, result structs.FileInfo, err error)
// 	}{
// 		{
// 			name: "basic file in USERS tree",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				return file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "testuser/photo.jpg",
// 					Size:     1024,
// 				})
// 			},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if result.Owner != "testuser" {
// 					t.Errorf("Owner = %v, want testuser", result.Owner)
// 				}
// 				if result.IsDir {
// 					t.Errorf("IsDir = true, want false")
// 				}
// 				// Size is 0 for MemOnly files because Exists() returns false
// 				if result.Size != 0 {
// 					t.Errorf("Size = %v, want 0 (MemOnly files don't exist on filesystem)", result.Size)
// 				}
// 			},
// 		},
// 		{
// 			name: "directory with children",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				child1 := file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "testuser/folder/file1.txt",
// 					ID:       "child1-id",
// 				})
// 				child2 := file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "testuser/folder/file2.txt",
// 					ID:       "child2-id",
// 				})
// 				return file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "testuser/folder",
// 					IsDir:    true,
// 					Children: []*file_model.WeblensFileImpl{child1, child2},
// 				})
// 			},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if !result.IsDir {
// 					t.Errorf("IsDir = false, want true")
// 				}
// 				if len(result.Children) != 2 {
// 					t.Errorf("len(Children) = %v, want 2", len(result.Children))
// 				}
// 			},
// 		},
// 		{
// 			name: "file with share",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				f := file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "shareowner/shared-file.txt",
// 					ID:       "shared-file-id",
// 				})
// 				// Create a share for this file
// 				share, err := share_model.NewFileShare(ctx, f.ID(), &user_model.User{Username: "shareowner"}, nil, false, false, false)
// 				if err != nil {
// 					t.Fatalf("failed to create share: %v", err)
// 				}
// 				err = share_model.SaveFileShare(ctx, share)
// 				if err != nil {
// 					t.Fatalf("failed to save share: %v", err)
// 				}
// 				return f
// 			},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if result.ShareID == "" {
// 					t.Errorf("ShareID is empty, expected non-empty")
// 				}
// 			},
// 		},
// 		{
// 			name: "file without share",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				return file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "user/no-share-file.txt",
// 					ID:       "no-share-file-id",
// 				})
// 			},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if result.ShareID != "" {
// 					t.Errorf("ShareID = %v, want empty", result.ShareID)
// 				}
// 			},
// 		},
// 		{
// 			name: "past file is not modifiable",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				return file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName:   "USERS",
// 					RelPath:    "testuser/past-file.txt",
// 					IsPastFile: true,
// 					PastID:     "original-file-id",
// 				})
// 			},
// 			opts: []reshape.FileInfoOptions{{IsPastFile: true}},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if !result.PastFile {
// 					t.Errorf("PastFile = false, want true")
// 				}
// 				if result.Modifiable {
// 					t.Errorf("Modifiable = true, want false for past file")
// 				}
// 				if result.PastID != "original-file-id" {
// 					t.Errorf("PastID = %v, want original-file-id", result.PastID)
// 				}
// 			},
// 		},
// 		{
// 			name: "file with contentID",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				return file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName:  "USERS",
// 					RelPath:   "testuser/media.jpg",
// 					ContentID: "content-hash-abc123",
// 				})
// 			},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if result.ContentID != "content-hash-abc123" {
// 					t.Errorf("ContentID = %v, want content-hash-abc123", result.ContentID)
// 				}
// 			},
// 		},
// 		{
// 			name: "modifiable override to false",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				return file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "testuser/override-file.txt",
// 				})
// 			},
// 			opts: []reshape.FileInfoOptions{{ModifiableOverride: option.Of(false)}},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if result.Modifiable {
// 					t.Errorf("Modifiable = true, want false due to override")
// 				}
// 			},
// 		},
// 		{
// 			name: "file with permissions restricting edit",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				return file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "testuser/restricted-file.txt",
// 				})
// 			},
// 			opts: []reshape.FileInfoOptions{{
// 				Perms: option.Of(share_model.Permissions{
// 					CanView:     true,
// 					CanEdit:     false,
// 					CanDownload: true,
// 					CanDelete:   false,
// 				}),
// 			}},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if result.Modifiable {
// 					t.Errorf("Modifiable = true, want false due to permissions")
// 				}
// 			},
// 		},
// 		{
// 			name: "directory with cover photo",
// 			setup: func(t *testing.T, ctx context.Context) *file_model.WeblensFileImpl {
// 				f := file_model.NewTestFile(file_model.TestFileOptions{
// 					RootName: "USERS",
// 					RelPath:  "testuser/album",
// 					IsDir:    true,
// 					ID:       "album-folder-id",
// 				})
// 				// Set cover photo for this folder
// 				_, err := cover_model.SetCoverPhoto(ctx, f.ID(), "cover-content-id")
// 				if err != nil {
// 					t.Fatalf("failed to set cover photo: %v", err)
// 				}
// 				return f
// 			},
// 			validate: func(t *testing.T, result structs.FileInfo, err error) {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if result.ContentID != "cover-content-id" {
// 					t.Errorf("ContentID = %v, want cover-content-id", result.ContentID)
// 				}
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Setup test DB for each subtest to get clean state
// 			subCtx := db.SetupTestDB(t, share_model.ShareCollectionKey)
// 			// Also setup cover collection
// 			_ = db.SetupTestDB(t, cover_model.CoverPhotoCollectionKey)
//
// 			f := tt.setup(t, subCtx)
// 			result, err := reshape.WeblensFileToFileInfo(subCtx, f, tt.opts...)
// 			tt.validate(t, result, err)
// 		})
// 	}
// }
