// Command bookstore-server is a stub gRPC server for the bookstore example.
//
//	go run ./examples/bookstore/server -addr :50053
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
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

func (auctionsServer) WatchAuction(
	req *bookstorev1.WatchAuctionRequest,
	stream bookstorev1.AuctionsService_WatchAuctionServer,
) error {
	g := gofakeit.New(0)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	high := float64(g.Number(50, 200))
	var sent int32
	for {
		if limit := req.GetLimit(); limit > 0 && sent >= limit {
			return nil
		}
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
		}
		high += float64(g.Number(5, 50))
		auction := req.GetAuction()
		if auction == 0 {
			auction = int64(g.Number(1, 5))
		}
		update := &bookstorev1.AuctionUpdate{
			Auction: auction,
			HighBid: high,
			Bidder:  fmt.Sprintf("paddle %d", g.Number(2, 19)),
			State:   bookstorev1.AuctionState_AUCTION_STATE_OPEN,
			At:      timestamppb.Now(),
		}
		if err := stream.Send(update); err != nil {
			return err
		}
		sent++
	}
}

// Bid records the client's bids while rival paddles keep raising the price.
func (auctionsServer) Bid(stream bookstorev1.AuctionsService_BidServer) error {
	g := gofakeit.New(0)

	var mu sync.Mutex
	auction, high, bidder := int64(1), float64(g.Number(20, 80)), "paddle 7"
	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				return
			}
			mu.Lock()
			if req.GetAuction() != 0 {
				auction = req.GetAuction()
			}
			if req.GetAmount() > high {
				high, bidder = req.GetAmount(), "you"
			}
			mu.Unlock()
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
		}
		mu.Lock()
		if bidder != "you" || g.Bool() {
			high += float64(g.Number(5, 50))
			bidder = fmt.Sprintf("paddle %d", g.Number(2, 19))
		}
		update := &bookstorev1.AuctionUpdate{
			Auction: auction,
			HighBid: high,
			Bidder:  bidder,
			State:   bookstorev1.AuctionState_AUCTION_STATE_OPEN,
			At:      timestamppb.Now(),
		}
		mu.Unlock()
		if err := stream.Send(update); err != nil {
			return err
		}
	}
}

func (inventoryServer) ImportBooks(stream bookstorev1.InventoryService_ImportBooksServer) error {
	var created int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&bookstorev1.ImportSummary{Created: created})
		}
		if err != nil {
			return err
		}
		created++
	}
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
