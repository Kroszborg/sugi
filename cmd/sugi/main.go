package main

import (
	"fmt"
	"os"

	"github.com/Kroszborg/sugi/internal/ai"
	"github.com/Kroszborg/sugi/internal/config"
	"github.com/Kroszborg/sugi/internal/forge"
	"github.com/Kroszborg/sugi/internal/git"
	"github.com/Kroszborg/sugi/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "sugi [path]",
	Short:   "sugi - Terminal UI git client",
	Version: version,
	Long: `sugi (杉) - A terminal UI git client.

Panels: Files, Branches, Commits, Diff, PRs, Stash, Blame,
Worktrees, Remotes, Bisect, Interactive Rebase, Conflict Resolver.`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE:         run,
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

	cfg := config.Load()

	// Detect forge from origin remote
	var forgeClient forge.ForgeClient
	if originURL := repo.OriginURL(); originURL != "" {
		info := forge.Detect(originURL)
		if info.IsKnown() {
			switch info.Type {
			case forge.ForgeGitHub:
				forgeClient = forge.NewGitHubClient(info, cfg.EffectiveGitHubToken())
			case forge.ForgeGitLab:
				forgeClient = forge.NewGitLabClient(info, cfg.EffectiveGitLabToken())
			}
		}
	}

	aiGen := ai.NewGenerator(cfg.GroqAPIKey, cfg.GroqModel)

	model := ui.New(repo, cfg, forgeClient, aiGen)

	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithOutput(os.Stderr), // use stderr so stdout stays clean in all terminals
	}
	if cfg.MouseEnabled {
		opts = append(opts, tea.WithMouseCellMotion())
	}

	p := tea.NewProgram(model, opts...)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
