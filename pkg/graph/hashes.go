package graph

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/auriora/onemount/pkg/quickxorhash"
	"io"
	"strings"
)

func SHA256Hash(data *[]byte) string {
	return strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256(*data)))
}

func SHA256HashStream(reader io.ReadSeeker) string {
	reader.Seek(0, 0)
	hash := sha256.New()
	io.Copy(hash, reader)
	reader.Seek(0, 0)
	return strings.ToUpper(fmt.Sprintf("%x", hash.Sum(nil)))
}

// SHA1Hash returns the SHA1 hash of some data as a string
func SHA1Hash(data *[]byte) string {
	// the onedrive API returns SHA1 hashes in all caps, so we do too
	return strings.ToUpper(fmt.Sprintf("%x", sha1.Sum(*data)))
}

// SHA1HashStream hashes the contents of a stream.
func SHA1HashStream(reader io.ReadSeeker) string {
	reader.Seek(0, 0)
	hash := sha1.New()
	io.Copy(hash, reader)
	reader.Seek(0, 0)
	return strings.ToUpper(fmt.Sprintf("%x", hash.Sum(nil)))
}

// QuickXORHash computes the Microsoft-specific QuickXORHash. Reusing rclone's
// implementation until I get the chance to rewrite/add test cases to remove the
// dependency.
func QuickXORHash(data *[]byte) string {
	hash := quickxorhash.Sum(*data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// QuickXORHashStream hashes a stream.
func QuickXORHashStream(reader io.ReadSeeker) string {
	reader.Seek(0, 0)
	hash := quickxorhash.New()
	io.Copy(hash, reader)
	reader.Seek(0, 0)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// Note: VerifyChecksum and ETagIsMatch methods have been moved to api.DriveItem
// They are available through the type alias DriveItem = api.DriveItem
