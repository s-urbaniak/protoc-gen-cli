// Command kitchen-sink is the example CLI for the kitchen-sink fixture. It is
// a hand-written root that mounts the generated command tree.
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	kitchensinkv1 "github.com/braveokafor/proto-to-cli/examples/kitchen-sink/go-cobra/gen/kitchensink/v1"
	secondv1 "github.com/braveokafor/proto-to-cli/examples/kitchen-sink/go-cobra/gen/second/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50055",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	lineFmt := func(w io.Writer, next func() ([]byte, error)) error {
		for {
			rec, err := next()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "line: %s\n", rec); err != nil {
				return err
			}
		}
	}
	printers := map[string]func(io.Writer, func() ([]byte, error)) error{"line": lineFmt}
	noPretty := map[string]func(io.Writer, func() ([]byte, error)) error{
		"line":        lineFmt,
		"json-pretty": nil,
	}
	views := map[string][]string{
		"kitchensink.v1.MessagesRequest": {
			"OUTER:outer.stringLeaf",
			"DEEPEST:outer.middle.inner.deep.deeper.deepest.leaf",
		},
		"kitchensink.v1.Outer": {"LEAF:stringLeaf", "MIDDLE:middle.leaf"},
	}

	root := &cobra.Command{
		Use:   "kitchen-sink",
		Short: "Kitchen-sink fixture CLI",
	}
	root.AddCommand(
		kitchensinkv1.NewFieldsServiceCommand(
			conn,
			kitchensinkv1.FieldsServiceOptions{
				DefaultOutput: "yaml",
				Printers:      noPretty,
				Views:         views,
			},
		),
		kitchensinkv1.NewNamesServiceCommand(
			conn,
			kitchensinkv1.NamesServiceOptions{
				DefaultOutput: "yaml",
				Printers:      noPretty,
				Views:         views,
			},
		),
		kitchensinkv1.NewStreamsServiceCommand(
			conn,
			kitchensinkv1.StreamsServiceOptions{
				DefaultOutput: "yaml",
				Printers:      noPretty,
				Views:         views,
			},
		),
		kitchensinkv1.NewRelayServiceCommand(conn),
		kitchensinkv1.NewFeedServiceCommand(conn),
		kitchensinkv1.NewIngestServiceCommand(conn),
		kitchensinkv1.NewAnnotationsServiceCommand(conn),
		secondv1.NewSecondServiceCommand(conn,
			secondv1.SecondServiceOptions{DefaultOutput: "yaml", Printers: printers}),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
