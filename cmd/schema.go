package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/eitanoid/habit-tracker/internal/schemas"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage habit schemas",
}

var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active habit schemas",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := HabitService.ListSchemas(cmd.Context())
		if err != nil {
			return err
		}

		if len(list) == 0 {
			fmt.Println("No habit schemas found.")
			return nil
		}

		fmt.Printf("%-24s %-10s %-20s %s\n", "HABIT ID", "VERSION", "NAME", "CREATED AT")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, s := range list {
			fmt.Printf("%-24s v%-9d %-20s %s\n", s.HabitID, s.Version, s.Name, s.CreatedAt.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

var (
	schemaName string
	schemaDesc string
)

var schemaCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new habit schema",
	Example: `  habit schema create -n "Ritalin" -m "Tracking ritalin doses"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		req := schemas.NewSchemaRequest(schemaName, schemaDesc).
			WithField("Dose", "Dose taken in mg", schemas.TypeInt, true).
			WithField("Taken at", "Timestamp of when it was taken", schemas.TypeTimestamp, true)

		s, err := HabitService.CreateSchema(cmd.Context(), req)
		if err != nil {
			return err
		}

		rawJSON, _ := json.MarshalIndent(s, "", "  ")
		fmt.Printf("Created schema for '%s' (%s, v%d):\n%s\n", s.Name, s.HabitID, s.Version, string(rawJSON))
		return nil
	},
}

func init() {
	schemaCreateCmd.Flags().StringVarP(&schemaName, "name", "n", "", "Name of the habit")
	schemaCreateCmd.Flags().StringVarP(&schemaDesc, "desc", "m", "", "Description of the habit")
	_ = schemaCreateCmd.MarkFlagRequired("name")

	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaCreateCmd)
	rootCmd.AddCommand(schemaCmd)
}
