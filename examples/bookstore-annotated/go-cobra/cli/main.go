// Command bookstore-annotated is the example CLI for the bookstore-annotated
// fixture. The (cli.v0.*) options in the proto control it, not code.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	bookstoreannotatedv1 "github.com/braveokafor/proto-to-cli/examples/bookstore-annotated/go-cobra/gen/bookstore/annotated/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50054",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	root := &cobra.Command{
		Use:   "bookstore",
		Short: "Bookstore CLI",
	}

	root.AddCommand(bookstoreannotatedv1.NewBookstoreServiceCommand(conn))
	root.AddCommand(bookstoreannotatedv1.NewAuctionsServiceCommand(conn))
	root.AddCommand(bookstoreannotatedv1.NewInventoryServiceCommand(conn))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := root.ExecuteContext(ctx); err != nil {
		var coded interface{ ExitCode() int }
		if errors.As(err, &coded) {
			os.Exit(coded.ExitCode())
		}
		os.Exit(1)
	}
}
