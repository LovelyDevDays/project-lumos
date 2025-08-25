package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "qdrant",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			<-sig
			fmt.Println("\nReceived interrupt signal, shutting down...")
			cancel()
		}()

		cmd.SetContext(ctx)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error executing command:", err)
		os.Exit(1)
	}
}

func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
