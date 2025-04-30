// Package testutil provides utility functions and constants for testing.
package testutil

import (
	"crypto/rand"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

// DmelfaSize is the size of the dmel.fa file in bytes (approximately 20MB)
const DmelfaSize = 20 * 1024 * 1024

// GenerateDmelfa generates a dmel.fa file with random data if it doesn't exist.
// The file is created at the location specified by DmelfaDir.
// The file is formatted as a FASTA file with a header and random DNA sequence data.
func GenerateDmelfa() error {
	// Check if the file already exists
	if _, err := os.Stat(DmelfaDir); err == nil {
		// File exists, no need to generate it
		return nil
	}

	// Ensure the directory exists
	dir := filepath.Dir(DmelfaDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for dmel.fa: %w", err)
	}

	// Create the file
	file, err := os.Create(DmelfaDir)
	if err != nil {
		return fmt.Errorf("failed to create dmel.fa file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to close dmel.fa file")
		}
	}(file)

	// Write the FASTA header
	header := ">X dna:chromosome chromosome:BDGP6.22:X:1:23542271:1 REF\n"
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header to dmel.fa: %w", err)
	}

	// Generate random DNA sequence data
	// FASTA format typically has 60 characters per line
	const lineLength = 60
	const bases = "ACGT"

	// Calculate how many bytes we need to generate
	// We subtract the header length and the footer length from the total size
	dataSize := DmelfaSize - len(header) - len("AAATAAAATAC\n")

	// Generate the random DNA sequence data in chunks to avoid memory issues
	const chunkSize = 1024 * 1024 // 1MB chunks
	remaining := dataSize

	for remaining > 0 {
		size := remaining
		if size > chunkSize {
			size = chunkSize
		}

		// Create a buffer for this chunk
		buffer := make([]byte, size)

		// Fill the buffer with random data
		if _, err := rand.Read(buffer); err != nil {
			return fmt.Errorf("failed to generate random data: %w", err)
		}

		// Convert the random data to DNA bases (A, C, G, T)
		for i := range buffer {
			// Use the random byte to select a base (mod 4)
			buffer[i] = bases[buffer[i]%4]

			// Add a newline every lineLength characters
			if i > 0 && i%lineLength == lineLength-1 {
				buffer[i] = '\n'
			}
		}

		// Write the chunk to the file
		if _, err := file.Write(buffer); err != nil {
			return fmt.Errorf("failed to write data to dmel.fa: %w", err)
		}

		remaining -= size
	}

	// Write the footer
	footer := "AAATAAAATAC\n"
	if _, err := file.WriteString(footer); err != nil {
		return fmt.Errorf("failed to write footer to dmel.fa: %w", err)
	}

	return nil
}

// EnsureDmelfaExists ensures that the dmel.fa file exists before tests run.
// It calls GenerateDmelfa to create the file if it doesn't exist.
func EnsureDmelfaExists() {
	if err := GenerateDmelfa(); err != nil {
		fmt.Printf("Warning: Failed to generate dmel.fa file: %v\n", err)
	}
}
