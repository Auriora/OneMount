package graph

import (
	"github.com/auriora/onemount/internal/testutil/framework"
	"os"
	"strings"
	"testing"
)

// TestUT_GR_08_01_SHA1Hash_ReaderInput_MatchesDirectCalculation tests the SHA1HashStream function with a reader.
//
//	Test Case ID    UT-GR-08-01
//	Title           SHA1 Hash Stream Calculation
//	Description     Tests the SHA1HashStream function with a reader
//	Preconditions   None
//	Steps           1. Create a byte array with test content
//	                2. Calculate the SHA1 hash of the content
//	                3. Create a reader from the content
//	                4. Calculate the SHA1 hash using SHA1HashStream
//	                5. Compare the two hashes
//	Expected Result The hash calculated from the reader matches the hash calculated from the byte array
//	Notes: This test verifies that the SHA1HashStream function correctly calculates SHA1 hashes from readers.
func TestUT_GR_08_01_SHA1Hash_ReaderInput_MatchesDirectCalculation(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("SHA1HashReaderFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create test content
		testContent := []byte("Hello, World! This is a test for SHA1 hash calculation.")

		// Calculate the SHA1 hash of the content using the direct method
		expectedHash := SHA1Hash(&testContent)

		// Create a reader from the content
		reader := strings.NewReader(string(testContent))

		// Calculate the SHA1 hash using SHA1HashStream
		actualHash := SHA1HashStream(reader)

		// Compare the two hashes
		assert := framework.NewAssert(t)
		assert.Equal(expectedHash, actualHash, "SHA1HashStream should produce the same hash as SHA1Hash for the same content")

		// Test with empty content
		emptyContent := []byte("")
		expectedEmptyHash := SHA1Hash(&emptyContent)
		emptyReader := strings.NewReader("")
		actualEmptyHash := SHA1HashStream(emptyReader)
		assert.Equal(expectedEmptyHash, actualEmptyHash, "SHA1HashStream should handle empty content correctly")

		// Test with larger content
		largeContent := make([]byte, 1024)
		for i := range largeContent {
			largeContent[i] = byte(i % 256)
		}
		expectedLargeHash := SHA1Hash(&largeContent)
		largeReader := strings.NewReader(string(largeContent))
		actualLargeHash := SHA1HashStream(largeReader)
		assert.Equal(expectedLargeHash, actualLargeHash, "SHA1HashStream should handle large content correctly")
	})
}

// TestUT_GR_09_01_QuickXORHash_ReaderInput_MatchesDirectCalculation tests the QuickXORHashStream function with a reader.
//
//	Test Case ID    UT-GR-09-01
//	Title           QuickXOR Hash Stream Calculation
//	Description     Tests the QuickXORHashStream function with a reader
//	Preconditions   None
//	Steps           1. Create a byte array with test content
//	                2. Calculate the QuickXOR hash of the content
//	                3. Create a reader from the content
//	                4. Calculate the QuickXOR hash using QuickXORHashStream
//	                5. Compare the two hashes
//	Expected Result The hash calculated from the reader matches the hash calculated from the byte array
//	Notes: This test verifies that the QuickXORHashStream function correctly calculates QuickXOR hashes from readers.
func TestUT_GR_09_01_QuickXORHash_ReaderInput_MatchesDirectCalculation(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("QuickXORHashReaderFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create test content
		testContent := []byte("Hello, World! This is a test for QuickXOR hash calculation.")

		// Calculate the QuickXOR hash of the content using the direct method
		expectedHash := QuickXORHash(&testContent)

		// Create a reader from the content
		reader := strings.NewReader(string(testContent))

		// Calculate the QuickXOR hash using QuickXORHashStream
		actualHash := QuickXORHashStream(reader)

		// Compare the two hashes
		assert := framework.NewAssert(t)
		assert.Equal(expectedHash, actualHash, "QuickXORHashStream should produce the same hash as QuickXORHash for the same content")

		// Test with empty content
		emptyContent := []byte("")
		expectedEmptyHash := QuickXORHash(&emptyContent)
		emptyReader := strings.NewReader("")
		actualEmptyHash := QuickXORHashStream(emptyReader)
		assert.Equal(expectedEmptyHash, actualEmptyHash, "QuickXORHashStream should handle empty content correctly")

		// Test with larger content
		largeContent := make([]byte, 1024)
		for i := range largeContent {
			largeContent[i] = byte(i % 256)
		}
		expectedLargeHash := QuickXORHash(&largeContent)
		largeReader := strings.NewReader(string(largeContent))
		actualLargeHash := QuickXORHashStream(largeReader)
		assert.Equal(expectedLargeHash, actualLargeHash, "QuickXORHashStream should handle large content correctly")
	})
}

// TestUT_GR_10_01_HashFunctions_AfterReading_ResetSeekPosition tests that hash functions reset the seek position.
//
//	Test Case ID    UT-GR-10-01
//	Title           Hash Functions Seek Position Reset
//	Description     Tests that hash functions reset the seek position
//	Preconditions   None
//	Steps           1. Create a temporary file with test content
//	                2. Read a portion of the file to move the seek position
//	                3. Calculate hashes using the hash stream functions
//	                4. Verify that the seek position is reset to the beginning after each hash calculation
//	Expected Result The seek position is reset to the beginning after each hash calculation
//	Notes: This test verifies that the hash stream functions correctly reset the seek position after calculating hashes.
func TestUT_GR_10_01_HashFunctions_AfterReading_ResetSeekPosition(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("HashSeekPositionFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary file for the test
		tempFile, err := os.CreateTemp("", "hash-test-*.txt")
		if err != nil {
			return nil, err
		}

		// Write test content to the file
		content := []byte("test content for hash functions")
		if _, err := tempFile.Write(content); err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return nil, err
		}

		// Seek to the beginning of the file
		if _, err := tempFile.Seek(0, 0); err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return nil, err
		}

		return map[string]interface{}{
			"tempFile": tempFile,
			"content":  content,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary file
		data := fixture.(map[string]interface{})
		tempFile := data["tempFile"].(*os.File)
		fileName := tempFile.Name()
		tempFile.Close()
		return os.Remove(fileName)
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Create test content
		testContent := "Hello, World! This is a test for hash functions seek position reset."
		reader := strings.NewReader(testContent)

		// Read a portion of the content to move the seek position
		buffer := make([]byte, 10)
		n, err := reader.Read(buffer)
		assert.NoError(err, "Should be able to read from reader")
		assert.Equal(10, n, "Should read 10 bytes")

		// Verify the seek position is not at the beginning
		currentPos, err := reader.Seek(0, 1) // Get current position
		assert.NoError(err, "Should be able to get current position")
		assert.Equal(int64(10), currentPos, "Current position should be 10")

		// Calculate SHA1 hash - this should reset the seek position
		sha1Hash := SHA1HashStream(reader)
		assert.NotEqual("", sha1Hash, "SHA1 hash should not be empty")

		// Verify the seek position is reset to the beginning
		currentPos, err = reader.Seek(0, 1) // Get current position
		assert.NoError(err, "Should be able to get current position")
		assert.Equal(int64(0), currentPos, "Seek position should be reset to 0 after SHA1HashStream")

		// Move seek position again
		reader.Seek(15, 0)
		currentPos, err = reader.Seek(0, 1)
		assert.NoError(err, "Should be able to get current position")
		assert.Equal(int64(15), currentPos, "Current position should be 15")

		// Calculate QuickXOR hash - this should also reset the seek position
		quickXorHash := QuickXORHashStream(reader)
		assert.NotEqual("", quickXorHash, "QuickXOR hash should not be empty")

		// Verify the seek position is reset to the beginning
		currentPos, err = reader.Seek(0, 1) // Get current position
		assert.NoError(err, "Should be able to get current position")
		assert.Equal(int64(0), currentPos, "Seek position should be reset to 0 after QuickXORHashStream")
	})
}
