// Command bookstore is the example CLI for the bookstore fixture: a
// hand-written root that mounts the generated command tree.
package main

import (
	"log"
	"os"

	bookstorev1 "github.com/braveokafor/proto-to-cli/examples/bookstore/go-cobra/gen/bookstore/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	root := &cobra.Command{
		Use:   "bookstore",
		Short: "Bookstore CLI",
	}

	// Flatten Bookstore service onto root; as a group it would stutter ("bookstore bookstore get-book").
	books := bookstorev1.NewBookstoreServiceCommand(conn)
	root.AddCommand(books.Commands()...)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
