// Command bookstore is the example CLI for the bookstore fixture.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"os/signal"
	"strings"

	bookstorev1 "github.com/braveokafor/protoc-gen-cli/examples/bookstore/go-cobra/gen/bookstore/v1"
	"github.com/spf13/cobra"
	"github.com/theory/jsonpath"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func decodeTitles(r io.Reader) iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		lines := bufio.NewScanner(r)
		for lines.Scan() {
			title := strings.TrimSpace(lines.Text())
			if title == "" {
				continue
			}
			body, err := json.Marshal(map[string]any{
				"book": map[string]any{"title": title},
			})
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(body, nil) {
				return
			}
		}
		if err := lines.Err(); err != nil {
			yield(nil, err)
		}
	}
}

func csvRows(w io.Writer, fields []bookstorev1.AuctionsServiceViewField, rows []any) error {
	cw := csv.NewWriter(w)
	labels := make([]string, len(fields))
	columns := make([]*jsonpath.Path, len(fields))
	for i, f := range fields {
		labels[i] = f.Label
		path, err := jsonpath.Parse(f.Path)
		if err != nil {
			return err
		}
		columns[i] = path
	}
	if err := cw.Write(labels); err != nil {
		return err
	}
	for _, row := range rows {
		cells := make([]string, len(fields))
		for i, path := range columns {
			if found := path.Select(row); len(found) > 0 {
				cells[i] = fmt.Sprint(found[0])
			}
		}
		if err := cw.Write(cells); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func printCSV(w io.Writer, v bookstorev1.AuctionsServiceView, body []byte) error {
	// --dry-run and --example print a request body, which has no view.
	if len(v.Fields) == 0 && len(v.Lists) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return err
	}
	if len(v.Lists) == 0 {
		return csvRows(w, v.Fields, []any{doc})
	}
	for _, list := range v.Lists {
		path, err := jsonpath.Parse(list.Path)
		if err != nil {
			return err
		}
		if err := csvRows(w, list.Fields, path.Select(doc)); err != nil {
			return err
		}
	}
	return nil
}

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

	auctionColumns := []bookstorev1.AuctionsServiceViewField{
		{Label: "ID", Path: "$.id"},
		{Label: "TITLE", Path: "$.lot.book.title"},
		{Label: "AUTHOR", Path: "$.lot.book.author"},
		{Label: "CONDITION", Path: "$.lot.condition"},
		{Label: "FLAWS", Path: "$.lot.flaws[*].kind"},
		{Label: "HIGH BID", Path: "$.highBid"},
		{Label: "BIDDER", Path: "$.highBidder"},
		{Label: "ENDS", Path: "$.endsAt"},
	}
	auctionOpts := bookstorev1.AuctionsServiceOptions{
		Printers: map[string]bookstorev1.AuctionsServicePrinter{
			"csv": printCSV,
		},
		Views: map[string]bookstorev1.AuctionsServiceView{
			"bookstore.v1.AuctionsService.CreateAuction": {Fields: auctionColumns},
			"bookstore.v1.AuctionsService.ListAuctions": {
				Lists: []bookstorev1.AuctionsServiceViewList{
					{Label: "auctions", Path: "$.auctions[*]", Fields: auctionColumns},
				},
			},
		},
	}

	catalog := bookstorev1.NewBookstoreServiceCommand(conn)
	// The generated name "bookstore" repeats under the root.
	catalog.Use = "catalog"
	root.AddCommand(catalog)
	root.AddCommand(bookstorev1.NewAuctionsServiceCommand(conn, auctionOpts))
	root.AddCommand(
		bookstorev1.NewInventoryServiceCommand(conn, bookstorev1.InventoryServiceOptions{
			// The built-ins read a JSON object. A plain title reaches titles.
			Decoders: []bookstorev1.InventoryServiceDecoder{
				bookstorev1.InventoryServiceDecoderJSON,
				bookstorev1.InventoryServiceDecoderYAML,
				{Name: "titles", Decode: decodeTitles},
			},
		}),
	)

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
