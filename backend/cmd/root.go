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
	TallyService *service.TallyService
)

var rootCmd = &cobra.Command{
	Use:   "tally",
	Short: "Modular, versioned tally tracking CLI",
	Long:  `A local-first, offline tally tracker backed by SQLite and versioned JSON schemas.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		repo, err = repository.NewSQLiteClient(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open sqlite database at %s: %s", dbPath, err.Error())
		}
		TallyService = service.NewTallyService(repo)
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if repo != nil {
			_ = repo.Close()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "habits.db", "Path to SQLite database file")
}
