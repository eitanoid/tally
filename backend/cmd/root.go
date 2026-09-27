package cmd

import (
	"fmt"
	"os"

	"github.com/eitanoid/tally/internal/repository"
	"github.com/eitanoid/tally/internal/service"
	"github.com/spf13/cobra"
)

var (
	dbPath       string
	repo         *repository.SqliteClient
	tallyService *service.TallyService
)

var rootCmd = &cobra.Command{
	Use:   "tally",
	Short: "Modular, versioned tally tracking CLI",
	Long:  `A local-first, offline tally tracker backed by SQLite and versioned JSON schemas.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		var err error
		repo, err = repository.NewSQLiteClient(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open sqlite database at %s: %s", dbPath, err.Error())
		}
		tallyService = service.NewTallyService(repo)
		return nil
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		if repo != nil {
			_ = repo.Close()
		}
	},
}

// Execute executes the cobra root cmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "habits.db", "Path to SQLite database file")
}
