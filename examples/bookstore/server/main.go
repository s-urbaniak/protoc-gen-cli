// Command bookstore-server is a stub gRPC server for the bookstore example.
//
//	go run ./examples/bookstore/server -addr :50053
package main

import (
	"context"
	"flag"
	"log"
	"net"

	pb "github.com/braveokafor/proto-to-cli/examples/bookstore/go-cobra/gen/bookstore/v1"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type bookstoreServer struct {
	pb.UnimplementedBookstoreServiceServer
}

func main() {
	addr := flag.String("addr", ":50053", "listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterBookstoreServiceServer(srv, bookstoreServer{})
	log.Printf("bookstore-server listening on %s", *addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// fake fills msg with fake data.
func fake(msg proto.Message) error {
	return fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{Faker: gofakeit.New(0)})
}

func (bookstoreServer) ListShelves(
	context.Context,
	*emptypb.Empty,
) (*pb.ListShelvesResponse, error) {
	resp := &pb.ListShelvesResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (bookstoreServer) CreateShelf(
	_ context.Context,
	req *pb.CreateShelfRequest,
) (*pb.Shelf, error) {
	shelf := req.GetShelf()
	if shelf == nil {
		shelf = &pb.Shelf{}
	}
	shelf.Id = gofakeit.Int64()
	return shelf, nil
}

func (bookstoreServer) GetShelf(_ context.Context, req *pb.GetShelfRequest) (*pb.Shelf, error) {
	if req.GetShelf() < 1 || req.GetShelf() > 100 {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	shelf := &pb.Shelf{}
	if err := fake(shelf); err != nil {
		return nil, err
	}
	shelf.Id = req.GetShelf()
	return shelf, nil
}

func (bookstoreServer) DeleteShelf(
	context.Context,
	*pb.DeleteShelfRequest,
) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (bookstoreServer) ListBooks(
	_ context.Context,
	req *pb.ListBooksRequest,
) (*pb.ListBooksResponse, error) {
	if req.GetShelf() < 1 || req.GetShelf() > 100 {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	resp := &pb.ListBooksResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (bookstoreServer) CreateBook(_ context.Context, req *pb.CreateBookRequest) (*pb.Book, error) {
	book := req.GetBook()
	if book == nil {
		book = &pb.Book{}
	}
	book.Id = gofakeit.Int64()
	return book, nil
}

func (bookstoreServer) GetBook(_ context.Context, req *pb.GetBookRequest) (*pb.Book, error) {
	if req.GetBook() < 1 || req.GetBook() > 100 {
		return nil, status.Errorf(
			codes.NotFound,
			"book %d not found on shelf %d",
			req.GetBook(),
			req.GetShelf(),
		)
	}
	book := &pb.Book{}
	if err := fake(book); err != nil {
		return nil, err
	}
	book.Id = req.GetBook()
	return book, nil
}

func (bookstoreServer) DeleteBook(context.Context, *pb.DeleteBookRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
