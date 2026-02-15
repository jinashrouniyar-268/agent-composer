# Agent Composer CLI

Create research agents directly from your terminal using the [Contextual AI](https://contextual.ai) platform.

## Installation

### Homebrew (Recommended)

```bash
brew install YOUR_ORG/agent-composer/agent
```

Or tap first:

```bash
brew tap YOUR_ORG/agent-composer
brew install agent
```

### Manual Download

Download the appropriate archive from the [Releases](https://github.com/YOUR_ORG/agent-composer/releases) page.

#### macOS Gatekeeper Warning

When downloading directly from GitHub releases on macOS, you may see:

> "Apple could not verify agent is free of malware that may harm your Mac or compromise your privacy."

This happens because the binary is not signed with an Apple Developer certificate. To bypass this:

```bash
# After extracting the archive, remove the quarantine attribute:
xattr -d com.apple.quarantine ./agent
```

Or right-click the binary, select "Open", and click "Open" in the dialog.

**Note:** Installing via Homebrew avoids this issue automatically.

## Commands

- **`agent init`** — Initialize a new agent project in the current directory.
- **`agent add tool`** — Add a tool to your research agent.
- **`agent run`** — Run your agent (connects to the Contextual AI platform).

## Releasing (for maintainers)

Releases are automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions.

### Creating a Release

1. **Ensure all changes are committed and pushed to `main`**

2. **Create and push a new tag:**
   ```bash
   git tag -l
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. **GitHub Actions will automatically:**
   - Build binaries for all platforms (macOS, Linux, Windows)
   - Create a GitHub release with the binaries
   - Generate checksums and changelog
   - Update the Homebrew formula in [homebrew-agent-composer](https://github.com/YOUR_ORG/homebrew-agent-composer) (if `HOMEBREW_TOKEN` is set)

4. **To update the Homebrew formula manually** (if not using CI):
   ```bash
   cd homebrew-agent-composer
   ./scripts/update-formula.sh 0.1.0
   git add Formula/agent.rb
   git commit -m "Update agent to v0.1.0"
   git push
   ```

### Version Naming

- Production releases: `v1.0.0`, `v1.1.0`, `v2.0.0`
- Pre-releases: `v0.1.0-alpha`, `v0.1.0-beta`, `v0.1.0-rc1`

Pre-release tags are automatically marked as pre-releases on GitHub.
