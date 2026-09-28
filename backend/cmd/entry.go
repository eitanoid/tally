package cmd

import (
	"fmt"
	"time"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/spf13/cobra"
)

var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Log and view tally entries",
}

var (
	entryTallyID string
	entryData    string
	entryLimit   int
)

var entryAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a new entry against a tally schema",
	Example: `  tally entry add -t <TALLY_ID> -d '{"book": "The Stranger", "pages": 20}'`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		entry, err := tallyService.RecordEntry(cmd.Context(), entryTallyID, entryData)
		if err != nil {
			return fmt.Errorf("failed to log entry: %w", err)
		}

		fmt.Printf("Logged entry %s for tally '%s' (v%d) at %s\n",
			entry.ID, entry.TallyID, entry.SchemaVersion, entry.CreatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var entryPatchCmd = &cobra.Command{
	Use:     "patch",
	Short:   "Update an existing entry data",
	Example: `  tally entry patch -t <ENTRY_ID> -d '{"book": "The Stranger", "pages": 15}'`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		entry, err := tallyService.UpdateEntry(cmd.Context(), entryTallyID, entryData)
		if err != nil {
			return fmt.Errorf("failed to log entry: %w", err)
		}

		fmt.Printf("Logged entry %s for tally '%s' (v%d) at %s\n",
			entry.ID, entry.TallyID, entry.SchemaVersion, entry.CreatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var entryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entries logged for a tally",
	RunE: func(cmd *cobra.Command, _ []string) error {
		paginatedResults, err := tallyService.ListEntries(cmd.Context(), entryTallyID, entryLimit, 0)
		if err != nil {
			return err
		}
		list := paginatedResults.Entries

		if len(list) == 0 {
			fmt.Printf("No entries found for tally '%s'.\n", entryTallyID)
			return nil
		}

		headers := "ENTRY ID\tVERSION\tCREATED AT\tDATA"
		return PrintTable(cmd.OutOrStdout(), headers, list, func(e entries.TallyEntry) string {
			return fmt.Sprintf("%s\tv%d\t%s\t%s",
				e.ID,
				e.SchemaVersion,
				e.CreatedAt.Format(time.RFC3339),
				e.Data,
			)
		})
	},
}

func init() {
	entryAddCmd.Flags().StringVarP(&entryTallyID, "target", "t", "", "Target tally ID")
	entryAddCmd.Flags().StringVarP(&entryData, "data", "d", "", "JSON data payload string")
	_ = entryAddCmd.MarkFlagRequired("target")
	_ = entryAddCmd.MarkFlagRequired("data")

	entryListCmd.Flags().StringVarP(&entryTallyID, "target", "t", "", "Target tally ID")
	entryListCmd.Flags().IntVarP(&entryLimit, "limit", "l", 20, "Max entries to fetch")
	_ = entryListCmd.MarkFlagRequired("target")

	entryPatchCmd.Flags().StringVarP(&entryTallyID, "target", "t", "", "Target entry ID")
	entryPatchCmd.Flags().StringVarP(&entryData, "data", "d", "", "JSON patch data string")
	_ = entryListCmd.MarkFlagRequired("target")
	_ = entryListCmd.MarkFlagRequired("data")

	entryCmd.AddCommand(entryAddCmd)
	entryCmd.AddCommand(entryListCmd)
	entryCmd.AddCommand(entryPatchCmd)
	rootCmd.AddCommand(entryCmd)
}
