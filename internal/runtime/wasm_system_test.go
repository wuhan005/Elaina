// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"testing"

	"github.com/stealthrocket/wasi-go"
)

func TestWASMCapabilities(t *testing.T) {
	ctx := context.Background()
	// A nil delegate makes any accidental forwarding fail the test.
	s := &wasmSystem{}
	for _, path := range []string{"/etc/passwd", "..", "../secret", "dir/../../secret", "dir/../file", "file\x00suffix", ""} {
		if _, errno := s.PathOpen(ctx, 3, 0, path, 0, 0, 0, 0); errno != wasi.EPERM {
			t.Fatalf("PathOpen(%q): %v", path, errno)
		}
		if _, errno := s.PathFileStatGet(ctx, 3, 0, path); errno != wasi.EPERM {
			t.Fatalf("PathFileStatGet(%q): %v", path, errno)
		}
		for _, errno := range []wasi.Errno{
			s.PathCreateDirectory(ctx, 3, path), s.PathRemoveDirectory(ctx, 3, path),
			s.PathUnlinkFile(ctx, 3, path), s.PathFileStatSetTimes(ctx, 3, 0, path, 0, 0, 0),
			s.PathLink(ctx, 3, 0, path, 3, "safe"), s.PathLink(ctx, 3, 0, "safe", 3, path),
			s.PathRename(ctx, 3, path, 3, "safe"), s.PathRename(ctx, 3, "safe", 3, path),
		} {
			if errno != wasi.EPERM {
				t.Fatalf("path operation accepted %q: %v", path, errno)
			}
		}
	}
	for _, path := range []string{"main.py", ".", "dir/file", "./file"} {
		if !validWASMPath(path) {
			t.Fatalf("rejected workspace path: %q", path)
		}
	}
	if s.PathSymlink(ctx, "/etc", 3, "link") != wasi.ENOTSUP {
		t.Fatal("symlinks must not expose host files")
	}
	if _, errno := s.SockOpen(ctx, wasi.InetFamily, wasi.StreamSocket, wasi.TCPProtocol, 0, 0); errno != wasi.EPERM {
		t.Fatal("InternetAccess=false allowed a socket")
	}
	if _, errno := s.SockAddressInfo(ctx, "localhost", "80", wasi.AddressInfo{}, nil); errno != wasi.EPERM {
		t.Fatal("InternetAccess=false allowed DNS")
	}
	s.internet = true
	if _, errno := s.SockOpen(ctx, wasi.UnixFamily, wasi.StreamSocket, wasi.IPProtocol, 0, 0); errno != wasi.EPERM {
		t.Fatal("Unix sockets must not expose host services")
	}
}
