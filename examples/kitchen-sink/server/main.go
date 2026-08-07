// Command kitchen-sink-server is a stub gRPC server for the kitchen-sink
// fixture.
package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"net"
	"os"

	importedv1 "github.com/braveokafor/protoc-gen-cli/examples/kitchen-sink/go-cobra/gen/imported/v1"
	kitchensinkv1 "github.com/braveokafor/protoc-gen-cli/examples/kitchen-sink/go-cobra/gen/kitchensink/v1"
	secondv1 "github.com/braveokafor/protoc-gen-cli/examples/kitchen-sink/go-cobra/gen/second/v1"
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

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		slog.Error("listen", "err", err)
		os.Exit(1)
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
	slog.Info("kitchen-sink-server listening", "addr", *addr)
	if err := srv.Serve(lis); err != nil {
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
}

func (fieldsServer) Scalars(
	_ context.Context,
	req *kitchensinkv1.ScalarsRequest,
) (*kitchensinkv1.ScalarsRequest, error) {
	slog.Info("scalars")
	return req, nil
}

func (fieldsServer) Messages(
	_ context.Context,
	req *kitchensinkv1.MessagesRequest,
) (*kitchensinkv1.MessagesRequest, error) {
	slog.Info("messages")
	return req, nil
}

func (fieldsServer) Repeated(
	_ context.Context,
	req *kitchensinkv1.RepeatedRequest,
) (*kitchensinkv1.RepeatedRequest, error) {
	slog.Info("repeated")
	return req, nil
}

func (fieldsServer) Maps(
	_ context.Context,
	req *kitchensinkv1.MapsRequest,
) (*kitchensinkv1.MapsRequest, error) {
	slog.Info("maps")
	return req, nil
}

func (fieldsServer) Wrappers(
	_ context.Context,
	req *kitchensinkv1.WrappersRequest,
) (*kitchensinkv1.WrappersRequest, error) {
	slog.Info("wrappers")
	return req, nil
}

func (fieldsServer) WellKnown(
	_ context.Context,
	req *kitchensinkv1.WellKnownRequest,
) (*kitchensinkv1.WellKnownRequest, error) {
	slog.Info("well known")
	return req, nil
}

func (fieldsServer) Enums(
	_ context.Context,
	req *kitchensinkv1.EnumsRequest,
) (*kitchensinkv1.EnumsRequest, error) {
	slog.Info("enums")
	return req, nil
}

func (fieldsServer) Oneofs(
	_ context.Context,
	req *kitchensinkv1.OneofsRequest,
) (*kitchensinkv1.OneofsRequest, error) {
	slog.Info("oneofs")
	return req, nil
}

func (fieldsServer) Optionals(
	_ context.Context,
	req *kitchensinkv1.OptionalsRequest,
) (*kitchensinkv1.OptionalsRequest, error) {
	slog.Info("optionals")
	return req, nil
}

func (fieldsServer) Collections(
	_ context.Context,
	req *kitchensinkv1.CollectionsRequest,
) (*kitchensinkv1.CollectionsRequest, error) {
	slog.Info("collections")
	return req, nil
}

func (namesServer) Collisions(
	_ context.Context,
	req *kitchensinkv1.CollisionsRequest,
) (*kitchensinkv1.CollisionsRequest, error) {
	slog.Info("collisions")
	return req, nil
}

func (namesServer) Imported(
	_ context.Context,
	req *importedv1.ImportedRequest,
) (*importedv1.ImportedRequest, error) {
	slog.Info("imported")
	return req, nil
}

func (namesServer) Second(
	_ context.Context,
	req *secondv1.PingRequest,
) (*secondv1.PingRequest, error) {
	slog.Info("second")
	return req, nil
}

func (namesServer) Reserved(
	_ context.Context,
	req *kitchensinkv1.ReservedRequest,
) (*kitchensinkv1.ReservedRequest, error) {
	slog.Info("reserved")
	return req, nil
}

func (namesServer) Empty(
	_ context.Context,
	req *emptypb.Empty,
) (*emptypb.Empty, error) {
	slog.Info("empty")
	return req, nil
}

func (namesServer) HTTPCall(
	_ context.Context,
	req *kitchensinkv1.CasingRequest,
) (*kitchensinkv1.CasingRequest, error) {
	slog.Info("http call")
	return req, nil
}

func (namesServer) LowerSnake(
	_ context.Context,
	req *kitchensinkv1.CasingRequest,
) (*kitchensinkv1.CasingRequest, error) {
	slog.Info("lower snake")
	return req, nil
}

func (namesServer) Borrow(
	_ context.Context,
	req *kitchensinkv1.Outer,
) (*kitchensinkv1.Outer, error) {
	slog.Info("borrow")
	return req, nil
}

func (namesServer) Nested(
	_ context.Context,
	req *kitchensinkv1.Envelope_Letter,
) (*kitchensinkv1.Envelope_Letter, error) {
	slog.Info("nested")
	return req, nil
}

func (secondServer) Ping(
	_ context.Context,
	req *secondv1.PingRequest,
) (*secondv1.PingRequest, error) {
	slog.Info("ping")
	return req, nil
}

func (streamsServer) Unary(
	_ context.Context,
	req *kitchensinkv1.StreamsRequest,
) (*kitchensinkv1.StreamsResponse, error) {
	slog.Info("unary")
	return &kitchensinkv1.StreamsResponse{Text: req.GetText()}, nil
}

func (streamsServer) ServerStream(
	req *kitchensinkv1.StreamsRequest,
	stream kitchensinkv1.StreamsService_ServerStreamServer,
) error {
	slog.Info("server stream")
	n := req.GetCount()
	if n <= 0 {
		n = 3
	}
	for i := int32(0); i < n; i++ {
		slog.Info("send", "index", i, "text", req.GetText())
		if err := stream.Send(
			&kitchensinkv1.StreamsResponse{Text: req.GetText(), Index: i},
		); err != nil {
			return err
		}
	}
	return nil
}

func (streamsServer) ClientStream(stream kitchensinkv1.StreamsService_ClientStreamServer) error {
	slog.Info("client stream")
	var count int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			slog.Info("client stream done", "received", count)
			return stream.SendAndClose(
				&kitchensinkv1.StreamsResponse{Text: "received", Index: count},
			)
		}
		if err != nil {
			return err
		}
		count++
		slog.Info("recv", "n", count)
	}
}

func (streamsServer) BidiStream(stream kitchensinkv1.StreamsService_BidiStreamServer) error {
	slog.Info("bidi stream")
	var i int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		slog.Info("echo", "index", i, "text", req.GetText())
		if err := stream.Send(
			&kitchensinkv1.StreamsResponse{Text: req.GetText(), Index: i},
		); err != nil {
			return err
		}
		i++
	}
}

func (streamsServer) ClientStreamEarlyReturn(
	stream kitchensinkv1.StreamsService_ClientStreamEarlyReturnServer,
) error {
	slog.Info("client stream early return")
	if _, err := stream.Recv(); err != nil {
		return err
	}
	return stream.SendAndClose(&kitchensinkv1.StreamsResponse{Text: "enough", Index: 1})
}

func (streamsServer) BidiStreamEarlyReturn(
	stream kitchensinkv1.StreamsService_BidiStreamEarlyReturnServer,
) error {
	slog.Info("bidi stream early return")
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	return stream.Send(&kitchensinkv1.StreamsResponse{Text: req.GetText(), Index: 0})
}

func (streamsServer) EmptyStream(
	_ *emptypb.Empty,
	stream kitchensinkv1.StreamsService_EmptyStreamServer,
) error {
	slog.Info("empty stream")
	for i := int32(0); i < 3; i++ {
		if err := stream.Send(&kitchensinkv1.StreamsResponse{Index: i}); err != nil {
			return err
		}
	}
	return nil
}

func (relayServer) Chat(stream kitchensinkv1.RelayService_ChatServer) error {
	slog.Info("chat")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		slog.Info("chat echo", "text", req.GetText())
		if err := stream.Send(&kitchensinkv1.ChatNote{Text: req.GetText()}); err != nil {
			return err
		}
	}
}

func (feedServer) Tail(
	req *kitchensinkv1.TailRequest,
	stream kitchensinkv1.FeedService_TailServer,
) error {
	slog.Info("tail")
	n := req.GetCount()
	if n <= 0 {
		n = 3
	}
	for i := int32(0); i < n; i++ {
		slog.Info("tick", "index", i)
		if err := stream.Send(&kitchensinkv1.FeedEvent{Index: i, Note: "tick"}); err != nil {
			return err
		}
	}
	return nil
}

func (ingestServer) Ingest(stream kitchensinkv1.IngestService_IngestServer) error {
	slog.Info("ingest")
	var count int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			slog.Info("ingest done", "received", count)
			return stream.SendAndClose(&kitchensinkv1.IngestSummary{Received: count})
		}
		if err != nil {
			return err
		}
		count++
		slog.Info("ingest record", "n", count)
	}
}

func (ingestServer) Absorb(stream kitchensinkv1.IngestService_AbsorbServer) error {
	slog.Info("absorb")
	var count int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			slog.Info("absorb done", "received", count)
			return stream.SendAndClose(&kitchensinkv1.IngestSummary{Received: count})
		}
		if err != nil {
			return err
		}
		count++
		slog.Info("absorb record", "n", count)
	}
}

func (annotationsServer) Echo(
	_ context.Context,
	req *kitchensinkv1.EchoRequest,
) (*kitchensinkv1.EchoRequest, error) {
	slog.Info("echo")
	return req, nil
}

func (annotationsServer) Knobs(
	_ context.Context,
	req *kitchensinkv1.KnobsRequest,
) (*kitchensinkv1.KnobsRequest, error) {
	slog.Info("knobs")
	return req, nil
}

func (annotationsServer) Curated(
	_ context.Context,
	req *kitchensinkv1.CuratedRequest,
) (*kitchensinkv1.CuratedRequest, error) {
	slog.Info("curated")
	return req, nil
}

func (annotationsServer) Flat(
	_ context.Context,
	req *kitchensinkv1.FlatRequest,
) (*kitchensinkv1.FlatRequest, error) {
	slog.Info("flat")
	return req, nil
}
