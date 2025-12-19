package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

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

		fmt.Printf("Creating worktree at %s...\n", worktreePath)
		gitCmd := exec.Command("git", gitArgs...)
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr

		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}

		fmt.Println("Opening terminal with Claude...")
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

func openTerminalWithClaude(dir string) error {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("wt", "-d", absPath, "cmd", "/k", "claude").Start()

	case "darwin":
		script := fmt.Sprintf(`
			tell application "Terminal"
				do script "cd '%s' && claude"
				activate
			end tell`, absPath)
		return exec.Command("osascript", "-e", script).Start()

	case "linux":
		terminals := []struct {
			name string
			args []string
		}{
			{"gnome-terminal", []string{"--working-directory", absPath, "--", "claude"}},
			{"konsole", []string{"--workdir", absPath, "-e", "claude"}},
			{"xfce4-terminal", []string{"--working-directory", absPath, "-e", "claude"}},
			{"xterm", []string{"-e", fmt.Sprintf("cd '%s' && claude", absPath)}},
		}

		for _, t := range terminals {
			if _, err := exec.LookPath(t.name); err == nil {
				return exec.Command(t.name, t.args...).Start()
			}
		}
		return fmt.Errorf("no supported terminal found")
	}

	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func init() {
	worktreeCmd.Flags().StringP("dir", "d", "../", "Base directory for worktrees")
	worktreeCmd.Flags().BoolP("new-branch", "b", false, "Create a new branch")
	worktreeCmd.Flags().StringP("origin", "o", "", "Origin branch (for new branches)")
	rootCmd.AddCommand(worktreeCmd)
}
