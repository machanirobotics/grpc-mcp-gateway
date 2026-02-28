package runtime

import (
	"context"
	"fmt"

	grpcmd "google.golang.org/grpc/metadata"
)

// InProcessServerStream is a generic, channel-backed implementation of
// grpc.ServerStream + Send(T). Generated non-blocking streaming handlers use
// this to call a gRPC server method directly in-process without a network hop.
//
// Usage:
//
//	stream := runtime.NewInProcessServerStream[*MyChunkType](ctx)
//	go func() {
//	    defer stream.Close()
//	    _ = srv.MyStreamingRPC(req, stream)
//	}()
//	for {
//	    chunk, ok := stream.Recv()
//	    if !ok { break }
//	    // process chunk
//	}
type InProcessServerStream[T any] struct {
	ctx context.Context
	ch  chan T
}

// NewInProcessServerStream creates a new InProcessServerStream with a buffered
// channel (capacity 16). The ctx is used by Send to abort early if cancelled.
func NewInProcessServerStream[T any](ctx context.Context) *InProcessServerStream[T] {
	return &InProcessServerStream[T]{ctx: ctx, ch: make(chan T, 16)}
}

// Send enqueues msg into the channel. Blocks if the buffer is full until space
// is available or ctx is cancelled.
func (s *InProcessServerStream[T]) Send(msg T) error {
	select {
	case s.ch <- msg:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// Recv reads the next item from the channel. Blocks until an item arrives or
// the channel is closed (via Close). Returns (zero, false) when closed.
func (s *InProcessServerStream[T]) Recv() (T, bool) {
	chunk, ok := <-s.ch
	return chunk, ok
}

// Close signals that no more items will be sent. The consumer goroutine will
// exit its Recv loop once all buffered items are drained. Must be called
// exactly once, after the producer (gRPC method) returns.
func (s *InProcessServerStream[T]) Close() { close(s.ch) }

// Context returns the stream context. The gRPC server method calls
// stream.Context() to get its context.
func (s *InProcessServerStream[T]) Context() context.Context { return s.ctx }

// The following methods satisfy grpc.ServerStream but are no-ops for in-process
// use since there is no network transport.
func (s *InProcessServerStream[T]) SetHeader(grpcmd.MD) error  { return nil }
func (s *InProcessServerStream[T]) SendHeader(grpcmd.MD) error { return nil }
func (s *InProcessServerStream[T]) SetTrailer(grpcmd.MD)       {}

// SendMsg type-asserts m to T and forwards it to Send. Returns an error if
// the assertion fails so that unsupported usage is caught immediately rather
// than silently dropping the message.
func (s *InProcessServerStream[T]) SendMsg(m any) error {
	if msg, ok := m.(T); ok {
		return s.Send(msg)
	}
	return fmt.Errorf("InProcessServerStream.SendMsg: unexpected type %T", m)
}

// RecvMsg is not supported for a server stream; it returns an error to make
// unsupported usage fail fast instead of silently succeeding.
func (s *InProcessServerStream[T]) RecvMsg(any) error {
	return fmt.Errorf("InProcessServerStream.RecvMsg: not supported")
}
