// Command kitchen-sink is the example CLI for the kitchen-sink fixture: a
// hand-written root that mounts the generated command tree.
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

// lineFormat writes each record on one line, with a prefix.
type lineFormat struct{}

func (lineFormat) Format(w io.Writer, r secondv1.Records) error {
	for {
		rec, err := r.Next()
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

func main() {
	conn, err := grpc.NewClient("localhost:50055",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	root := &cobra.Command{
		Use:   "kitchen-sink",
		Short: "Kitchen-sink fixture CLI",
	}
	root.AddCommand(
		kitchensinkv1.NewFieldsServiceCommand(conn),
		kitchensinkv1.NewNamesServiceCommand(conn),
		secondv1.NewSecondServiceCommand(conn),
	)

	secondv1.RegisterFormat("line", lineFormat{})
	secondv1.UnregisterFormat("json-pretty")
	secondv1.SetDefaultFormat("yaml")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
