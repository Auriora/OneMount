package fs

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT_UR_01_01_UploadSession_UpdateProgress tests that updateProgress correctly updates session state
func TestUT_UR_01_01_UploadSession_UpdateProgress(t *testing.T) {
	// Create a test upload session
	session := &UploadSession{
		ID:                  "test-id",
		Size:                1024 * 1024, // 1MB
		LastSuccessfulChunk: -1,
		BytesUploaded:       0,
	}

	// Update progress for chunk 0
	session.updateProgress(0, 512*1024) // 512KB

	// Verify the progress was updated
	assert.Equal(t, 0, session.LastSuccessfulChunk)
	assert.Equal(t, uint64(512*1024), session.BytesUploaded)
	assert.True(t, time.Since(session.LastProgressTime) < time.Second)
}

// TestUT_UR_01_02_UploadSession_CanResumeUpload tests the resume capability check
func TestUT_UR_01_02_UploadSession_CanResumeUpload(t *testing.T) {
	// Test case 1: Session that can be resumed
	session := &UploadSession{
		CanResume:           true,
		LastSuccessfulChunk: 2,
		UploadURL:           "https://example.com/upload",
	}
	assert.True(t, session.canResumeUpload())

	// Test case 2: Session that cannot be resumed (no upload URL)
	session.UploadURL = ""
	assert.False(t, session.canResumeUpload())

	// Test case 3: Session that cannot be resumed (not marked as resumable)
	session.UploadURL = "https://example.com/upload"
	session.CanResume = false
	assert.False(t, session.canResumeUpload())

	// Test case 4: Session that cannot be resumed (no successful chunks)
	session.CanResume = true
	session.LastSuccessfulChunk = -1
	assert.False(t, session.canResumeUpload())
}

// TestUT_UR_01_03_UploadSession_GetResumeOffset tests the resume offset calculation
func TestUT_UR_01_03_UploadSession_GetResumeOffset(t *testing.T) {
	session := &UploadSession{
		LastSuccessfulChunk: 2, // Chunks 0, 1, 2 completed
	}

	expectedOffset := uint64(3) * uploadChunkSize // Should resume from chunk 3
	assert.Equal(t, expectedOffset, session.getResumeOffset())

	// Test with no successful chunks
	session.LastSuccessfulChunk = -1
	assert.Equal(t, uint64(0), session.getResumeOffset())
}

// TestUT_UR_01_04_UploadSession_MarkAsResumable tests marking session as resumable
func TestUT_UR_01_04_UploadSession_MarkAsResumable(t *testing.T) {
	session := &UploadSession{
		Size: 25 * 1024 * 1024, // 25MB
	}

	session.markAsResumable()

	assert.True(t, session.CanResume)
	assert.Equal(t, -1, session.LastSuccessfulChunk)

	expectedChunks := int(math.Ceil(float64(25*1024*1024) / float64(uploadChunkSize)))
	assert.Equal(t, expectedChunks, session.TotalChunks)
}

// TestUT_UR_01_05_UploadSession_JSONSerialization tests JSON serialization with new fields
func TestUT_UR_01_05_UploadSession_JSONSerialization(t *testing.T) {
	session := &UploadSession{
		ID:                  "test-id",
		Name:                "test-file.txt",
		Size:                1024 * 1024,
		LastSuccessfulChunk: 5,
		TotalChunks:         10,
		BytesUploaded:       512 * 1024,
		LastProgressTime:    time.Now(),
		RecoveryAttempts:    2,
		CanResume:           true,
		UploadURL:           "https://example.com/upload",
	}

	// Serialize to JSON
	data, err := json.Marshal(session)
	require.NoError(t, err)

	// Deserialize from JSON
	var restored UploadSession
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, session.ID, restored.ID)
	assert.Equal(t, session.Name, restored.Name)
	assert.Equal(t, session.Size, restored.Size)
	assert.Equal(t, session.LastSuccessfulChunk, restored.LastSuccessfulChunk)
	assert.Equal(t, session.TotalChunks, restored.TotalChunks)
	assert.Equal(t, session.BytesUploaded, restored.BytesUploaded)
	assert.Equal(t, session.RecoveryAttempts, restored.RecoveryAttempts)
	assert.Equal(t, session.CanResume, restored.CanResume)
	assert.Equal(t, session.UploadURL, restored.UploadURL)
}

// TestUT_UR_02_01_DownloadSession_UpdateProgress tests download progress tracking
func TestUT_UR_02_01_DownloadSession_UpdateProgress(t *testing.T) {
	session := &DownloadSession{
		ID:                  "test-download-id",
		Size:                2 * 1024 * 1024, // 2MB
		LastSuccessfulChunk: -1,
		BytesDownloaded:     0,
	}

	// Update progress for chunk 0
	session.updateProgress(0, 1024*1024) // 1MB

	// Verify the progress was updated
	assert.Equal(t, 0, session.LastSuccessfulChunk)
	assert.Equal(t, uint64(1024*1024), session.BytesDownloaded)
	assert.True(t, time.Since(session.LastProgressTime) < time.Second)
}

// TestUT_UR_02_02_DownloadSession_CanResumeDownload tests download resume capability
func TestUT_UR_02_02_DownloadSession_CanResumeDownload(t *testing.T) {
	// Test case 1: Session that can be resumed
	session := &DownloadSession{
		CanResume:           true,
		LastSuccessfulChunk: 1,
		TotalChunks:         4,
	}
	assert.True(t, session.canResumeDownload())

	// Test case 2: Session that cannot be resumed (no chunks completed yet)
	session.LastSuccessfulChunk = -1
	assert.False(t, session.canResumeDownload())

	// Test case 3: Session that cannot be resumed (total chunks unknown)
	session.LastSuccessfulChunk = 1
	session.TotalChunks = 0
	assert.False(t, session.canResumeDownload())

	// Test case 4: Session that cannot be resumed (not marked as resumable)
	session.TotalChunks = 4
	session.CanResume = false
	assert.False(t, session.canResumeDownload())
}

// TestUT_UR_02_03_DownloadSession_GetResumeOffset tests download resume offset calculation
func TestUT_UR_02_03_DownloadSession_GetResumeOffset(t *testing.T) {
	session := &DownloadSession{
		LastSuccessfulChunk: 1, // Chunks 0, 1 completed
		ChunkSize:           downloadChunkSize,
	}

	expectedOffset := uint64(2) * downloadChunkSize // Should resume from chunk 2
	assert.Equal(t, expectedOffset, session.getResumeOffset())

	// Test with no successful chunks
	session.LastSuccessfulChunk = -1
	assert.Equal(t, uint64(0), session.getResumeOffset())
}

// TestUT_UR_02_04_DownloadSession_MarkAsResumable tests marking download session as resumable
func TestUT_UR_02_04_DownloadSession_MarkAsResumable(t *testing.T) {
	session := &DownloadSession{
		ID: "test-download-id",
	}

	size := uint64(5 * 1024 * 1024) // 5MB
	chunkSize := downloadChunkSize

	session.markAsResumable(size, chunkSize)

	assert.True(t, session.CanResume)
	assert.Equal(t, size, session.Size)
	assert.Equal(t, chunkSize, session.ChunkSize)
	assert.Equal(t, -1, session.LastSuccessfulChunk)

	expectedChunks := int(math.Ceil(float64(size) / float64(chunkSize)))
	assert.Equal(t, expectedChunks, session.TotalChunks)
}

// TestUT_UR_03_01_RecoveryPersistence tests that recovery state is persisted correctly
func TestUT_UR_03_01_RecoveryPersistence(t *testing.T) {
	// This test would require a mock database or test database setup
	// For now, we'll test the JSON serialization which is used for persistence

	session := &UploadSession{
		ID:                  "persistence-test",
		LastSuccessfulChunk: 3,
		TotalChunks:         10,
		BytesUploaded:       3 * uploadChunkSize,
		RecoveryAttempts:    1,
		CanResume:           true,
	}

	// Simulate persistence by marshaling and unmarshaling
	data, err := json.Marshal(session)
	require.NoError(t, err)

	var restored UploadSession
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Verify recovery state is preserved
	assert.Equal(t, session.LastSuccessfulChunk, restored.LastSuccessfulChunk)
	assert.Equal(t, session.TotalChunks, restored.TotalChunks)
	assert.Equal(t, session.BytesUploaded, restored.BytesUploaded)
	assert.Equal(t, session.RecoveryAttempts, restored.RecoveryAttempts)
	assert.Equal(t, session.CanResume, restored.CanResume)
}
