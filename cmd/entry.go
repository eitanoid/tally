package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Log and view habit entries",
}

var (
	entryHabitID string
	entryData    string
	entryLimit   int
)

var entryLogCmd = &cobra.Command{
	Use:     "log",
	Short:   "Log a new entry against a habit schema",
	Example: `  habit entry log -h <HABIT_ID> -j '{"Dose": 20, "Taken at": "2026-09-25T10:00:00Z"}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entry, err := HabitService.RecordEntry(cmd.Context(), entryHabitID, entryData)
		if err != nil {
			return fmt.Errorf("failed to log entry: %w", err)
		}

		fmt.Printf("Logged entry %s for habit '%s' (v%d) at %s\n",
			entry.ID, entry.HabitID, entry.SchemaVersion, entry.CreatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var entryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entries logged for a habit",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := HabitService.ListEntries(cmd.Context(), entryHabitID, entryLimit)
		if err != nil {
			return err
		}

		if len(list) == 0 {
			fmt.Printf("No entries found for habit '%s'.\n", entryHabitID)
			return nil
		}

		fmt.Printf("%-27s %-10s %-20s %s\n", "ENTRY ID", "VERSION", "CREATED AT", "DATA")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, e := range list {
			fmt.Printf("%-27s v%-9d %-20s %s\n",
				e.ID, e.SchemaVersion, e.CreatedAt.Format("2006-01-02 15:04:05"), e.Data)
		}
		return nil
	},
}

func init() {
	entryLogCmd.Flags().StringVarP(&entryHabitID, "habit", "t", "", "Target habit ID")
	entryLogCmd.Flags().StringVarP(&entryData, "data", "j", "", "JSON data payload string")
	_ = entryLogCmd.MarkFlagRequired("habit")
	_ = entryLogCmd.MarkFlagRequired("data")

	entryListCmd.Flags().StringVarP(&entryHabitID, "habit", "z", "", "Target habit ID")
	entryListCmd.Flags().IntVarP(&entryLimit, "limit", "l", 20, "Max entries to fetch")
	_ = entryListCmd.MarkFlagRequired("habit")

	entryCmd.AddCommand(entryLogCmd)
	entryCmd.AddCommand(entryListCmd)
	rootCmd.AddCommand(entryCmd)
}
