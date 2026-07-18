// Command kitchen-sink-server is a stub gRPC server for the kitchen-sink
// fixture: every RPC echoes its request.
package main

import (
	"context"
	"flag"
	"log"
	"net"

	kitchensinkv1 "github.com/braveokafor/proto-to-cli/examples/kitchen-sink/go-cobra/gen/kitchensink/v1"
	"google.golang.org/grpc"
)

type fieldsServer struct {
	kitchensinkv1.UnimplementedFieldsServiceServer
}

func main() {
	addr := flag.String("addr", ":50055", "listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	kitchensinkv1.RegisterFieldsServiceServer(srv, fieldsServer{})
	log.Printf("kitchen-sink-server listening on %s", *addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func (fieldsServer) Scalars(
	_ context.Context,
	req *kitchensinkv1.ScalarsRequest,
) (*kitchensinkv1.ScalarsRequest, error) {
	return req, nil
}
