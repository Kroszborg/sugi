package main

import (
	"fmt"
	"os"

	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sugi [path]",
	Short: "sugi - Terminal UI git client",
	Long: `sugi (杉) - A terminal UI git client that beats lazygit.

Features: GitHub/GitLab PR integration, AI commit messages,
commit graph, difftastic support, and more.`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          run,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sugi %s\n", version)
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	startPath := "."
	if len(args) > 0 {
		startPath = args[0]
	}

	repo, err := git.NewClient(startPath)
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	model := ui.New(repo)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
