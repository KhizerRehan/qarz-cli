package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"qarz/internal/projection"
	"qarz/internal/store"
)

func newSyncCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Push calendar events that are still pending",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if app.Projector == nil {
				return errors.New("not signed in to Google; run 'qarz setup'")
			}

			ids, err := store.PendingLoanIDs(cmd.Context(), app.DB)
			if err != nil {
				return err
			}
			if len(ids) == 0 {
				fmt.Fprintln(app.Out, "nothing pending")
				return nil
			}

			pushed := 0
			var stuck []int64

			for _, id := range ids {
				err := app.Projector.Project(cmd.Context(), id)
				switch {
				case err == nil:
					pushed++
				case errors.Is(err, projection.ErrPending):
					stuck = append(stuck, id)
					fmt.Fprintf(app.Err, "loan %d still pending: %v\n", id, err)
				default:
					return err
				}
			}

			fmt.Fprintf(app.Out, "pushed %d of %d pending loans\n", pushed, len(ids))
			if len(stuck) > 0 {
				fmt.Fprintf(app.Out, "still pending: %v\n", stuck)
			}
			return nil
		},
	}
}
