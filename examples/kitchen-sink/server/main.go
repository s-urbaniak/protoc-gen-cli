// Command kitchen-sink-server is a stub gRPC server for the kitchen-sink
// fixture.
package main

import (
	"context"
	"flag"
	"io"
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

type streamsServer struct {
	kitchensinkv1.UnimplementedStreamsServiceServer
}

type relayServer struct {
	kitchensinkv1.UnimplementedRelayServiceServer
}

type feedServer struct {
	kitchensinkv1.UnimplementedFeedServiceServer
}

type ingestServer struct {
	kitchensinkv1.UnimplementedIngestServiceServer
}

type annotationsServer struct {
	kitchensinkv1.UnimplementedAnnotationsServiceServer
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
	kitchensinkv1.RegisterStreamsServiceServer(srv, streamsServer{})
	kitchensinkv1.RegisterRelayServiceServer(srv, relayServer{})
	kitchensinkv1.RegisterFeedServiceServer(srv, feedServer{})
	kitchensinkv1.RegisterIngestServiceServer(srv, ingestServer{})
	kitchensinkv1.RegisterAnnotationsServiceServer(srv, annotationsServer{})
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

func (fieldsServer) Collections(
	_ context.Context,
	req *kitchensinkv1.CollectionsRequest,
) (*kitchensinkv1.CollectionsRequest, error) {
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

func (namesServer) HTTPCall(
	_ context.Context,
	req *kitchensinkv1.CasingRequest,
) (*kitchensinkv1.CasingRequest, error) {
	return req, nil
}

func (namesServer) LowerSnake(
	_ context.Context,
	req *kitchensinkv1.CasingRequest,
) (*kitchensinkv1.CasingRequest, error) {
	return req, nil
}

func (namesServer) Borrow(
	_ context.Context,
	req *kitchensinkv1.Outer,
) (*kitchensinkv1.Outer, error) {
	return req, nil
}

func (namesServer) Nested(
	_ context.Context,
	req *kitchensinkv1.Envelope_Letter,
) (*kitchensinkv1.Envelope_Letter, error) {
	return req, nil
}

func (secondServer) Ping(
	_ context.Context,
	req *secondv1.PingRequest,
) (*secondv1.PingRequest, error) {
	return req, nil
}

func (streamsServer) Unary(
	_ context.Context,
	req *kitchensinkv1.StreamsRequest,
) (*kitchensinkv1.StreamsResponse, error) {
	return &kitchensinkv1.StreamsResponse{Text: req.GetText()}, nil
}

func (streamsServer) ServerStream(
	req *kitchensinkv1.StreamsRequest,
	stream kitchensinkv1.StreamsService_ServerStreamServer,
) error {
	n := req.GetCount()
	if n <= 0 {
		n = 3
	}
	for i := int32(0); i < n; i++ {
		if err := stream.Send(
			&kitchensinkv1.StreamsResponse{Text: req.GetText(), Index: i},
		); err != nil {
			return err
		}
	}
	return nil
}

func (streamsServer) ClientStream(stream kitchensinkv1.StreamsService_ClientStreamServer) error {
	var count int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(
				&kitchensinkv1.StreamsResponse{Text: "received", Index: count},
			)
		}
		if err != nil {
			return err
		}
		count++
	}
}

func (streamsServer) BidiStream(stream kitchensinkv1.StreamsService_BidiStreamServer) error {
	var i int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(
			&kitchensinkv1.StreamsResponse{Text: req.GetText(), Index: i},
		); err != nil {
			return err
		}
		i++
	}
}

func (streamsServer) EmptyStream(
	_ *emptypb.Empty,
	stream kitchensinkv1.StreamsService_EmptyStreamServer,
) error {
	for i := int32(0); i < 3; i++ {
		if err := stream.Send(&kitchensinkv1.StreamsResponse{Index: i}); err != nil {
			return err
		}
	}
	return nil
}

func (relayServer) Chat(stream kitchensinkv1.RelayService_ChatServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&kitchensinkv1.ChatNote{Text: req.GetText()}); err != nil {
			return err
		}
	}
}

func (feedServer) Tail(
	req *kitchensinkv1.TailRequest,
	stream kitchensinkv1.FeedService_TailServer,
) error {
	n := req.GetCount()
	if n <= 0 {
		n = 3
	}
	for i := int32(0); i < n; i++ {
		if err := stream.Send(&kitchensinkv1.FeedEvent{Index: i, Note: "tick"}); err != nil {
			return err
		}
	}
	return nil
}

func (ingestServer) Ingest(stream kitchensinkv1.IngestService_IngestServer) error {
	var count int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&kitchensinkv1.IngestSummary{Received: count})
		}
		if err != nil {
			return err
		}
		count++
	}
}

func (ingestServer) Absorb(stream kitchensinkv1.IngestService_AbsorbServer) error {
	var count int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&kitchensinkv1.IngestSummary{Received: count})
		}
		if err != nil {
			return err
		}
		count++
	}
}

func (annotationsServer) Echo(
	_ context.Context,
	req *kitchensinkv1.EchoRequest,
) (*kitchensinkv1.EchoRequest, error) {
	return req, nil
}

func (annotationsServer) Knobs(
	_ context.Context,
	req *kitchensinkv1.KnobsRequest,
) (*kitchensinkv1.KnobsRequest, error) {
	return req, nil
}

func (annotationsServer) Curated(
	_ context.Context,
	req *kitchensinkv1.CuratedRequest,
) (*kitchensinkv1.CuratedRequest, error) {
	return req, nil
}

func (annotationsServer) Flat(
	_ context.Context,
	req *kitchensinkv1.FlatRequest,
) (*kitchensinkv1.FlatRequest, error) {
	return req, nil
}
