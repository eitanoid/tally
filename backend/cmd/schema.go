package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eitanoid/tally/internal/schemas"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage tally schemas",
}

var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active tally schemas",
	RunE: func(cmd *cobra.Command, _ []string) error {
		list, err := tallyService.ListSchemas(cmd.Context())
		if err != nil {
			return err
		}

		if len(list) == 0 {
			fmt.Println("No tally schemas found.")
			return nil
		}

		PrintTable(cmd.OutOrStdout(), "TALLY ID\tVERSION\tNAME\tCREATED AT\tSCHEMA", list, func(s schemas.TallySchema) string {
			return fmt.Sprintf("%s\tv%d\t%s\t%s\t%s", s.TallyID, s.Version, s.Name, s.CreatedAt.Format(time.RFC3339), s.JSONSchemaRaw[:100]+"...")

		})
		return nil
	},
}

var schemaGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an active tally schemas",
	RunE: func(cmd *cobra.Command, args []string) error {
		var tallyID string
		if len(args) >= 1 {
			tallyID = args[0]
		} else {
			return errors.New("expected tally id. got nothing")
		}
		s, err := tallyService.GetLatestSchema(cmd.Context(), tallyID)
		if err != nil {
			return err
		}

		PrintTable(cmd.OutOrStdout(), "TALLY ID\tVERSION\tNAME\tCREATED AT\tSCHEMA", []*schemas.TallySchema{s}, func(s *schemas.TallySchema) string {
			return fmt.Sprintf("%s\tv%d\t%s\t%s\t%s", s.TallyID, s.Version, s.Name, s.CreatedAt.Format(time.RFC3339), s.JSONSchemaRaw)
		})
		return nil
	},
}

var (
	schemaName   string
	schemaDesc   string
	schemaFields []string
)

var schemaCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new tally schema",
	Example: `  tally schema create -n "Reading" -m "Tracking reading habits" -f "book:Name of the book:string:true" -f "pages:number of pages read:integer:true"`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		req := schemas.NewSchemaRequest(schemaName, schemaDesc)
		for _, field := range schemaFields {
			name, desc, typ, required, err := processField(field)
			if err != nil {
				return err
			}
			req.WithField(name, desc, schemas.SupportedType(typ), required)
		}

		s, err := tallyService.CreateSchema(cmd.Context(), req)
		if err != nil {
			return err
		}

		rawJSON, _ := json.MarshalIndent(s, "", "  ")
		fmt.Printf("Created schema for '%s' (%s, v%d):\n%s\n", s.Name, s.TallyID, s.Version, string(rawJSON))
		return nil
	},
}

// a field is defined as name:description:type:required
func processField(field string) (name string, description string, fieldType string, required bool, err error) {
	fields := strings.Split(field, ":")
	if len(fields) != 4 {
		return "", "", "", false, fmt.Errorf("invalid field format '%s': expected 4 colon-separated parts (name:description:type:required), got %d", field, len(fields))
	}

	name = strings.TrimSpace(fields[0])
	description = strings.TrimSpace(fields[1])
	fieldType = strings.TrimSpace(fields[2])
	reqStr := strings.TrimSpace(fields[3])

	if name == "" {
		return "", "", "", false, fmt.Errorf("field name cannot be empty in '%s'", field)
	}
	if fieldType == "" {
		return "", "", "", false, fmt.Errorf("field type cannot be empty in '%s'", field)
	}

	required, err = strconv.ParseBool(reqStr)
	if err != nil {
		return "", "", "", false, fmt.Errorf("invalid boolean value '%s' for required flag in '%s': %w", reqStr, field, err)
	}

	return name, description, fieldType, required, nil
}
func init() {
	schemaCreateCmd.Flags().StringVarP(&schemaName, "name", "n", "", "Name of the tally")
	schemaCreateCmd.Flags().StringVarP(&schemaDesc, "desc", "m", "", "Description of the tally")
	schemaCreateCmd.Flags().StringSliceVarP(&schemaFields, "field", "f", []string{}, "Fields to define")

	_ = schemaCreateCmd.MarkFlagRequired("name")

	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaCreateCmd)
	schemaCmd.AddCommand(schemaGetCmd)
	rootCmd.AddCommand(schemaCmd)
}
