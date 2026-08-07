// Command kitchen-sink is the example CLI for the kitchen-sink fixture.
package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	kitchensinkv1 "github.com/braveokafor/protoc-gen-cli/examples/kitchen-sink/go-cobra/gen/kitchensink/v1"
	secondv1 "github.com/braveokafor/protoc-gen-cli/examples/kitchen-sink/go-cobra/gen/second/v1"
	"github.com/spf13/cobra"
	"github.com/theory/jsonpath"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func printLine(w io.Writer, body []byte) error {
	_, err := fmt.Fprintf(w, "line: %s\n", body)
	return err
}

func printCSV(w io.Writer, fields []kitchensinkv1.FieldsServiceViewField, body []byte) error {
	// --dry-run and --example print a request body, which has no view.
	if len(fields) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	labels := make([]string, len(fields))
	cells := make([]string, len(fields))
	for i, f := range fields {
		labels[i] = f.Label
		path, err := jsonpath.Parse(f.Path)
		if err != nil {
			return err
		}
		if found := path.Select(doc); len(found) > 0 {
			cells[i] = fmt.Sprint(found[0])
		}
	}
	if err := cw.Write(labels); err != nil {
		return err
	}
	if err := cw.Write(cells); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

func main() {
	conn, err := grpc.NewClient("localhost:50055",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	fieldsPrinters := map[string]kitchensinkv1.FieldsServicePrinter{
		"line": func(w io.Writer, _ kitchensinkv1.FieldsServiceView, body []byte) error {
			return printLine(w, body)
		},
		"csv": func(w io.Writer, v kitchensinkv1.FieldsServiceView, body []byte) error {
			return printCSV(w, v.Fields, body)
		},
		"json": nil,
	}
	namesPrinters := map[string]kitchensinkv1.NamesServicePrinter{
		"line": func(w io.Writer, _ kitchensinkv1.NamesServiceView, body []byte) error {
			return printLine(w, body)
		},
		"json": nil,
	}
	streamsPrinters := map[string]kitchensinkv1.StreamsServicePrinter{
		"line": func(w io.Writer, _ kitchensinkv1.StreamsServiceView, body []byte) error {
			return printLine(w, body)
		},
		"json": nil,
	}
	secondPrinters := map[string]secondv1.SecondServicePrinter{
		"line": func(w io.Writer, _ secondv1.SecondServiceView, body []byte) error {
			return printLine(w, body)
		},
	}

	views := map[string]kitchensinkv1.FieldsServiceView{
		"kitchensink.v1.FieldsService.Scalars": {Fields: []kitchensinkv1.FieldsServiceViewField{
			{Label: "ID", Path: "$.int64Field"},
			{Label: "NAME", Path: "$.stringField"},
			{Label: "ACTIVE", Path: "$.boolField"},
		}},
		"kitchensink.v1.FieldsService.Messages": {Fields: []kitchensinkv1.FieldsServiceViewField{
			{Label: "LEAF", Path: "$.stringLeaf"},
			{Label: "MIDDLE", Path: "$.middle.leaf"},
		}},
	}

	root := &cobra.Command{
		Use:   "kitchen-sink",
		Short: "Kitchen-sink fixture CLI",
	}
	root.AddCommand(
		kitchensinkv1.NewFieldsServiceCommand(
			conn,
			kitchensinkv1.FieldsServiceOptions{
				DefaultPrinter: "yaml",
				Printers:       fieldsPrinters,
				Views:          views,
			},
		),
		kitchensinkv1.NewNamesServiceCommand(
			conn,
			kitchensinkv1.NamesServiceOptions{
				DefaultPrinter: "yaml",
				Printers:       namesPrinters,
			},
		),
		kitchensinkv1.NewStreamsServiceCommand(
			conn,
			kitchensinkv1.StreamsServiceOptions{
				DefaultPrinter: "yaml",
				Printers:       streamsPrinters,
			},
		),
		kitchensinkv1.NewRelayServiceCommand(conn),
		kitchensinkv1.NewFeedServiceCommand(conn),
		kitchensinkv1.NewIngestServiceCommand(conn),
		kitchensinkv1.NewAnnotationsServiceCommand(conn),
		secondv1.NewSecondServiceCommand(conn,
			secondv1.SecondServiceOptions{DefaultPrinter: "yaml", Printers: secondPrinters}),
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
