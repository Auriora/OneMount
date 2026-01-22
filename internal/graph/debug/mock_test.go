package debug

import (
	"fmt"
	"testing"

	"github.com/auriora/onemount/internal/graph/mock"
)

// TestMockPackage tests that we can access the mock package
func TestUT_Graph_Debug_MockPackage(t *testing.T) {
	fmt.Println("=== MOCK PACKAGE TEST STARTED ===")

	// Create a mock graph provider
	mockProvider := mock.NewMockGraphProvider()

	// Print the mock graph provider
	fmt.Printf("Mock graph provider: %+v\n", mockProvider)

	fmt.Println("=== MOCK PACKAGE TEST COMPLETED ===")
}
