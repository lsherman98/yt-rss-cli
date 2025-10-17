# Releasing yt-rss-cli

This guide explains how to release a new version of `ytrss` using GoReleaser.

## Prerequisites

1. **Install GoReleaser** (for local testing):

    ```bash
    brew install goreleaser/tap/goreleaser
    # or
    go install github.com/goreleaser/goreleaser@latest
    ```

2. **GitHub Token**:
    - The GitHub Actions workflow uses the automatic `GITHUB_TOKEN`
    - For Homebrew tap publishing, you'll need a personal access token with `repo` scope

## Release Process

### 1. Prepare the Release

1. Make sure all changes are committed and pushed to `main`
2. Update CHANGELOG.md (optional but recommended)
3. Ensure tests pass and the app works locally

### 2. Create and Push a Tag

The release is triggered by pushing a git tag starting with `v`:

```bash
# Create a tag (e.g., v1.0.0, v1.1.0, v2.0.0-beta.1)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 3. Automated Release

Once you push the tag:

1. GitHub Actions will automatically trigger (`.github/workflows/release.yml`)
2. GoReleaser will:
    - Build binaries for multiple platforms (macOS, Linux, Windows)
    - Create archives (tar.gz for Unix, zip for Windows)
    - Generate checksums
    - Create a GitHub Release
    - Upload all artifacts
    - Generate release notes from commits

### 4. Verify the Release

1. Go to https://github.com/lsherman98/yt-rss-cli/releases
2. Check that the release was created with all binaries
3. Download and test a binary for your platform

## Testing Locally (Before Tagging)

You can test the release process locally without creating a tag:

```bash
# Dry run - doesn't create a release
goreleaser release --snapshot --clean

# Check the generated binaries in ./dist/
ls -la dist/
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

-   **MAJOR** version (v2.0.0): Breaking changes
-   **MINOR** version (v1.1.0): New features, backwards compatible
-   **PATCH** version (v1.0.1): Bug fixes, backwards compatible
-   **Pre-release** (v1.0.0-beta.1): Testing versions

## Troubleshooting

### Release Failed

Check the GitHub Actions logs:

1. Go to https://github.com/lsherman98/yt-rss-cli/actions
2. Click on the failed workflow
3. Review the error messages

Common issues:

-   **Missing GITHUB_TOKEN**: Should be automatic in Actions
-   **Invalid tag format**: Must start with `v`
-   **Build errors**: Test locally first with `goreleaser release --snapshot`

### Delete a Bad Release

```bash
# Delete the remote tag
git push --delete origin v1.0.0

# Delete the local tag
git tag -d v1.0.0

# Delete the GitHub Release manually from the web interface
```

Then create a new tag and release.