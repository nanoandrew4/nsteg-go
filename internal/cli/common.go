package cli

import (
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"time"
)

func MarkFlagsRequired(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
}

func NewSpinner() *spinner.Spinner {
	return spinner.New(spinner.CharSets[4], 100*time.Millisecond)
}
