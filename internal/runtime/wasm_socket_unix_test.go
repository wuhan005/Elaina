//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stealthrocket/wasi-go"
	preview1 "github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	wasmunix "github.com/stealthrocket/wasi-go/systems/unix"
	"github.com/stealthrocket/wazergo"
	"github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"golang.org/x/sys/unix"
)

type wasmReceiveSystem struct{ wasi.System }

func TestWASIPollOneOffLength(t *testing.T) {
	host := preview1.NewHostModule()
	// Use a stub so this test can detect forwarding without constructing any guest slices.
	poll := host.Functions()["poll_oneoff"].WithFunc(func(_ *preview1.Module, _ context.Context, _ api.Module, stack []uint64) {
		stack[0] = uint64(wasi.ESUCCESS)
	})
	poll = wasmWASIABI(host.Name(), poll)

	const maxCount = min(math.MaxInt, math.MaxUint32) / 48 // WASI subscription size in bytes.
	for _, test := range []struct {
		count uint64
		want  wasi.Errno
	}{
		{1, wasi.ESUCCESS},
		{maxCount, wasi.ESUCCESS},
		{maxCount + 1, wasi.EFAULT},
	} {
		stack := []uint64{64, 0, test.count, 0}
		poll.Func(nil, context.Background(), nil, stack)
		if got := wasi.Errno(stack[0]); got != test.want {
			t.Errorf("poll_oneoff with count %d: got %v, want %v", test.count, got, test.want)
		}
	}
}

func TestWASMDNSResultLimit(t *testing.T) {
	for _, count := range []uint64{16, 256, 257, 0xffffffff} {
		called := false
		fn := wasmWASIABI("", wazergo.Function[*preview1.Module]{
			Name: "sock_getaddrinfo",
			Func: func(*preview1.Module, context.Context, api.Module, []uint64) { called = true },
		})

		stack := make([]uint64, 8)
		stack[6] = count
		fn.Func(nil, context.Background(), nil, stack)

		if count > 256 {
			if called || stack[0] != uint64(wasi.EINVAL) {
				t.Fatalf("unbounded DNS allocation: count=%d", count)
			}
		} else if !called {
			t.Fatalf("bounded DNS call rejected: count=%d", count)
		}
	}
}

func TestWASMReadOnlyToolchain(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.a"), []byte("archive"), 0600); err != nil {
		t.Fatal(err)
	}

	fd, err := unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		t.Fatal(err)
	}

	host := &wasmunix.System{}
	defer func() { _ = host.Close(ctx) }()
	root := host.Preopen(wasmunix.FD(fd), "/toolchain", wasi.FDStat{
		RightsBase:       wasmReadOnlyRights & wasi.DirectoryRights,
		RightsInheriting: wasmReadOnlyRights,
	})

	s := &wasmSystem{
		System: host,
	}
	file, errno := s.PathOpen(ctx, root, 0, "package.a", 0, wasi.FDReadRight|wasi.FDFileStatGetRight, wasi.AllRights, 0)
	if errno != wasi.ESUCCESS {
		t.Fatalf("read-only open with Go's inheriting rights: %v", errno)
	}

	buffer := make([]byte, 7)
	if _, errno := s.FDRead(ctx, file, []wasi.IOVec{buffer}); errno != wasi.ESUCCESS || string(buffer) != "archive" {
		t.Fatalf("read-only archive: %q %v", buffer, errno)
	}

	for _, open := range []wasi.OpenFlags{0, wasi.OpenCreate, wasi.OpenTruncate} {
		if _, errno := s.PathOpen(ctx, root, 0, "package.a", open, wasi.FDWriteRight, 0, 0); errno != wasi.ENOTCAPABLE {
			t.Fatalf("writable open: %v", errno)
		}
	}

	for _, errno := range []wasi.Errno{
		s.PathCreateDirectory(ctx, root, "new"), s.PathUnlinkFile(ctx, root, "package.a"),
		s.PathRename(ctx, root, "package.a", root, "changed.a"),
		s.PathLink(ctx, root, 0, "package.a", root, "linked.a"),
		s.PathFileStatSetTimes(ctx, root, 0, "package.a", 0, 0, 0),
		s.FDFileStatSetSize(ctx, file, 0), s.FDFileStatSetTimes(ctx, file, 0, 0, 0),
	} {
		if errno != wasi.ENOTCAPABLE {
			t.Fatalf("toolchain mutation: %v", errno)
		}
	}
}

func TestWASMListenDenied(t *testing.T) {
	ctx := context.Background()
	host := &wasmunix.System{}
	defer func() { _ = host.Close(ctx) }()
	s := &wasmSystem{
		System:   host,
		internet: true,
	}
	open := func(kind wasi.SocketType) wasi.FD {
		t.Helper()
		fd, errno := s.SockOpen(ctx, wasi.InetFamily, kind, wasi.IPProtocol, wasi.SockListenRights|wasi.SockConnectionRights, wasi.SockConnectionRights)
		if errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		return fd
	}

	stream := open(wasi.StreamSocket)
	for _, addr := range []wasi.SocketAddress{&wasi.Inet4Address{}, &wasi.Inet4Address{Addr: [4]byte{127, 0, 0, 1}, Port: 9995}, &wasi.Inet6Address{}} {
		if _, errno := s.SockBind(ctx, stream, addr); errno != wasi.EPERM {
			t.Fatalf("TCP bind %v: %v", addr, errno)
		}
	}
	if errno := s.SockListen(ctx, stream, 8); errno != wasi.EPERM {
		t.Fatalf("listen: %v", errno)
	}
	if _, _, _, errno := s.SockAccept(ctx, stream, 0); errno != wasi.EPERM {
		t.Fatalf("accept: %v", errno)
	}

	datagram := open(wasi.DatagramSocket)
	for _, addr := range []wasi.SocketAddress{&wasi.Inet4Address{Port: 9995}, &wasi.Inet6Address{Port: 53}, &wasi.UnixAddress{Name: "socket"}} {
		if _, errno := s.SockBind(ctx, datagram, addr); errno != wasi.EPERM {
			t.Fatalf("UDP bind %v: %v", addr, errno)
		}
	}
	local, errno := s.SockBind(ctx, datagram, &wasi.Inet4Address{Addr: [4]byte{127, 0, 0, 1}})
	if errno != wasi.ESUCCESS || local.(*wasi.Inet4Address).Port == 0 {
		t.Fatalf("UDP ephemeral bind: %v %v", local, errno)
	}

	unbound := open(wasi.DatagramSocket)
	s.internet = false
	if _, errno := s.SockBind(ctx, unbound, &wasi.Inet4Address{}); errno != wasi.EPERM {
		t.Fatalf("InternetAccess=false allowed bind: %v", errno)
	}
}

func (*wasmReceiveSystem) SockRecv(context.Context, wasi.FD, []wasi.IOVec, wasi.RIFlags) (wasi.Size, wasi.ROFlags, wasi.Errno) {
	return 122, 0, wasi.ESUCCESS
}

func TestWASMStdioIsolation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	host := &wasmunix.System{}
	defer func() { _ = host.Close(ctx) }()

	output := &wasmOutputBuffer{
		limit:  wasmOutputLimit,
		cancel: cancel,
	}
	s := &wasmSystem{
		System:  host,
		outputs: map[wasi.FD]*wasmOutputBuffer{1: output},
	}
	for range 3 {
		fd, err := unix.Open(os.DevNull, unix.O_RDWR, 0)
		if err != nil {
			t.Fatal(err)
		}
		host.Preopen(wasmunix.FD(fd), "stdio", wasi.FDStat{
			RightsBase: wasi.FileRights,
		})
	}

	if n, errno := s.FDWrite(ctx, 1, []wasi.IOVec{[]byte("captured")}); errno != wasi.ESUCCESS || n != 8 || output.buffer.String() != "captured" {
		t.Fatalf("stdout: n=%d errno=%v output=%q", n, errno, output.buffer.String())
	}

	if errno := s.FDStatSetRights(ctx, 1, wasi.FileRights&^wasi.FDWriteRight, 0); errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}
	if _, errno := s.FDWrite(ctx, 1, []wasi.IOVec{[]byte("denied")}); errno != wasi.ENOTCAPABLE {
		t.Fatal("stdout bypassed reduced rights")
	}

	if errno := s.FDClose(ctx, 1); errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}
	path := filepath.Join(t.TempDir(), "reused-fd")
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	guest := host.Register(wasmunix.FD(fd), wasi.FDStat{
		RightsBase: wasi.FileRights,
	})
	if guest != 1 {
		t.Fatalf("expected stdout descriptor reuse, got %d", guest)
	}
	if _, errno := s.FDWrite(ctx, guest, []wasi.IOVec{[]byte("file-only")}); errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}

	data, err := os.ReadFile(path)
	if err != nil || string(data) != "file-only" || output.buffer.String() != "captured" {
		t.Fatalf("descriptor reuse leaked into stdout: file=%q output=%q err=%v", data, output.buffer.String(), err)
	}
}

func TestWASMSocketReceiveABI(t *testing.T) {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer func() { _ = r.Close(ctx) }()

	// This module only exports one page of memory for ABI result pointers.
	m, err := r.Instantiate(ctx, []byte{
		0, 97, 115, 109, 1, 0, 0, 0,
		5, 3, 1, 0, 1,
		7, 10, 1, 6, 109, 101, 109, 111, 114, 121, 2, 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	count := types.Ptr[types.Int32](m.Memory(), 4)
	flags := types.Ptr[types.Uint16](m.Memory(), 2)
	buffers := types.MakeList(types.Ptr[wasi.IOVec](m.Memory(), 16), 0)
	errno := wasmSocketReceive(&preview1.Module{
		WASI: &wasmReceiveSystem{},
	}, ctx, 4, buffers, 0, count, flags)
	if errno != 0 || count.Load() != 122 || flags.Load() != 0 {
		t.Fatalf("sock_recv overwrote the adjacent byte count: errno=%d count=%d flags=%d", errno, count.Load(), flags.Load())
	}
}
