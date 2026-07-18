// Command kitchen-sink mounts the fixture CLI exactly as generated.
package main

import (
	"log"
	"os"

	kitchensinkv1 "github.com/braveokafor/proto-to-cli/examples/kitchen-sink/go-cobra/gen/kitchensink/v1"
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

	root := &cobra.Command{
		Use:   "kitchen-sink",
		Short: "Kitchen-sink fixture CLI",
	}
	root.AddCommand(kitchensinkv1.NewFieldsServiceCommand(conn))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
