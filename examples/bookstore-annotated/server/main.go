// Command bookstore-annotated-server is a stub gRPC server for the
// bookstore-annotated example.
//
//	go run ./examples/bookstore-annotated/server -addr :50054
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

	bookstoreannotatedv1 "github.com/braveokafor/proto-to-cli/examples/bookstore-annotated/go-cobra/gen/bookstore/annotated/v1"
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
	bookstoreannotatedv1.UnimplementedBookstoreServiceServer
}

type auctionsServer struct {
	bookstoreannotatedv1.UnimplementedAuctionsServiceServer
}

type inventoryServer struct {
	bookstoreannotatedv1.UnimplementedInventoryServiceServer
}

func main() {
	addr := flag.String("addr", ":50054", "listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	bookstoreannotatedv1.RegisterBookstoreServiceServer(srv, bookstoreServer{})
	bookstoreannotatedv1.RegisterAuctionsServiceServer(srv, auctionsServer{})
	bookstoreannotatedv1.RegisterInventoryServiceServer(srv, inventoryServer{})
	log.Printf("bookstore-annotated-server listening on %s", *addr)
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
) (*bookstoreannotatedv1.ListShelvesResponse, error) {
	resp := &bookstoreannotatedv1.ListShelvesResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (bookstoreServer) CreateShelf(
	_ context.Context,
	req *bookstoreannotatedv1.CreateShelfRequest,
) (*bookstoreannotatedv1.Shelf, error) {
	shelf := req.GetShelf()
	if shelf == nil {
		shelf = &bookstoreannotatedv1.Shelf{}
	}
	shelf.SetId(gofakeit.Int64())
	return shelf, nil
}

func (bookstoreServer) GetShelf(
	_ context.Context,
	req *bookstoreannotatedv1.GetShelfRequest,
) (*bookstoreannotatedv1.Shelf, error) {
	if req.GetShelf() < 1 || req.GetShelf() > 100 {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	shelf := &bookstoreannotatedv1.Shelf{}
	if err := fake(shelf); err != nil {
		return nil, err
	}
	shelf.SetId(req.GetShelf())
	return shelf, nil
}

func (bookstoreServer) DeleteShelf(
	context.Context,
	*bookstoreannotatedv1.DeleteShelfRequest,
) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (bookstoreServer) ListBooks(
	_ context.Context,
	req *bookstoreannotatedv1.ListBooksRequest,
) (*bookstoreannotatedv1.ListBooksResponse, error) {
	if req.GetShelf() < 1 || req.GetShelf() > 100 {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	resp := &bookstoreannotatedv1.ListBooksResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (bookstoreServer) CreateBook(
	_ context.Context,
	req *bookstoreannotatedv1.CreateBookRequest,
) (*bookstoreannotatedv1.Book, error) {
	book := req.GetBook()
	if book == nil {
		book = &bookstoreannotatedv1.Book{}
	}
	book.SetId(gofakeit.Int64())
	return book, nil
}

func (bookstoreServer) GetBook(
	_ context.Context,
	req *bookstoreannotatedv1.GetBookRequest,
) (*bookstoreannotatedv1.Book, error) {
	if req.GetBook() < 1 || req.GetBook() > 100 {
		return nil, status.Errorf(
			codes.NotFound,
			"book %d not found on shelf %d",
			req.GetBook(),
			req.GetShelf(),
		)
	}
	book := &bookstoreannotatedv1.Book{}
	if err := fake(book); err != nil {
		return nil, err
	}
	book.SetId(req.GetBook())
	return book, nil
}

func (bookstoreServer) DeleteBook(
	context.Context,
	*bookstoreannotatedv1.DeleteBookRequest,
) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (auctionsServer) CreateAuction(
	_ context.Context,
	req *bookstoreannotatedv1.CreateAuctionRequest,
) (*bookstoreannotatedv1.Auction, error) {
	starts := req.GetStartsAt()
	if starts == nil {
		starts = timestamppb.Now()
	}
	return bookstoreannotatedv1.Auction_builder{
		Id:     gofakeit.Int64(),
		Lot:    req.GetLot(),
		State:  bookstoreannotatedv1.AuctionState_AUCTION_STATE_SCHEDULED,
		EndsAt: timestamppb.New(starts.AsTime().Add(time.Hour)),
	}.Build(), nil
}

func (auctionsServer) ListAuctions(
	_ context.Context,
	req *bookstoreannotatedv1.ListAuctionsRequest,
) (*bookstoreannotatedv1.ListAuctionsResponse, error) {
	resp := &bookstoreannotatedv1.ListAuctionsResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	if s := req.GetState(); s != bookstoreannotatedv1.AuctionState_AUCTION_STATE_UNSPECIFIED {
		for _, a := range resp.GetAuctions() {
			a.SetState(s)
		}
	}
	return resp, nil
}

func (auctionsServer) WatchAuction(
	req *bookstoreannotatedv1.WatchAuctionRequest,
	stream bookstoreannotatedv1.AuctionsService_WatchAuctionServer,
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
		update := bookstoreannotatedv1.AuctionUpdate_builder{
			Auction: auction,
			HighBid: high,
			Bidder:  fmt.Sprintf("paddle %d", g.Number(2, 19)),
			State:   bookstoreannotatedv1.AuctionState_AUCTION_STATE_OPEN,
			At:      timestamppb.Now(),
		}.Build()
		if err := stream.Send(update); err != nil {
			return err
		}
		sent++
	}
}

// Bid records the client's bids while rival paddles increase the price.
func (auctionsServer) Bid(stream bookstoreannotatedv1.AuctionsService_BidServer) error {
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
		update := bookstoreannotatedv1.AuctionUpdate_builder{
			Auction: auction,
			HighBid: high,
			Bidder:  bidder,
			State:   bookstoreannotatedv1.AuctionState_AUCTION_STATE_OPEN,
			At:      timestamppb.Now(),
		}.Build()
		mu.Unlock()
		if err := stream.Send(update); err != nil {
			return err
		}
	}
}

func (inventoryServer) ImportBooks(
	stream bookstoreannotatedv1.InventoryService_ImportBooksServer,
) error {
	var created int32
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(bookstoreannotatedv1.ImportSummary_builder{
				Created: created,
			}.Build())
		}
		if err != nil {
			return err
		}
		created++
	}
}

func (inventoryServer) ExportReport(
	_ context.Context,
	req *bookstoreannotatedv1.ExportReportRequest,
) (*bookstoreannotatedv1.Report, error) {
	return bookstoreannotatedv1.Report_builder{
		Filename: req.GetFilename(),
		Books:    int64(gofakeit.Number(1, 100)),
	}.Build(), nil
}
