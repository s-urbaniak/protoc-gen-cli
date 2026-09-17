package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	connectv1 "github.com/braveokafor/protoc-gen-cli/examples/connect/go-cobra/gen/connect/v1"
	connectv1connect "github.com/braveokafor/protoc-gen-cli/examples/connect/go-cobra/gen/connect/v1/connectv1connect"
)

type streamService struct {
	connectv1connect.UnimplementedStreamServiceHandler
	firstResponse chan<- struct{}
	waitForEOF    <-chan struct{}
	early         bool
}

func (streamService) Unary(
	_ context.Context,
	req *connect.Request[connectv1.EchoRequest],
) (*connect.Response[connectv1.EchoResponse], error) {
	return connect.NewResponse(&connectv1.EchoResponse{Text: req.Msg.Text}), nil
}

func (streamService) Server(
	_ context.Context,
	req *connect.Request[connectv1.EchoRequest],
	stream *connect.ServerStream[connectv1.EchoResponse],
) error {
	return stream.Send(&connectv1.EchoResponse{Text: req.Msg.Text})
}

func (streamService) Client(
	_ context.Context,
	stream *connect.ClientStream[connectv1.EchoRequest],
) (*connect.Response[connectv1.EchoResponse], error) {
	var texts []string
	for stream.Receive() {
		texts = append(texts, stream.Msg().Text)
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&connectv1.EchoResponse{Text: strings.Join(texts, ",")}), nil
}

func (s streamService) Bidi(
	_ context.Context,
	stream *connect.BidiStream[connectv1.EchoRequest, connectv1.EchoResponse],
) error {
	req, err := stream.Receive()
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}
	if s.early {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("stop early"))
	}
	if err := stream.Send(&connectv1.EchoResponse{Text: req.Text}); err != nil {
		return err
	}
	if s.firstResponse != nil {
		close(s.firstResponse)
	}
	if s.waitForEOF != nil {
		<-s.waitForEOF
	}
	for {
		req, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&connectv1.EchoResponse{Text: req.Text}); err != nil {
			return err
		}
	}
}

func newStreamClient(t *testing.T, svc streamService) connectv1connect.StreamServiceClient {
	t.Helper()
	_, handler := connectv1connect.NewStreamServiceHandler(svc)
	server := httptest.NewUnstartedServer(handler)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)
	return connectv1connect.NewStreamServiceClient(server.Client(), server.URL)
}

func TestGeneratedConnectCommandsAllShapes(t *testing.T) {
	client := newStreamClient(t, streamService{})
	var _ connectv1.StreamServiceCLIClient = client

	for _, name := range []string{"unary", "server", "client", "bidi"} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			cmd := connectv1.NewStreamServiceCommand(client)
			cmd.SetOut(&out)
			cmd.SetArgs([]string{name, "--text", name})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), name) {
				t.Fatalf("output %q does not contain %q", out.String(), name)
			}
		})
	}
}

func TestGeneratedConnectBidiPrintsBeforeRequestEOF(t *testing.T) {
	allowEOF := make(chan struct{})
	serverResponded := make(chan struct{})
	client := newStreamClient(t, streamService{
		firstResponse: serverResponded,
		waitForEOF:    allowEOF,
	})
	printed := make(chan struct{})
	decoder := connectv1.StreamServiceDecoder{
		Name: "blocking",
		Decode: func(io.Reader) iter.Seq2[[]byte, error] {
			return func(yield func([]byte, error) bool) {
				if !yield([]byte(`{"text":"first"}`), nil) {
					return
				}
				<-allowEOF
			}
		},
	}
	cmd := connectv1.NewStreamServiceCommand(client, connectv1.StreamServiceOptions{
		Decoders: []connectv1.StreamServiceDecoder{decoder},
		Printers: map[string]connectv1.StreamServicePrinter{
			"signal": func(w io.Writer, _ connectv1.StreamServiceView, body []byte) error {
				close(printed)
				_, err := fmt.Fprintln(w, string(body))
				return err
			},
		},
	})
	cmd.SetArgs([]string{"bidi", "-d", "ignored", "-o", "signal"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	done := make(chan error, 1)
	go func() { done <- cmd.Execute() }()

	select {
	case <-serverResponded:
	case <-time.After(time.Second):
		t.Fatal("server did not receive the first request")
	}
	select {
	case <-printed:
	case <-time.After(time.Second):
		t.Fatal("response was not printed before the request source closed")
	}
	close(allowEOF)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("command did not finish after the request source closed")
	}
}

func TestGeneratedConnectBidiKeepsEarlyServerStatus(t *testing.T) {
	client := newStreamClient(t, streamService{early: true})
	cmd := connectv1.NewStreamServiceCommand(client)
	cmd.SetArgs([]string{"bidi", "-d", `{"text":"first"}`, "-d", `{"text":"second"}`})
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	if err == nil || err.Error() != "stop early" {
		t.Fatalf("error = %v, want stop early", err)
	}
	var coded interface{ ExitCode() int }
	if !errors.As(err, &coded) || coded.ExitCode() != 1 {
		t.Fatalf("exit code = %v, want 1", coded)
	}
}

func TestGeneratedConnectBidiCanceledExitCode(t *testing.T) {
	client := newStreamClient(t, streamService{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := connectv1.NewStreamServiceCommand(client)
	cmd.SetArgs([]string{"bidi", "--text", "canceled"})
	cmd.SetErr(io.Discard)
	err := cmd.ExecuteContext(ctx)
	var coded interface{ ExitCode() int }
	if !errors.As(err, &coded) || coded.ExitCode() != 130 {
		t.Fatalf("error %v has exit code %v, want 130", err, coded)
	}
}
