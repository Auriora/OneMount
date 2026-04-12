package graph

import (
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestUT_GR_11_01_SHA256Hash_VariousInputs_ReturnsCorrectHash tests the SHA256Hash function with different inputs.
//
//	Test Case ID    UT-GR-11-01
//	Title           SHA256 Hash Calculation
//	Description     Tests the SHA256Hash function with different inputs
//	Preconditions   None
//	Steps           1. Create byte arrays with different test content
//	                2. Calculate the SHA256 hash of each content
//	                3. Compare the results with expected values
//	Expected Result SHA256Hash returns the correct hash for each input
//	Notes: This test verifies that the SHA256Hash function correctly calculates SHA256 hashes.
func TestUT_GR_11_01_SHA256Hash_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("SHA256HashFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Test case 1: Empty byte array
		emptyData := []byte("")
		emptyHash := SHA256Hash(&emptyData)
		// SHA256 of empty string is E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855
		assert.Equal("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855", emptyHash, "SHA256 of empty string should match known value")

		// Test case 2: Small content ("hello world")
		helloData := []byte("hello world")
		helloHash := SHA256Hash(&helloData)
		// SHA256 of "hello world" is B94D27B9934D3E08A52E52D7DA7DABFAC484EFE37A5380EE9088F7ACE2EFCDE9
		assert.Equal("B94D27B9934D3E08A52E52D7DA7DABFAC484EFE37A5380EE9088F7ACE2EFCDE9", helloHash, "SHA256 of 'hello world' should match known value")

		// Test case 3: Binary data (non-UTF8 content)
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
		binaryHash := SHA256Hash(&binaryData)
		assert.NotEqual("", binaryHash, "SHA256 of binary data should not be empty")
		assert.Equal(64, len(binaryHash), "SHA256 hash should be 64 characters long")

		// Test case 4: Larger content (test with 1KB of data)
		largeData := make([]byte, 1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		largeHash := SHA256Hash(&largeData)
		assert.NotEqual("", largeHash, "SHA256 of large data should not be empty")
		assert.Equal(64, len(largeHash), "SHA256 hash should be 64 characters long")

		// Test case 5: Verify hash is uppercase (as per OneDrive API requirement)
		testData := []byte("test")
		testHash := SHA256Hash(&testData)
		assert.Equal(testHash, strings.ToUpper(testHash), "SHA256 hash should be uppercase")
	})
}

// TestUT_GR_12_01_SHA256HashStream_VariousInputs_ReturnsCorrectHash tests the SHA256HashStream function with different inputs.
//
//	Test Case ID    UT-GR-12-01
//	Title           SHA256 Hash Stream Calculation
//	Description     Tests the SHA256HashStream function with different inputs
//	Preconditions   None
//	Steps           1. Create readers with different test content
//	                2. Calculate the SHA256 hash of each content using SHA256HashStream
//	                3. Compare the results with expected values
//	Expected Result SHA256HashStream returns the correct hash for each input
//	Notes: This test verifies that the SHA256HashStream function correctly calculates SHA256 hashes from readers.
func TestUT_GR_12_01_SHA256HashStream_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("SHA256HashStreamFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Test case 1: Test with strings.Reader for small content
		testContent := "hello world"
		testData := []byte(testContent)
		expectedHash := SHA256Hash(&testData)

		reader := strings.NewReader(testContent)
		actualHash := SHA256HashStream(reader)
		assert.Equal(expectedHash, actualHash, "SHA256HashStream should produce the same hash as SHA256Hash for the same content")

		// Test case 2: Test with empty content
		emptyData := []byte("")
		expectedEmptyHash := SHA256Hash(&emptyData)
		emptyReader := strings.NewReader("")
		actualEmptyHash := SHA256HashStream(emptyReader)
		assert.Equal(expectedEmptyHash, actualEmptyHash, "SHA256HashStream should handle empty content correctly")

		// Test case 3: Test with binary content
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
		expectedBinaryHash := SHA256Hash(&binaryData)
		binaryReader := strings.NewReader(string(binaryData))
		actualBinaryHash := SHA256HashStream(binaryReader)
		assert.Equal(expectedBinaryHash, actualBinaryHash, "SHA256HashStream should handle binary content correctly")

		// Test case 4: Test with larger content
		largeData := make([]byte, 1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		expectedLargeHash := SHA256Hash(&largeData)
		largeReader := strings.NewReader(string(largeData))
		actualLargeHash := SHA256HashStream(largeReader)
		assert.Equal(expectedLargeHash, actualLargeHash, "SHA256HashStream should handle large content correctly")
	})
}

// TestUT_GR_13_01_SHA1Hash_VariousInputs_ReturnsCorrectHash tests the SHA1Hash function with different inputs.
//
//	Test Case ID    UT-GR-13-01
//	Title           SHA1 Hash Calculation
//	Description     Tests the SHA1Hash function with different inputs
//	Preconditions   None
//	Steps           1. Create byte arrays with different test content
//	                2. Calculate the SHA1 hash of each content
//	                3. Compare the results with expected values
//	Expected Result SHA1Hash returns the correct hash for each input
//	Notes: This test verifies that the SHA1Hash function correctly calculates SHA1 hashes.
func TestUT_GR_13_01_SHA1Hash_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("SHA1HashFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test case 1: Empty byte array
		emptyData := []byte("")
		emptyHash := SHA1Hash(&emptyData)
		assert.Equal("DA39A3EE5E6B4B0D3255BFEF95601890AFD80709", emptyHash, "SHA1 of empty string should match known value")

		// Test case 2: Small content ("hello world")
		helloData := []byte("hello world")
		helloHash := SHA1Hash(&helloData)
		assert.Equal("2AAE6C35C94FCFB415DBE95F408B9CE91EE846ED", helloHash, "SHA1 of 'hello world' should match known value")

		// Test case 3: Binary data
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
		binaryHash := SHA1Hash(&binaryData)
		assert.NotEqual("", binaryHash, "SHA1 of binary data should not be empty")
		assert.Equal(40, len(binaryHash), "SHA1 hash should be 40 characters long")

		// Test case 4: Larger content (1KB)
		largeData := make([]byte, 1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		largeHash := SHA1Hash(&largeData)
		assert.NotEqual("", largeHash, "SHA1 of large data should not be empty")
		assert.Equal(40, len(largeHash), "SHA1 hash should be 40 characters long")

		// Test case 5: Verify hash is uppercase (OneDrive API requirement)
		testData := []byte("test")
		testHash := SHA1Hash(&testData)
		assert.Equal(testHash, strings.ToUpper(testHash), "SHA1 hash should be uppercase")
		assert.Equal("A94A8FE5CCB19BA61C4C0873D391E987982FBBD3", testHash, "SHA1 of 'test' should match known value")
	})
}

// TestUT_GR_14_01_SHA1HashStream_VariousInputs_ReturnsCorrectHash tests the SHA1HashStream function with different inputs.
//
//	Test Case ID    UT-GR-14-01
//	Title           SHA1 Hash Stream Calculation
//	Description     Tests the SHA1HashStream function with different inputs
//	Preconditions   None
//	Steps           1. Create readers with different test content
//	                2. Calculate the SHA1 hash of each content using SHA1HashStream
//	                3. Compare the results with expected values
//	Expected Result SHA1HashStream returns the correct hash for each input
//	Notes: This test verifies that the SHA1HashStream function correctly calculates SHA1 hashes from readers.
func TestUT_GR_14_01_SHA1HashStream_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("SHA1HashStreamFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test case 1: Stream should produce same hash as direct function
		testContent := "hello world"
		testData := []byte(testContent)
		expectedHash := SHA1Hash(&testData)

		reader := strings.NewReader(testContent)
		actualHash := SHA1HashStream(reader)
		assert.Equal(expectedHash, actualHash, "SHA1HashStream should produce the same hash as SHA1Hash for the same content")

		// Test case 2: Empty content
		emptyData := []byte("")
		expectedEmptyHash := SHA1Hash(&emptyData)
		emptyReader := strings.NewReader("")
		actualEmptyHash := SHA1HashStream(emptyReader)
		assert.Equal(expectedEmptyHash, actualEmptyHash, "SHA1HashStream should handle empty content correctly")

		// Test case 3: Binary content
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
		expectedBinaryHash := SHA1Hash(&binaryData)
		binaryReader := strings.NewReader(string(binaryData))
		actualBinaryHash := SHA1HashStream(binaryReader)
		assert.Equal(expectedBinaryHash, actualBinaryHash, "SHA1HashStream should handle binary content correctly")

		// Test case 4: Larger content (1KB)
		largeData := make([]byte, 1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		expectedLargeHash := SHA1Hash(&largeData)
		largeReader := strings.NewReader(string(largeData))
		actualLargeHash := SHA1HashStream(largeReader)
		assert.Equal(expectedLargeHash, actualLargeHash, "SHA1HashStream should handle large content correctly")

		// Test case 5: Verify reader is reset after hashing (can be read again)
		reusableReader := strings.NewReader("reusable content")
		hash1 := SHA1HashStream(reusableReader)
		hash2 := SHA1HashStream(reusableReader)
		assert.Equal(hash1, hash2, "SHA1HashStream should reset the reader, producing identical hashes on repeated calls")
	})
}

// TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash tests the QuickXORHash function with different inputs.
//
//	Test Case ID    UT-GR-15-01
//	Title           QuickXOR Hash Calculation
//	Description     Tests the QuickXORHash function with different inputs
//	Preconditions   None
//	Steps           1. Create byte arrays with different test content
//	                2. Calculate the QuickXOR hash of each content
//	                3. Compare the results with expected values
//	Expected Result QuickXORHash returns the correct hash for each input
//	Notes: This test verifies that the QuickXORHash function correctly calculates QuickXOR hashes.
func TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("QuickXORHashFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test case 1: Empty byte array
		emptyData := []byte("")
		emptyHash := QuickXORHash(&emptyData)
		assert.NotEqual("", emptyHash, "QuickXORHash of empty data should not be empty string")
		// Empty data should produce the zero hash (all zeros base64-encoded)
		assert.Equal("AAAAAAAAAAAAAAAAAAAAAAAAAAA=", emptyHash, "QuickXORHash of empty data should be zero hash")

		// Test case 2: Small content — verify non-empty and base64 format
		helloData := []byte("hello world")
		helloHash := QuickXORHash(&helloData)
		assert.NotEqual("", helloHash, "QuickXORHash of 'hello world' should not be empty")
		assert.NotEqual("AAAAAAAAAAAAAAAAAAAAAAAAAAA=", helloHash, "QuickXORHash of non-empty data should differ from zero hash")

		// Test case 3: Deterministic — same input always produces same output
		helloHash2 := QuickXORHash(&helloData)
		assert.Equal(helloHash, helloHash2, "QuickXORHash should be deterministic")

		// Test case 4: Different inputs produce different hashes
		otherData := []byte("hello world!")
		otherHash := QuickXORHash(&otherData)
		assert.NotEqual(helloHash, otherHash, "Different inputs should produce different QuickXOR hashes")

		// Test case 5: Large file (1MB) — verify performance and correctness
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		largeHash := QuickXORHash(&largeData)
		assert.NotEqual("", largeHash, "QuickXORHash of large data should not be empty")
		assert.NotEqual("AAAAAAAAAAAAAAAAAAAAAAAAAAA=", largeHash, "QuickXORHash of large data should not be zero hash")

		// Test case 6: Single byte
		singleByte := []byte{0x42}
		singleHash := QuickXORHash(&singleByte)
		assert.NotEqual("", singleHash, "QuickXORHash of single byte should not be empty")

		// Test case 7: Binary data with null bytes
		binaryData := []byte{0x00, 0x00, 0x00, 0x00}
		binaryHash := QuickXORHash(&binaryData)
		assert.NotEqual("", binaryHash, "QuickXORHash of null bytes should not be empty")
	})
}

// TestUT_GR_16_01_QuickXORHashStream_VariousInputs_ReturnsCorrectHash tests the QuickXORHashStream function with different inputs.
//
//	Test Case ID    UT-GR-16-01
//	Title           QuickXOR Hash Stream Calculation
//	Description     Tests the QuickXORHashStream function with different inputs
//	Preconditions   None
//	Steps           1. Create readers with different test content
//	                2. Calculate the QuickXOR hash of each content using QuickXORHashStream
//	                3. Compare the results with expected values
//	Expected Result QuickXORHashStream returns the correct hash for each input
//	Notes: This test verifies that the QuickXORHashStream function correctly calculates QuickXOR hashes from readers.
func TestUT_GR_16_01_QuickXORHashStream_VariousInputs_ReturnsCorrectHash(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("QuickXORHashStreamFixture")

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test case 1: Stream should produce same hash as direct function
		testContent := "hello world"
		testData := []byte(testContent)
		expectedHash := QuickXORHash(&testData)

		reader := strings.NewReader(testContent)
		actualHash := QuickXORHashStream(reader)
		assert.Equal(expectedHash, actualHash, "QuickXORHashStream should produce the same hash as QuickXORHash")

		// Test case 2: Empty content
		emptyData := []byte("")
		expectedEmptyHash := QuickXORHash(&emptyData)
		emptyReader := strings.NewReader("")
		actualEmptyHash := QuickXORHashStream(emptyReader)
		assert.Equal(expectedEmptyHash, actualEmptyHash, "QuickXORHashStream should handle empty content correctly")

		// Test case 3: Binary content
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
		expectedBinaryHash := QuickXORHash(&binaryData)
		binaryReader := strings.NewReader(string(binaryData))
		actualBinaryHash := QuickXORHashStream(binaryReader)
		assert.Equal(expectedBinaryHash, actualBinaryHash, "QuickXORHashStream should handle binary content correctly")

		// Test case 4: Large content (1MB)
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		expectedLargeHash := QuickXORHash(&largeData)
		largeReader := strings.NewReader(string(largeData))
		actualLargeHash := QuickXORHashStream(largeReader)
		assert.Equal(expectedLargeHash, actualLargeHash, "QuickXORHashStream should handle large content correctly")

		// Test case 5: Reader is reset after hashing
		reusableReader := strings.NewReader("reusable content")
		hash1 := QuickXORHashStream(reusableReader)
		hash2 := QuickXORHashStream(reusableReader)
		assert.Equal(hash1, hash2, "QuickXORHashStream should reset the reader, producing identical hashes on repeated calls")
	})
}
