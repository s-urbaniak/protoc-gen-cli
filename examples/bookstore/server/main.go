// Command bookstore-server is a stub gRPC server for the bookstore example.
//
//	go run ./examples/bookstore/server -addr :50053
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	bookstorev1 "github.com/braveokafor/proto-to-cli/examples/bookstore/go-cobra/gen/bookstore/v1"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/sudorandom/fauxrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type bookstoreServer struct {
	bookstorev1.UnimplementedBookstoreServiceServer
}

type auctionsServer struct {
	bookstorev1.UnimplementedAuctionsServiceServer
}

type inventoryServer struct {
	bookstorev1.UnimplementedInventoryServiceServer
}

func main() {
	addr := flag.String("addr", ":50053", "listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	bookstorev1.RegisterBookstoreServiceServer(srv, bookstoreServer{})
	bookstorev1.RegisterAuctionsServiceServer(srv, auctionsServer{})
	bookstorev1.RegisterInventoryServiceServer(srv, inventoryServer{})
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
) (*bookstorev1.ListShelvesResponse, error) {
	resp := &bookstorev1.ListShelvesResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (bookstoreServer) CreateShelf(
	_ context.Context,
	req *bookstorev1.CreateShelfRequest,
) (*bookstorev1.Shelf, error) {
	shelf := req.GetShelf()
	if shelf == nil {
		shelf = &bookstorev1.Shelf{}
	}
	shelf.Id = gofakeit.Int64()
	return shelf, nil
}

func (bookstoreServer) GetShelf(
	_ context.Context,
	req *bookstorev1.GetShelfRequest,
) (*bookstorev1.Shelf, error) {
	if req.GetShelf() < 1 || req.GetShelf() > 100 {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	shelf := &bookstorev1.Shelf{}
	if err := fake(shelf); err != nil {
		return nil, err
	}
	shelf.Id = req.GetShelf()
	return shelf, nil
}

func (bookstoreServer) DeleteShelf(
	context.Context,
	*bookstorev1.DeleteShelfRequest,
) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (bookstoreServer) ListBooks(
	_ context.Context,
	req *bookstorev1.ListBooksRequest,
) (*bookstorev1.ListBooksResponse, error) {
	if req.GetShelf() < 1 || req.GetShelf() > 100 {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	resp := &bookstorev1.ListBooksResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (bookstoreServer) CreateBook(
	_ context.Context,
	req *bookstorev1.CreateBookRequest,
) (*bookstorev1.Book, error) {
	book := req.GetBook()
	if book == nil {
		book = &bookstorev1.Book{}
	}
	book.Id = gofakeit.Int64()
	return book, nil
}

func (bookstoreServer) GetBook(
	_ context.Context,
	req *bookstorev1.GetBookRequest,
) (*bookstorev1.Book, error) {
	if req.GetBook() < 1 || req.GetBook() > 100 {
		return nil, status.Errorf(
			codes.NotFound,
			"book %d not found on shelf %d",
			req.GetBook(),
			req.GetShelf(),
		)
	}
	book := &bookstorev1.Book{}
	if err := fake(book); err != nil {
		return nil, err
	}
	book.Id = req.GetBook()
	return book, nil
}

func (bookstoreServer) DeleteBook(
	context.Context,
	*bookstorev1.DeleteBookRequest,
) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (auctionsServer) CreateAuction(
	_ context.Context,
	req *bookstorev1.CreateAuctionRequest,
) (*bookstorev1.Auction, error) {
	starts := req.GetStartsAt()
	if starts == nil {
		starts = timestamppb.Now()
	}
	return &bookstorev1.Auction{
		Id:     gofakeit.Int64(),
		Lot:    req.GetLot(),
		State:  bookstorev1.AuctionState_AUCTION_STATE_SCHEDULED,
		EndsAt: timestamppb.New(starts.AsTime().Add(time.Hour)),
	}, nil
}

func (auctionsServer) ListAuctions(
	_ context.Context,
	req *bookstorev1.ListAuctionsRequest,
) (*bookstorev1.ListAuctionsResponse, error) {
	resp := &bookstorev1.ListAuctionsResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	if s := req.GetState(); s != bookstorev1.AuctionState_AUCTION_STATE_UNSPECIFIED {
		for _, a := range resp.GetAuctions() {
			a.State = s
		}
	}
	return resp, nil
}

func (inventoryServer) ExportReport(
	_ context.Context,
	req *bookstorev1.ExportReportRequest,
) (*bookstorev1.Report, error) {
	return &bookstorev1.Report{
		Filename: req.GetFilename(),
		Books:    int64(gofakeit.Number(1, 100)),
	}, nil
}
