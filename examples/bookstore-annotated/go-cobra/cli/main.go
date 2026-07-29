// Command bookstore-annotated is the example CLI for the bookstore-annotated
// fixture. The (cli.v0.*) options in the proto control it, not code.
package main

import (
	"log"
	"os"

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

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
