#!/bin/bash

# Release script for UDDIN-LANG
# Usage: ./scripts/release.sh [VERSION]
# Example: ./scripts/release.sh v1.0.0

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if version is provided
if [ -z "$1" ]; then
    echo -e "${RED}Error: Version is required${NC}"
    echo "Usage: $0 <VERSION>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format (should start with 'v' and be semver)
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    echo -e "${RED}Error: Invalid version format${NC}"
    echo "Version should be in format: v1.0.0 (semver)"
    exit 1
fi

echo -e "${GREEN}🚀 Creating release ${VERSION}${NC}"
echo ""

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${YELLOW}Warning: Working directory has uncommitted changes${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag ${VERSION} already exists${NC}"
    exit 1
fi

# Run tests
echo -e "${GREEN}Running tests...${NC}"
if ! go test ./...; then
    echo -e "${RED}Tests failed. Aborting release.${NC}"
    exit 1
fi

# Build binary to verify it works
echo -e "${GREEN}Building binary...${NC}"
make build VERSION=$VERSION

# Verify version
BUILT_VERSION=$(./uddinlang --version)
echo -e "${GREEN}Built version: ${BUILT_VERSION}${NC}"

# Create git tag
echo -e "${GREEN}Creating git tag ${VERSION}...${NC}"
git tag -a "$VERSION" -m "Release $VERSION"

echo ""
echo -e "${GREEN}✅ Release ${VERSION} created successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review the tag: git show $VERSION"
echo "  2. Push the tag: git push origin $VERSION"
echo "  3. (Optional) Push commits: git push origin main"
echo ""
echo "To publish to GitHub:"
echo "  git push origin $VERSION"
echo ""
echo "To install as library:"
echo "  go get github.com/bonkzero404/uddin-lang@$VERSION"

