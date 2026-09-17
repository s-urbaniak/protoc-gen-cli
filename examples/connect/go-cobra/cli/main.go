// Command connect-example proves generated Connect clients satisfy the CLI API.
package main

import (
	"net/http"

	"connectrpc.com/connect"
	connectv1 "github.com/braveokafor/protoc-gen-cli/examples/connect/go-cobra/gen/connect/v1"
	connectv1connect "github.com/braveokafor/protoc-gen-cli/examples/connect/go-cobra/gen/connect/v1/connectv1connect"
	"github.com/spf13/cobra"
)

func main() {
	client := connectv1connect.NewStreamServiceClient(http.DefaultClient, "http://localhost:8080")
	var _ connectv1.StreamServiceCLIClient = client
	root := &cobra.Command{Use: "connect-example"}
	root.AddCommand(connectv1.NewStreamServiceCommand(client))
	_ = root.Execute()
}

var _ = connect.ProtocolConnect
