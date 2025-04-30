package graph

import (
	"github.com/rs/zerolog/log"
	"testing"

	"github.com/bcherrington/onemount/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		path          string
		shouldSucceed bool
		expectedName  string
	}{
		{
			name:          "RootPath_ShouldReturnRootItem",
			path:          "/",
			shouldSucceed: true,
			expectedName:  "root",
		},
		{
			name:          "NonexistentPath_ShouldReturnError",
			path:          "/lkjfsdlfjdwjkfl",
			shouldSucceed: false,
			expectedName:  "",
		},
		{
			name:          "DocumentsPath_ShouldReturnDocumentsItem",
			path:          "/Documents",
			shouldSucceed: true,
			expectedName:  "Documents",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			var auth Auth
			err := auth.FromFile(testutil.AuthTokensPath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to load auth tokens")
				return
			}

			item, err := GetItemPath(tc.path, &auth)

			if tc.shouldSucceed {
				require.NoError(t, err, "Failed to get item at path %s", tc.path)
				assert.Equal(t, tc.expectedName, item.Name, "Item name did not match expected value")
			} else {
				assert.Error(t, err, "Expected an error for path %s but got none", tc.path)
			}
		})
	}
}
