# claude-cli-git-helper

A CLI tool to streamline git worktree workflows with Claude Code integration.

Creates a new git worktree, optionally symlinks dependencies, and opens a terminal with Claude ready to go.

## Installation

### macOS

```bash
# Download (Apple Silicon)
curl -L https://github.com/CEske/claude-cli-git-helper/releases/latest/download/claude-cli-git-helper_darwin_arm64.zip -o claude-cli-git-helper.zip

# Download (Intel)
curl -L https://github.com/CEske/claude-cli-git-helper/releases/latest/download/claude-cli-git-helper_darwin_amd64.zip -o claude-cli-git-helper.zip

# Unzip and install
unzip claude-cli-git-helper.zip
chmod +x claude-cli-git-helper
sudo mv claude-cli-git-helper /usr/local/bin/
```

### Windows

1. Download the latest release from [Releases](https://github.com/CEske/claude-cli-git-helper/releases)
2. Extract `claude-cli-git-helper_windows_amd64.zip`
3. Move `claude-cli-git-helper.exe` to a folder (e.g., `C:\tools`)
4. Add the folder to your PATH:
   - Press `Win + R` → type `sysdm.cpl` → Enter
   - Go to **Advanced** → **Environment Variables**
   - Under **User variables**, select **Path** → **Edit** → **New**
   - Add `C:\tools`
   - Click **OK** and restart your terminal

## Usage

```bash
# Create worktree from existing branch
claude-cli-git-helper worktree feature-x feature-x

# Create worktree with new branch
claude-cli-git-helper worktree feature-y my-new-branch -b

# Create new branch from specific origin
claude-cli-git-helper worktree feature-z my-new-branch -b -o main

# Custom worktree directory
claude-cli-git-helper worktree feature-x some-branch -d ~/projects/worktrees

# Fresh dependency install instead of symlinks
claude-cli-git-helper worktree feature-x some-branch --symlink-deps=false
```

## Flags

| Flag             | Short | Default | Description                                   |
| ---------------- | ----- | ------- | --------------------------------------------- |
| `--dir`          | `-d`  | `../`   | Base directory for worktrees                  |
| `--new-branch`   | `-b`  | `false` | Create a new branch                           |
| `--origin`       | `-o`  | `""`    | Origin branch (for new branches)              |
| `--symlink-deps` | `-s`  | `true`  | Symlink dependencies instead of fresh install |

## Features

- **Git validation** — Verifies you're in a git repository and branches exist
- **Dependency symlinking** — Symlinks `node_modules` and `vendor` from source repo
- **Fresh installs** — Optionally run `npm install` / `composer install` instead
- **Claude integration** — Opens a new terminal with Claude Code ready to use

## Supported Dependency Managers

| Directory      | Lock Files                                                         | Install Command    |
| -------------- | ------------------------------------------------------------------ | ------------------ |
| `node_modules` | `package.json`, `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml` | `npm install`      |
| `vendor`       | `composer.json`, `composer.lock`                                   | `composer install` |

## Requirements

- Git
- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code)
- **Windows**: Windows Terminal (`wt`)
- **macOS**: Default Terminal app

## License

MIT
