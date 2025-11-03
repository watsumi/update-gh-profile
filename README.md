[日本語版 (Japanese) / README.ja.md](README.ja.md)

# Update GH Profile

GitHub Actions workflow that automatically analyzes repository metrics for GitHub users and updates README.md

## Overview

This tool collects and visualizes the following information from all repositories of a GitHub user:

- Language usage ranking
- Commit history graph
- Commit time distribution analysis
- Top 5 languages by commit
- Summary card (stars, repositories, commits, PRs)

## Setup

### Requirements

- Go 1.21 or higher
- GitHub Personal Access Token (environment variable `GITHUB_TOKEN`)

### Local Execution

```bash
# Install dependencies
go mod download

# Run
export GITHUB_TOKEN=your_token_here
go run cmd/update-gh-profile/main.go

# Or, build and run
go build -o update-gh-profile cmd/update-gh-profile/main.go
./update-gh-profile
```

### Usage with GitHub Actions

This repository can be used in GitHub Actions workflows of other repositories.

For details, see:

- [README_ACTION.md](README_ACTION.md) - Basic usage instructions
- [USAGE_ACTION.md](USAGE_ACTION.md) - Detailed usage examples
- [PRIVATE_REPO_SETUP.md](PRIVATE_REPO_SETUP.md) - Setup instructions for private repositories

#### Quick Start

```yaml
- uses: watsumi/update-gh-profile@main
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    exclude_forks: "true"
```

#### Workflow Configuration Example

Complete workflow file example for use in a profile repository (`username/username` format):

```yaml
name: Update GitHub Profile

on:
  schedule:
    # Run daily at 00:00 UTC (09:00 JST)
    - cron: "0 0 * * *"
  workflow_dispatch: # Manual execution is also possible

permissions:
  contents: write # Required to update README.md

jobs:
  update-profile:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Update GitHub Profile
        uses: watsumi/update-gh-profile@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          exclude_forks: "true"
          exclude_languages: "HTML,CSS,JSON" # Languages to exclude (comma-separated)
        # Note: This action automatically commits and pushes internally
        # Authentication is automatically configured via permissions: contents: write,
        # so github_token_write is not needed
```

**Notes:**

- **About `permissions: contents: write`**: This permission setting grants write access to `secrets.GITHUB_TOKEN` and automatically authenticates `git push` operations on the repository checked out by `actions/checkout@v4`. Therefore, the `github_token_write` parameter is not needed (removed).
- **Auto commit and push**: This action automatically commits and pushes after updating README.md. No additional steps are required.
- **Token configuration**: Pass `secrets.GITHUB_TOKEN` to `github_token`. For reading private repositories, set a Personal Access Token with broader permissions to `github_token`.
- **Automatic user detection**: This tool only fetches repositories owned by the authenticated user. The authenticated user is automatically detected, so username specification is not required.
- **Fork exclusion**: Setting `exclude_forks: "true"` excludes forked repositories from statistics.
- **Language exclusion**: You can specify languages to exclude from rankings using the `exclude_languages` parameter. Multiple languages can be specified as comma-separated values (e.g., `"HTML,CSS,JSON"`). Case-insensitive matching is used. Excluded languages are removed from both "Language Ranking" and "Top 5 Languages by Commit" graphs.
