package helpers

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	// Use crypto/rand for better randomness
	if _, err := rand.Read(b); err != nil {
		// Fallback to simple pattern if random fails
		for i := range b {
			b[i] = charset[i%len(charset)]
		}
		return string(b)
	}

	// Map random bytes to charset
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

// CopyDirectory recursively copies a directory from src to dst
func CopyDirectory(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// GetFileStatus retrieves the file status from extended attributes
// This is used to check sync status of files in the mounted filesystem
func GetFileStatus(filePath string) (string, error) {
	// Try to get the onemount.status extended attribute
	attrName := "user.onemount.status"

	// Get attribute size
	size, err := syscall.Getxattr(filePath, attrName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get attribute size: %w", err)
	}

	// Get attribute value
	buf := make([]byte, size)
	_, err = syscall.Getxattr(filePath, attrName, buf)
	if err != nil {
		return "", fmt.Errorf("failed to get attribute value: %w", err)
	}

	return string(buf), nil
}

// SetFileStatus sets the file status extended attribute
func SetFileStatus(filePath, status string) error {
	attrName := "user.onemount.status"
	return syscall.Setxattr(filePath, attrName, []byte(status), 0)
}

// GetFileETag retrieves the ETag from extended attributes
func GetFileETag(filePath string) (string, error) {
	attrName := "user.onemount.etag"

	size, err := syscall.Getxattr(filePath, attrName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get etag attribute size: %w", err)
	}

	buf := make([]byte, size)
	_, err = syscall.Getxattr(filePath, attrName, buf)
	if err != nil {
		return "", fmt.Errorf("failed to get etag attribute value: %w", err)
	}

	return string(buf), nil
}

// WaitForFileStatus waits for a file to reach a specific status, with timeout
func WaitForFileStatus(filePath, expectedStatus string, timeout, checkInterval time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := GetFileStatus(filePath)
		if err == nil && status == expectedStatus {
			return nil
		}

		time.Sleep(checkInterval)
	}

	return fmt.Errorf("timeout waiting for file status %s", expectedStatus)
}
