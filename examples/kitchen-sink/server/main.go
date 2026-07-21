// Command kitchen-sink-server is a stub gRPC server for the kitchen-sink
// fixture: every RPC echoes its request.
package main

import (
	"context"
	"flag"
	"log"
	"net"

	importedv1 "github.com/braveokafor/proto-to-cli/examples/kitchen-sink/go-cobra/gen/imported/v1"
	kitchensinkv1 "github.com/braveokafor/proto-to-cli/examples/kitchen-sink/go-cobra/gen/kitchensink/v1"
	secondv1 "github.com/braveokafor/proto-to-cli/examples/kitchen-sink/go-cobra/gen/second/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fieldsServer struct {
	kitchensinkv1.UnimplementedFieldsServiceServer
}

type namesServer struct {
	kitchensinkv1.UnimplementedNamesServiceServer
}

type secondServer struct {
	secondv1.UnimplementedSecondServiceServer
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
	kitchensinkv1.RegisterNamesServiceServer(srv, namesServer{})
	secondv1.RegisterSecondServiceServer(srv, secondServer{})
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

func (fieldsServer) Messages(
	_ context.Context,
	req *kitchensinkv1.MessagesRequest,
) (*kitchensinkv1.MessagesRequest, error) {
	return req, nil
}

func (fieldsServer) Repeated(
	_ context.Context,
	req *kitchensinkv1.RepeatedRequest,
) (*kitchensinkv1.RepeatedRequest, error) {
	return req, nil
}

func (fieldsServer) Maps(
	_ context.Context,
	req *kitchensinkv1.MapsRequest,
) (*kitchensinkv1.MapsRequest, error) {
	return req, nil
}

func (fieldsServer) Wrappers(
	_ context.Context,
	req *kitchensinkv1.WrappersRequest,
) (*kitchensinkv1.WrappersRequest, error) {
	return req, nil
}

func (fieldsServer) WellKnown(
	_ context.Context,
	req *kitchensinkv1.WellKnownRequest,
) (*kitchensinkv1.WellKnownRequest, error) {
	return req, nil
}

func (fieldsServer) Enums(
	_ context.Context,
	req *kitchensinkv1.EnumsRequest,
) (*kitchensinkv1.EnumsRequest, error) {
	return req, nil
}

func (fieldsServer) Oneofs(
	_ context.Context,
	req *kitchensinkv1.OneofsRequest,
) (*kitchensinkv1.OneofsRequest, error) {
	return req, nil
}

func (fieldsServer) Optionals(
	_ context.Context,
	req *kitchensinkv1.OptionalsRequest,
) (*kitchensinkv1.OptionalsRequest, error) {
	return req, nil
}

func (namesServer) Collisions(
	_ context.Context,
	req *kitchensinkv1.CollisionsRequest,
) (*kitchensinkv1.CollisionsRequest, error) {
	return req, nil
}

func (namesServer) Imported(
	_ context.Context,
	req *importedv1.ImportedRequest,
) (*importedv1.ImportedRequest, error) {
	return req, nil
}

func (namesServer) Second(
	_ context.Context,
	req *secondv1.PingRequest,
) (*secondv1.PingRequest, error) {
	return req, nil
}

func (namesServer) Reserved(
	_ context.Context,
	req *kitchensinkv1.ReservedRequest,
) (*kitchensinkv1.ReservedRequest, error) {
	return req, nil
}

func (namesServer) Empty(
	_ context.Context,
	req *emptypb.Empty,
) (*emptypb.Empty, error) {
	return req, nil
}

func (secondServer) Ping(
	_ context.Context,
	req *secondv1.PingRequest,
) (*secondv1.PingRequest, error) {
	return req, nil
}
