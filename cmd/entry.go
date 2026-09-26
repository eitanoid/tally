package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

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

var entryLogCmd = &cobra.Command{
	Use:     "log",
	Short:   "Log a new entry against a tally schema",
	Example: `  tally entry log -t <TALLY_ID> -d '{"book": "The Stranger", "pages": 20}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entry, err := TallyService.RecordEntry(cmd.Context(), entryTallyID, entryData)
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
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := TallyService.ListEntries(cmd.Context(), entryTallyID, entryLimit)
		if err != nil {
			return err
		}

		if len(list) == 0 {
			fmt.Printf("No entries found for tally '%s'.\n", entryTallyID)
			return nil
		}

		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ENTRY ID\tVERSION\tCREATED AT\tDATA")
		for _, e := range list {
			fmt.Fprintf(w, "%s\tv%d\t%s\t%s\n",
				e.ID, e.SchemaVersion, e.CreatedAt.Format(time.RFC3339), e.Data)
		}
		if err := w.Flush(); err != nil {
			return err
		}
		header, body, _ := strings.Cut(buf.String(), "\n")
		fmt.Println(header)
		fmt.Println(strings.Repeat("-", len(header)))
		fmt.Print(body)

		return nil
	},
}

func init() {
	entryLogCmd.Flags().StringVarP(&entryTallyID, "tally", "t", "", "Target tally ID")
	entryLogCmd.Flags().StringVarP(&entryData, "data", "d", "", "JSON data payload string")
	_ = entryLogCmd.MarkFlagRequired("tally")
	_ = entryLogCmd.MarkFlagRequired("data")

	entryListCmd.Flags().StringVarP(&entryTallyID, "tally", "t", "", "Target tally ID")
	entryListCmd.Flags().IntVarP(&entryLimit, "limit", "l", 20, "Max entries to fetch")
	_ = entryListCmd.MarkFlagRequired("tally")

	entryCmd.AddCommand(entryLogCmd)
	entryCmd.AddCommand(entryListCmd)
	rootCmd.AddCommand(entryCmd)
}
