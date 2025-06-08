#!/bin/bash

# Script to clean up old GitHub Actions runners
# This script removes offline runners from GitHub using the GitHub API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
GITHUB_TOKEN="${GITHUB_TOKEN}"
REPO="Auriora/OneMount"

if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${RED}Error: GITHUB_TOKEN environment variable is required${NC}"
    exit 1
fi

echo -e "${YELLOW}🧹 Cleaning up old GitHub Actions runners...${NC}"

# List of offline runner IDs to remove (based on current status)
OFFLINE_RUNNERS=(
    "39:onemount-simple-runner"
    "40:onemount-simple-runner-2"
)

# Keep these runners (they are online and working)
KEEP_RUNNERS=(
    "41:onemount-runner-1"
    "42:onemount-runner-2"
)

echo -e "${GREEN}✅ Keeping these working runners:${NC}"
for runner in "${KEEP_RUNNERS[@]}"; do
    IFS=':' read -r id name <<< "$runner"
    echo "  - $name (ID: $id) - ONLINE"
done

echo -e "${YELLOW}🗑️  Removing these offline runners:${NC}"
for runner in "${OFFLINE_RUNNERS[@]}"; do
    IFS=':' read -r id name <<< "$runner"
    echo "  - $name (ID: $id) - OFFLINE"
    
    # Remove runner using GitHub API
    response=$(curl -s -w "%{http_code}" \
        -X DELETE \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/$REPO/actions/runners/$id" \
        -o /dev/null)
    
    if [ "$response" = "204" ]; then
        echo -e "    ${GREEN}✅ Removed successfully${NC}"
    else
        echo -e "    ${RED}❌ Failed to remove (HTTP: $response)${NC}"
    fi
    
    # Small delay to avoid rate limiting
    sleep 1
done

echo -e "${GREEN}🎉 Cleanup completed!${NC}"

# Verify final state
echo -e "${YELLOW}📊 Final runner count:${NC}"
final_count=$(curl -s \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/$REPO/actions/runners" | \
    jq -r '.total_count')

echo -e "  Total runners: ${GREEN}$final_count${NC} (should be 2)"

if [ "$final_count" = "2" ]; then
    echo -e "${GREEN}✅ Perfect! Only the 2 working runners remain.${NC}"
else
    echo -e "${YELLOW}⚠️  Expected 2 runners, found $final_count. You may need to manually remove remaining offline runners.${NC}"
fi
