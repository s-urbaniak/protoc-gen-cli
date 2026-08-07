// Command bookstore-server is a stub gRPC server for the bookstore example.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	bookstorev1 "github.com/braveokafor/protoc-gen-cli/examples/bookstore/go-cobra/gen/bookstore/v1"
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

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		slog.Error("listen", "err", err)
		os.Exit(1)
	}
	srv := grpc.NewServer()
	bookstorev1.RegisterBookstoreServiceServer(srv, bookstoreServer{})
	bookstorev1.RegisterAuctionsServiceServer(srv, auctionsServer{})
	bookstorev1.RegisterInventoryServiceServer(srv, inventoryServer{})
	slog.Info("bookstore-server listening", "addr", *addr)
	if err := srv.Serve(lis); err != nil {
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
}

const (
	maxID      = 100
	importCost = time.Millisecond
	reportCost = 2 * time.Second
)

// fake fills msg with fake data.
func fake(msg proto.Message) error {
	return fauxrpc.SetDataOnMessage(msg, fauxrpc.GenOptions{Faker: gofakeit.New(0)})
}

func inStore(id int64) bool { return id >= 1 && id <= maxID }

func fakeBook(g *gofakeit.Faker) *bookstorev1.Book {
	return &bookstorev1.Book{Id: g.Int64(), Title: g.BookTitle(), Author: g.BookAuthor()}
}

func fakeShelf(g *gofakeit.Faker) *bookstorev1.Shelf {
	return &bookstorev1.Shelf{Id: g.Int64(), Theme: g.BookGenre()}
}

func (bookstoreServer) ListShelves(
	context.Context,
	*emptypb.Empty,
) (*bookstorev1.ListShelvesResponse, error) {
	slog.Info("list shelves")
	resp := &bookstorev1.ListShelvesResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	g := gofakeit.New(0)
	for i := range resp.GetShelves() {
		resp.Shelves[i] = fakeShelf(g)
	}
	return resp, nil
}

func (bookstoreServer) CreateShelf(
	_ context.Context,
	req *bookstorev1.CreateShelfRequest,
) (*bookstorev1.Shelf, error) {
	slog.Info("create shelf", "theme", req.GetShelf().GetTheme())
	shelf := req.GetShelf()
	if shelf == nil {
		shelf = &bookstorev1.Shelf{}
	}
	shelf.Id = int64(gofakeit.Number(1, maxID))
	return shelf, nil
}

func (bookstoreServer) GetShelf(
	_ context.Context,
	req *bookstorev1.GetShelfRequest,
) (*bookstorev1.Shelf, error) {
	slog.Info("get shelf", "shelf", req.GetShelf())
	if !inStore(req.GetShelf()) {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	shelf := fakeShelf(gofakeit.New(0))
	shelf.Id = req.GetShelf()
	return shelf, nil
}

func (bookstoreServer) DeleteShelf(
	_ context.Context,
	req *bookstorev1.DeleteShelfRequest,
) (*emptypb.Empty, error) {
	slog.Info("delete shelf", "shelf", req.GetShelf())
	if !inStore(req.GetShelf()) {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	return &emptypb.Empty{}, nil
}

func (bookstoreServer) ListBooks(
	_ context.Context,
	req *bookstorev1.ListBooksRequest,
) (*bookstorev1.ListBooksResponse, error) {
	slog.Info("list books", "shelf", req.GetShelf())
	if !inStore(req.GetShelf()) {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	resp := &bookstorev1.ListBooksResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	g := gofakeit.New(0)
	for i := range resp.GetBooks() {
		resp.Books[i] = fakeBook(g)
	}
	return resp, nil
}

func (bookstoreServer) CreateBook(
	_ context.Context,
	req *bookstorev1.CreateBookRequest,
) (*bookstorev1.Book, error) {
	slog.Info("create book", "shelf", req.GetShelf(), "title", req.GetBook().GetTitle())
	if !inStore(req.GetShelf()) {
		return nil, status.Errorf(codes.NotFound, "shelf %d not found", req.GetShelf())
	}
	book := req.GetBook()
	if book == nil {
		book = &bookstorev1.Book{}
	}
	book.Id = int64(gofakeit.Number(1, maxID))
	return book, nil
}

func (bookstoreServer) GetBook(
	_ context.Context,
	req *bookstorev1.GetBookRequest,
) (*bookstorev1.Book, error) {
	slog.Info("get book", "shelf", req.GetShelf(), "book", req.GetBook())
	if !inStore(req.GetBook()) {
		return nil, status.Errorf(
			codes.NotFound,
			"book %d not found on shelf %d",
			req.GetBook(),
			req.GetShelf(),
		)
	}
	book := fakeBook(gofakeit.New(0))
	book.Id = req.GetBook()
	return book, nil
}

func (bookstoreServer) DeleteBook(
	_ context.Context,
	req *bookstorev1.DeleteBookRequest,
) (*emptypb.Empty, error) {
	slog.Info("delete book", "shelf", req.GetShelf(), "book", req.GetBook())
	if !inStore(req.GetBook()) {
		return nil, status.Errorf(
			codes.NotFound,
			"book %d not found on shelf %d",
			req.GetBook(),
			req.GetShelf(),
		)
	}
	return &emptypb.Empty{}, nil
}

func (auctionsServer) CreateAuction(
	_ context.Context,
	req *bookstorev1.CreateAuctionRequest,
) (*bookstorev1.Auction, error) {
	slog.Info("create auction",
		"title", req.GetLot().GetBook().GetTitle(),
		"reserve", req.GetLot().GetReservePrice())
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
	slog.Info("list auctions", "state", req.GetState())
	resp := &bookstorev1.ListAuctionsResponse{}
	if err := fake(resp); err != nil {
		return nil, err
	}
	g := gofakeit.New(0)
	for _, a := range resp.GetAuctions() {
		a.Lot.Book = fakeBook(g)
		if s := req.GetState(); s != bookstorev1.AuctionState_AUCTION_STATE_UNSPECIFIED {
			a.State = s
		}
	}
	return resp, nil
}

func (auctionsServer) WatchAuction(
	req *bookstorev1.WatchAuctionRequest,
	stream bookstorev1.AuctionsService_WatchAuctionServer,
) error {
	slog.Info("watch auction",
		"auction", req.GetAuction(),
		"author", req.GetAuthor(),
		"limit", req.GetLimit())
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
		slog.Info("auction update", "auction", auction, "high", high, "bidder", update.GetBidder())
		if err := stream.Send(update); err != nil {
			return err
		}
		sent++
	}
}

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
			slog.Info("bid received", "auction", req.GetAuction(), "amount", req.GetAmount())
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
		slog.Info("auction update", "high", update.GetHighBid(), "bidder", update.GetBidder())
		if err := stream.Send(update); err != nil {
			return err
		}
	}
}

func (inventoryServer) ImportBooks(stream bookstorev1.InventoryService_ImportBooksServer) error {
	slog.Info("import books")
	var created int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			slog.Info("import done", "created", created)
			return stream.SendAndClose(&bookstorev1.ImportSummary{Created: created})
		}
		if err != nil {
			return err
		}
		time.Sleep(importCost)
		created++
		slog.Info("import book", "n", created, "title", req.GetBook().GetTitle())
	}
}

func (inventoryServer) ExportReport(
	_ context.Context,
	req *bookstorev1.ExportReportRequest,
) (*bookstorev1.Report, error) {
	slog.Info("export report", "shelf", req.GetShelf(), "filename", req.GetFilename())
	time.Sleep(reportCost)
	return &bookstorev1.Report{
		Filename: req.GetFilename(),
		Books:    int64(gofakeit.Number(1, 100)),
	}, nil
}
