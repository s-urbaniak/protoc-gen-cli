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

	auctionOpts := bookstorev1.AuctionsServiceOptions{
		Views: map[string][]string{
			"bookstore.v1.Auction": {
				"ID:id",
				"TITLE:lot.book.title",
				"AUTHOR:lot.book.author",
				"CONDITION:lot.condition",
				"FLAWS:lot.flaws.#",
				"HIGH BID:highBid",
				"BIDDER:highBidder",
				"ENDS:endsAt",
			},
		},
	}

	catalog := bookstorev1.NewBookstoreServiceCommand(conn)
	// Renamed in code: the generated name "bookstore" would stutter under the root.
	catalog.Use = "catalog"
	root.AddCommand(catalog)
	root.AddCommand(bookstorev1.NewAuctionsServiceCommand(conn, auctionOpts))
	root.AddCommand(bookstorev1.NewInventoryServiceCommand(conn))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
