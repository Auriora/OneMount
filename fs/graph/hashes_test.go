package graph

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkSHA1(b *testing.B) {
	data, _ := os.ReadFile("dmel.fa")
	for i := 0; i < b.N; i++ {
		SHA1Hash(&data)
	}
}

func BenchmarkSHA256(b *testing.B) {
	data, _ := os.ReadFile("dmel.fa")
	for i := 0; i < b.N; i++ {
		SHA256Hash(&data)
	}
}

func BenchmarkQuickXORHash(b *testing.B) {
	data, _ := os.ReadFile("dmel.fa")
	for i := 0; i < b.N; i++ {
		QuickXORHash(&data)
	}
}

func BenchmarkSHA1Stream(b *testing.B) {
	data, _ := os.Open("dmel.fa")
	for i := 0; i < b.N; i++ {
		SHA1HashStream(data)
	}
}

func BenchmarkSHA256Stream(b *testing.B) {
	data, _ := os.Open("dmel.fa")
	for i := 0; i < b.N; i++ {
		SHA256HashStream(data)
	}
}

func BenchmarkQuickXORHashStream(b *testing.B) {
	data, _ := os.Open("dmel.fa")
	for i := 0; i < b.N; i++ {
		QuickXORHashStream(data)
	}
}

func TestSha1HashReader(t *testing.T) {
	content := []byte("this is some text to hash")
	expected := SHA1Hash(&content)

	reader := bytes.NewReader(content)
	actual := SHA1HashStream(reader)
	assert.Equal(t, expected, actual)
}

func TestQuickXORHashReader(t *testing.T) {
	content := []byte("this is some text to hash")
	expected := QuickXORHash(&content)

	reader := bytes.NewReader(content)
	actual := QuickXORHashStream(reader)
	assert.Equal(t, expected, actual)
}

func TestHashSeekPosition(t *testing.T) {
	t.Parallel()

	// Create a temporary file for testing
	tmp, err := os.CreateTemp("", "onedriverHashTest")
	assert.NoError(t, err, "Failed to create temporary file")
	defer os.Remove(tmp.Name())

	// Write some content to the file
	content := []byte("some test content")
	_, err = io.Copy(tmp, bytes.NewBuffer(content))
	assert.NoError(t, err, "Failed to write to temporary file")
	tmp.Close()

	// Open the file for reading
	file, err := os.Open(tmp.Name())
	assert.NoError(t, err, "Failed to open temporary file")
	defer file.Close()

	// Read a portion of the file to move the seek position
	buffer := make([]byte, 5)
	_, err = file.Read(buffer)
	assert.NoError(t, err, "Failed to read from file")

	// Verify that the seek position is not at the beginning
	currentPos, err := file.Seek(0, io.SeekCurrent)
	assert.NoError(t, err, "Failed to get current position")
	assert.Equal(t, int64(5), currentPos, "File position should be at offset 5")

	// Test that QuickXORHashStream resets the seek position
	quickXORHash := QuickXORHashStream(file)
	assert.Equal(t, QuickXORHash(&content), quickXORHash, "QuickXORHashStream returned incorrect result")

	// Verify that the seek position is reset to the beginning
	currentPos, err = file.Seek(0, io.SeekCurrent)
	assert.NoError(t, err, "Failed to get current position")
	assert.Equal(t, int64(0), currentPos, "File position should be reset to the beginning")

	// Test that SHA1HashStream resets the seek position
	_, err = file.Read(buffer)
	assert.NoError(t, err, "Failed to read from file")
	sha1Hash := SHA1HashStream(file)
	assert.Equal(t, SHA1Hash(&content), sha1Hash, "SHA1HashStream returned incorrect result")

	// Verify that the seek position is reset to the beginning
	currentPos, err = file.Seek(0, io.SeekCurrent)
	assert.NoError(t, err, "Failed to get current position")
	assert.Equal(t, int64(0), currentPos, "File position should be reset to the beginning")

	// Test that SHA256HashStream resets the seek position
	_, err = file.Read(buffer)
	assert.NoError(t, err, "Failed to read from file")
	sha256Hash := SHA256HashStream(file)
	assert.Equal(t, SHA256Hash(&content), sha256Hash, "SHA256HashStream returned incorrect result")

	// Verify that the seek position is reset to the beginning
	currentPos, err = file.Seek(0, io.SeekCurrent)
	assert.NoError(t, err, "Failed to get current position")
	assert.Equal(t, int64(0), currentPos, "File position should be reset to the beginning")
}
