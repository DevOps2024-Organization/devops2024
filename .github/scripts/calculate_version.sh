#!/bin/bash

# Fetch tags and sort them to get the latest tag
git fetch --tags
latest_tag=$(git tag | sort -V | tail -n 1)
echo "Latest tag: $latest_tag"

# If no tags exist, start with v0.0.0
if [ -z "$latest_tag" ]; then
	latest_tag="v0.0.0"
fi

# Break the tag into components
major=$(echo $latest_tag | cut -d. -f1)
minor=$(echo $latest_tag | cut -d. -f2)
patch=$(echo $latest_tag | cut -d. -f3)

# Increment the patch version
new_patch=$((patch + 1))

# Construct the new tag
new_tag="${major}.${minor}.${new_patch}"
echo "New tag: $new_tag"

# Set the new tag for later steps in the GitHub Actions workflow
echo "NEW_VERSION=$new_tag" >>$GITHUB_ENV
