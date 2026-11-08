//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stealthrocket/wasi-go"
	wasmunix "github.com/stealthrocket/wasi-go/systems/unix"
)

func newWASMTestSocket(t *testing.T, kind wasi.SocketType) (*wasmNetworkSystem, wasi.FD) {
	t.Helper()

	s := &wasmNetworkSystem{System: &wasmunix.System{}}
	t.Cleanup(func() { _ = s.Close(context.Background()) })

	fd, errno := s.SockOpen(context.Background(), wasi.InetFamily, kind, wasi.IPProtocol,
		wasi.SockListenRights|wasi.SockConnectionRights, wasi.SockConnectionRights)
	if errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}

	return s, fd
}

func connectWASMTestSocket(t *testing.T) (*wasmNetworkSystem, wasi.FD, net.Conn) {
	t.Helper()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()

	s, fd := newWASMTestSocket(t, wasi.StreamSocket)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, errno := s.SockConnect(ctx, fd, &wasi.Inet4Address{Addr: [4]byte{127, 0, 0, 1}, Port: listener.Addr().(*net.TCPAddr).Port}); errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}

	peer, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = peer.Close() })

	return s, fd, peer
}

func TestWASMNetworkCancellation(t *testing.T) {
	for _, operation := range []string{"receive", "read", "send", "write", "accept", "receive datagram"} {
		t.Run(operation, func(t *testing.T) {
			var s *wasmNetworkSystem
			var fd wasi.FD
			switch operation {
			case "accept", "receive datagram":
				kind := wasi.StreamSocket
				if operation == "receive datagram" {
					kind = wasi.DatagramSocket
				}
				s, fd = newWASMTestSocket(t, kind)
				if _, errno := s.SockBind(context.Background(), fd, &wasi.Inet4Address{Addr: [4]byte{127, 0, 0, 1}}); errno != wasi.ESUCCESS {
					t.Fatal(errno)
				}
				if operation == "accept" {
					if errno := s.SockListen(context.Background(), fd, 1); errno != wasi.ESUCCESS {
						t.Fatal(errno)
					}
				}
			default:
				s, fd, _ = connectWASMTestSocket(t)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			buffers := []wasi.IOVec{make([]byte, 64*1024)}
			var errno wasi.Errno
			start := time.Now()
			switch operation {
			case "receive":
				_, _, errno = s.SockRecv(ctx, fd, buffers, 0)
			case "read":
				_, errno = s.FDRead(ctx, fd, buffers)
			case "send", "write":
				if errno := s.SockSetOpt(ctx, fd, wasi.SendBufferSize, wasi.IntValue(4096)); errno != wasi.ESUCCESS {
					t.Fatal(errno)
				}
				for errno == wasi.ESUCCESS {
					if operation == "send" {
						_, errno = s.SockSend(ctx, fd, buffers, 0)
					} else {
						_, errno = s.FDWrite(ctx, fd, buffers)
					}
				}
			case "accept":
				_, _, _, errno = s.SockAccept(ctx, fd, 0)
			case "receive datagram":
				_, _, _, errno = s.SockRecvFrom(ctx, fd, buffers, 0)
			}

			if errno != wasi.ECANCELED || time.Since(start) > time.Second {
				t.Fatalf("blocked call did not cancel promptly: %v after %s", errno, time.Since(start))
			}
		})
	}
}

func TestWASMNetworkFlagsAndTimeout(t *testing.T) {
	s, fd, _ := connectWASMTestSocket(t)
	ctx := context.Background()
	if errno := s.FDStatSetFlags(ctx, fd, wasi.NonBlock); errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}

	buffers := []wasi.IOVec{make([]byte, 16)}
	if _, _, errno := s.SockRecv(ctx, fd, buffers, 0); errno != wasi.EAGAIN {
		t.Fatalf("nonblocking receive: %v", errno)
	}

	if errno := s.FDStatSetFlags(ctx, fd, 0); errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}
	if errno := s.SockSetOpt(ctx, fd, wasi.RecvTimeout, wasi.TimeValue(50*time.Millisecond)); errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}
	start := time.Now()
	if _, _, errno := s.SockRecv(ctx, fd, buffers, 0); errno != wasi.EAGAIN || time.Since(start) > time.Second {
		t.Fatalf("receive timeout: %v after %s", errno, time.Since(start))
	}

	stat, errno := s.FDStatGet(ctx, fd)
	if errno != wasi.ESUCCESS || stat.Flags.Has(wasi.NonBlock) {
		t.Fatal("host nonblocking mode leaked into guest flags")
	}
}

func TestWASMNetworkWaitAll(t *testing.T) {
	s, fd, peer := connectWASMTestSocket(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = peer.Write([]byte("abc"))
		time.Sleep(10 * time.Millisecond)
		_, _ = peer.Write([]byte("def"))
	}()

	buffers := []wasi.IOVec{make([]byte, 2), make([]byte, 4)}
	n, _, errno := s.SockRecv(ctx, fd, buffers, wasi.RecvWaitAll)
	<-done

	if errno != wasi.ESUCCESS || n != 6 || string(buffers[0])+string(buffers[1]) != "abcdef" {
		t.Fatalf("fragmented receive: n=%d errno=%v data=%q", n, errno, buffers)
	}
}
