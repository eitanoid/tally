package cmd

import (
	"fmt"
	"os"

	"github.com/eitanoid/habit-tracker/internal/repository"
	"github.com/eitanoid/habit-tracker/internal/service"
	"github.com/spf13/cobra"
)

var (
	dbPath       string
	repo         *repository.SqliteClient
	HabitService *service.HabitService
)

var rootCmd = &cobra.Command{
	Use:   "habit",
	Short: "Modular, versioned habit tracking CLI",
	Long:  `A local-first, offline habit tracker backed by SQLite and versioned JSON schemas.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		repo, err = repository.NewSQLiteClient(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open sqlite database at %s: %s", dbPath, err.Error())
		}
		HabitService = service.NewHabitService(repo)
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
