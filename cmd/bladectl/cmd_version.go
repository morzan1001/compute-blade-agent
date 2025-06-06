package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

func init() {
	rootCmd.AddCommand(cmdVersion)
}

var cmdVersion = &cobra.Command{
	Use:     "version",
	Short:   "Shows version information",
	Example: "bladectl version",
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		clients := clientsFromContext(ctx)

		header := []string{
			"Component",
			"Version",
			"Commit",
			"Build Time",
		}

		// Table writer setup
		tbl := tablewriter.NewTable(os.Stdout,
			tablewriter.WithHeader(header),
			tablewriter.WithHeaderAlignment(tw.AlignLeft),
			tablewriter.WithHeaderAutoFormat(tw.Off),
		)

		commit := Commit
		if len(commit) > 7 {
			commit = commit[:7]
		}

		_ = tbl.Append([]string{"bladectl", Version, commit, BuildTime.Format(time.RFC3339)})

		var wg sync.WaitGroup
		for idx, client := range clients {
			wg.Add(1)
			go func() {
				defer wg.Done()

				if status, err := client.GetStatus(ctx, &emptypb.Empty{}); err == nil && status.Version != nil {
					commit := status.Version.Commit
					if len(commit) > 7 {
						commit = commit[:7]
					}

					_ = tbl.Append([]string{
						fmt.Sprintf("api: %s", bladeNames[idx]),
						status.Version.Version,
						commit,
						time.Unix(status.Version.Date, 0).Format(time.RFC3339),
					})
				} else {
					log.Printf("Error (%s) getting status: %v", bladeNames[idx], err)
				}
			}()
		}

		wg.Wait()

		_ = tbl.Render()

		return nil
	},
}
