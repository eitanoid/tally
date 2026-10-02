package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// PrintTable formats slices into aligned, header-underlined CLI tables.
func PrintTable[T any](w io.Writer, headers string, items []T, formatRow func(item T) string) error {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)

	// Write headers
	if _, err := fmt.Fprintln(tw, headers); err != nil {
		return err
	}

	// Write rows
	for _, item := range items {
		if _, err := fmt.Fprintln(tw, formatRow(item)); err != nil {
			return err
		}
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("failed to flush table writer: %w", err)
	}

	// Separate header from body to draw the divider line
	header, body, _ := strings.Cut(buf.String(), "\n")

	_, _ = fmt.Fprintln(w, header)
	_, _ = fmt.Fprintln(w, strings.Repeat("-", len(header)))
	if body != "" {
		_, _ = fmt.Fprint(w, body)
	}

	return nil
}
