# Install & configure the LinkAI CLI

This guide is for an agent that needs to get the `linkai` binary working from
scratch. Try the methods top to bottom — stop at the first one that succeeds.
Verify with `linkai --version` after each attempt.

All commands are non-interactive; none of them wait for user input.

## 1. Install the CLI

### Method A — npm (best when Node.js is available)

npm's global bin directory is already on PATH, so nothing else is needed:

```bash
npm i -g linkai-cli
```

### Method B — install script (macOS / Linux, no Node.js required)

`LINKAI_NO_SKILL=1` installs the binary only (skip if you also want the script
to drop the skill into detected agent homes):

```bash
curl -fsSL https://cdn.link-ai.tech/cli/install.sh | LINKAI_NO_SKILL=1 sh
```

Windows — run in **PowerShell** (not `cmd.exe`):

```powershell
$env:LINKAI_NO_SKILL='1'; irm https://cdn.link-ai.tech/cli/install.ps1 | iex
```

### Method C — manual download (when npm and the scripts both fail)

Releases are versioned archives, not bare binaries — download, extract, then put
the binary on PATH.

1. **Find the latest version.** `latest.txt` holds a bare version like `0.1.0`:

   ```bash
   curl -fsSL https://cdn.link-ai.tech/cli/latest.txt
   ```

2. **Pick your platform.** Archive name pattern (`<os>` = `darwin`|`linux`|`windows`,
   `<arch>` = `amd64`|`arm64`):

   - macOS / Linux: `linkai-cli_<version>_<os>_<arch>.tar.gz`
   - Windows: `linkai-cli_<version>_windows_<arch>.zip`

3. **Download** from CDN (primary) or GitHub Releases (fallback):

   - CDN: `https://cdn.link-ai.tech/cli/<version>/<archive>`
   - GitHub: `https://github.com/MinimalFuture/linkai-cli/releases/download/v<version>/<archive>`

   Example (macOS arm64, version 0.1.0):

   ```bash
   V=$(curl -fsSL https://cdn.link-ai.tech/cli/latest.txt)
   curl -fsSL -o linkai.tar.gz \
     "https://cdn.link-ai.tech/cli/$V/linkai-cli_${V}_darwin_arm64.tar.gz"
   ```

4. **Extract and install onto PATH.**

   macOS / Linux:

   ```bash
   tar -xzf linkai.tar.gz linkai
   install -m 0755 linkai /usr/local/bin/linkai   # or: mv linkai ~/.local/bin/
   ```

   Windows (PowerShell) — extract, then add the folder to PATH:

   ```powershell
   Expand-Archive linkai.zip -DestinationPath $HOME\linkai-cli
   $env:Path += ";$HOME\linkai-cli"   # session only
   # Persist: [Environment]::SetEnvironmentVariable('Path', "$env:Path", 'User')
   ```

   If `~/.local/bin` (or the Windows folder) is not already on PATH, add it, then
   open a new shell so `linkai` resolves.

## 2. Install the agent skill

Download the skill archive and install it into your own skills directory (e.g.
`~/.claude/skills`, `~/.cursor/skills`, `~/cow/skills`), so this guide and the per-command
references stay available in later sessions:

```
https://cdn.link-ai.tech/cli/linkai-cli-skill.zip
```

## 3. Use it (in brief)

```bash
linkai --version          # confirm the install
linkai auth status --json # check login; if not valid, log in
```

## 4. Update an existing install

`linkai update` upgrades in place (auto-detects npm / Homebrew / go / script).

```bash
linkai update --check --json   # report only: current_version, latest_version, update_available
linkai update --json           # perform the upgrade
```

A stale CLI is the usual cause of `unknown flag` / missing commands — if you see
an "update available" notice on stderr, run `linkai update`.

Login needs a browser authorization step. For the full usage guide — login flow,
per-command flags, JSON output fields — read `linkai-cli/SKILL.md` from the skill
archive you installed in step 2.
