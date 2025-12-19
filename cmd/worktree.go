package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type DependencyConfig struct {
	Dir       string
	LockFiles []string
	Install   string
}

var dependencies = []DependencyConfig{
	{
		Dir:       "node_modules",
		LockFiles: []string{"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml"},
		Install:   "npm ci",
	},
	{
		Dir:       "vendor",
		LockFiles: []string{"composer.json", "composer.lock"},
		Install:   "composer install",
	},
}

var worktreeCmd = &cobra.Command{
	Use:   "worktree [name] [branch]",
	Short: "Create a git worktree and open Claude in a new terminal",
	Args:  cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := validateGitRepo(); err != nil {
			return err
		}

		branch := args[1]
		newBranch, _ := cmd.Flags().GetBool("new-branch")
		originBranch, _ := cmd.Flags().GetString("origin")

		if !newBranch {
			if err := validateBranchExists(branch); err != nil {
				return err
			}
		}

		if originBranch != "" {
			if err := validateBranchExists(originBranch); err != nil {
				return fmt.Errorf("origin branch invalid: %w", err)
			}
		}

		baseDir, _ := cmd.Flags().GetString("dir")
		worktreePath := filepath.Join(baseDir, args[0])
		if _, err := os.Stat(worktreePath); err == nil {
			return fmt.Errorf("directory already exists: %s", worktreePath)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		branch := args[1]

		baseDir, _ := cmd.Flags().GetString("dir")
		newBranch, _ := cmd.Flags().GetBool("new-branch")
		originBranch, _ := cmd.Flags().GetString("origin")
		symlinkDeps, _ := cmd.Flags().GetBool("symlink-deps")

		worktreePath := filepath.Join(baseDir, name)

		gitArgs := []string{"worktree", "add"}

		if newBranch {
			gitArgs = append(gitArgs, "-b", branch, worktreePath)
			if originBranch != "" {
				gitArgs = append(gitArgs, originBranch)
			}
		} else {
			gitArgs = append(gitArgs, worktreePath, branch)
		}

		gitCmd := exec.Command("git", gitArgs...)
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr

		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}

		if symlinkDeps {
			sourceRepo, err := getRepoRoot()
			if err != nil {
				return err
			}
			if err := symlinkDependencies(sourceRepo, worktreePath); err != nil {
				return err
			}
		} else {
			if err := installFreshDependencies(worktreePath); err != nil {
				return err
			}
		}

		return openTerminalWithClaude(worktreePath)
	},
}

func validateGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}
	return nil
}

func validateBranchExists(branch string) error {
	localCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if localCmd.Run() == nil {
		return nil
	}

	remoteCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
	if remoteCmd.Run() == nil {
		return nil
	}

	return fmt.Errorf("branch '%s' not found locally or in origin", branch)
}

func getRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(string(out))), nil
}

func openTerminalWithClaude(dir string) error {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		return exec.Command("wt", "-d", absPath, "cmd", "/k", "claude").Start()
	}

	script := fmt.Sprintf(`tell application "Terminal"
		do script "cd '%s' && claude"
		activate
	end tell`, absPath)
	return exec.Command("osascript", "-e", script).Start()
}

func installFreshDependencies(worktreePath string) error {
	for _, dep := range dependencies {
		hasLockFile := false
		for _, lockFile := range dep.LockFiles {
			if _, err := os.Stat(filepath.Join(worktreePath, lockFile)); err == nil {
				hasLockFile = true
				break
			}
		}

		if !hasLockFile {
			continue
		}

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", dep.Install)
		} else {
			cmd = exec.Command("sh", "-c", dep.Install)
		}

		cmd.Dir = worktreePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %s: %w", dep.Install, err)
		}
	}

	return nil
}

func symlinkDependencies(sourceRepo, worktreePath string) error {
	for _, dep := range dependencies {
		sourceDep := filepath.Join(sourceRepo, dep.Dir)
		targetDep := filepath.Join(worktreePath, dep.Dir)

		if _, err := os.Stat(sourceDep); os.IsNotExist(err) {
			continue
		}

		if err := createSymlink(sourceDep, targetDep); err != nil {
			return fmt.Errorf("failed to symlink %s: %w", dep.Dir, err)
		}
	}

	return nil
}

func createSymlink(source, target string) error {
	os.RemoveAll(target)

	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", "mklink", "/J", target, source).Run()
	}

	return os.Symlink(source, target)
}

func init() {
	worktreeCmd.Flags().StringP("dir", "d", "../", "Base directory for worktrees")
	worktreeCmd.Flags().BoolP("new-branch", "b", false, "Create a new branch")
	worktreeCmd.Flags().StringP("origin", "o", "", "Origin branch (for new branches)")
	worktreeCmd.Flags().BoolP("symlink-deps", "s", true, "Symlink dependencies instead of fresh install")
	rootCmd.AddCommand(worktreeCmd)
}
