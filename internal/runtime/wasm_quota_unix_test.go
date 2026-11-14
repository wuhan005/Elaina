//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stealthrocket/wasi-go"
	wasmunix "github.com/stealthrocket/wasi-go/systems/unix"
	"golang.org/x/sys/unix"

	"github.com/wuhan005/Elaina/internal/config"
)

func wasmQuotaTestSystem(t *testing.T, workspace, file int64, files int) (*wasmSystem, *wasmQuotaSystem, wasi.FD) {
	t.Helper()
	previous := config.App
	t.Cleanup(func() { config.App = previous })
	config.App.WASMLimits.MaxWorkspaceBytes = workspace
	config.App.WASMLimits.MaxFileBytes = file
	config.App.WASMLimits.MaxFiles = files
	dir := t.TempDir()
	quota, err := newWASMQuotaSystem(dir)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		t.Fatal(err)
	}
	host := &wasmunix.System{}
	t.Cleanup(func() { _ = host.Close(context.Background()) })
	root := host.Preopen(wasmunix.FD(fd), "/", wasi.FDStat{
		FileType:         wasi.DirectoryType,
		RightsBase:       wasi.DirectoryRights,
		RightsInheriting: wasi.AllRights,
	})
	quota.host, quota.System = host, host
	return &wasmSystem{
		System: quota,
	}, quota, root
}

func wasmQuotaTestOpen(t *testing.T, s *wasmSystem, root wasi.FD, name string) wasi.FD {
	t.Helper()
	fd, errno := s.PathOpen(context.Background(), root, 0, name, wasi.OpenCreate, wasi.FileRights, 0, 0)
	if errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}
	return fd
}

func wasmQuotaTestWrite(t *testing.T, s *wasmSystem, fd wasi.FD, data string, want wasi.Errno) {
	t.Helper()
	n, errno := s.FDWrite(context.Background(), fd, []wasi.IOVec{[]byte(data)})
	if errno != want || (want == wasi.ESUCCESS && int(n) != len(data)) || (want != wasi.ESUCCESS && n != 0) {
		t.Fatalf("write %q: n=%d, errno=%v, want=%v", data, n, errno, want)
	}
}

func TestWASMFileQuotaWrites(t *testing.T) {
	ctx := context.Background()
	for _, test := range []struct {
		name string
		run  func(*wasmSystem, wasi.FD) wasi.Errno
	}{
		{"write", func(s *wasmSystem, fd wasi.FD) wasi.Errno {
			_, e := s.FDWrite(ctx, fd, []wasi.IOVec{[]byte("123456789")})
			return e
		}},
		{"pwrite", func(s *wasmSystem, fd wasi.FD) wasi.Errno {
			_, e := s.FDPwrite(ctx, fd, []wasi.IOVec{[]byte("x")}, 8)
			return e
		}},
		{"pwrite overflow", func(s *wasmSystem, fd wasi.FD) wasi.Errno {
			_, e := s.FDPwrite(ctx, fd, []wasi.IOVec{[]byte("x")}, ^wasi.FileSize(0))
			return e
		}},
		{"truncate", func(s *wasmSystem, fd wasi.FD) wasi.Errno { return s.FDFileStatSetSize(ctx, fd, 9) }},
		{"allocate", func(s *wasmSystem, fd wasi.FD) wasi.Errno { return s.FDAllocate(ctx, fd, 0, 9) }},
		{"allocate overflow", func(s *wasmSystem, fd wasi.FD) wasi.Errno { return s.FDAllocate(ctx, fd, ^wasi.FileSize(0), 2) }},
		{"sparse write", func(s *wasmSystem, fd wasi.FD) wasi.Errno {
			_, _ = s.FDSeek(ctx, fd, 1<<30, wasi.SeekStart)
			_, e := s.FDWrite(ctx, fd, []wasi.IOVec{[]byte("x")})
			return e
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, quota, root := wasmQuotaTestSystem(t, 12, 8, 4)
			fd := wasmQuotaTestOpen(t, s, root, "file")
			if errno := test.run(s, fd); errno != wasi.EDQUOT {
				t.Fatalf("quota bypass: %v", errno)
			}
			stat, errno := s.FDFileStatGet(ctx, fd)
			if errno != wasi.ESUCCESS || stat.Size != 0 || quota.bytes != 0 {
				t.Fatalf("rejected operation modified file: %+v, %v", stat, errno)
			}
		})
	}

	t.Run("total and reclaimed bytes", func(t *testing.T) {
		s, quota, root := wasmQuotaTestSystem(t, 12, 8, 4)
		a := wasmQuotaTestOpen(t, s, root, "a")
		b := wasmQuotaTestOpen(t, s, root, "b")
		wasmQuotaTestWrite(t, s, a, "12345678", wasi.ESUCCESS)
		wasmQuotaTestWrite(t, s, b, "12345", wasi.EDQUOT)
		wasmQuotaTestWrite(t, s, b, "1234", wasi.ESUCCESS)
		if errno := s.FDFileStatSetSize(ctx, a, 4); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		wasmQuotaTestWrite(t, s, b, "5678", wasi.ESUCCESS)
		if quota.bytes != 12 {
			t.Fatalf("incorrect usage: %d", quota.bytes)
		}
		if _, err := newWASMQuotaSystem(quota.dir); err != nil {
			t.Fatalf("next compiler stage rejected quota boundary: %v", err)
		}
		if err := writeWASMWorkspaceFile(quota.dir, "helper", []byte("x")); err == nil {
			t.Fatal("host helper bypassed total quota")
		}
	})

	t.Run("positional write and overwrite", func(t *testing.T) {
		s, quota, root := wasmQuotaTestSystem(t, 8, 8, 4)
		fd := wasmQuotaTestOpen(t, s, root, "a")
		if n, errno := s.FDPwrite(ctx, fd, []wasi.IOVec{[]byte("1234")}, 4); errno != wasi.ESUCCESS || n != 4 || quota.bytes != 8 {
			t.Fatalf("positional write at boundary: %d, %v", n, errno)
		}
		wasmQuotaTestWrite(t, s, fd, "5678", wasi.ESUCCESS)
		if quota.bytes != 8 {
			t.Fatal("overwrite was charged as growth")
		}
	})

	t.Run("append and reduced rights", func(t *testing.T) {
		s, _, root := wasmQuotaTestSystem(t, 16, 8, 4)
		a := wasmQuotaTestOpen(t, s, root, "a")
		b := wasmQuotaTestOpen(t, s, root, "a")
		wasmQuotaTestWrite(t, s, a, "12345678", wasi.ESUCCESS)
		if errno := s.FDStatSetFlags(ctx, b, wasi.Append); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		wasmQuotaTestWrite(t, s, b, "x", wasi.EDQUOT)
		if _, errno := s.FDPwrite(ctx, b, []wasi.IOVec{[]byte("x")}, 0); errno != wasi.EDQUOT {
			t.Fatalf("pwrite with append bypassed quota: %v", errno)
		}
		if errno := s.FDStatSetRights(ctx, b, wasi.FDWriteRight, 0); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		wasmQuotaTestWrite(t, s, b, "x", wasi.EDQUOT)
		if _, errno := s.PathOpen(ctx, root, 0, "a", wasi.OpenTruncate, wasi.FDWriteRight, 0, 0); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		wasmQuotaTestWrite(t, s, b, "12345678", wasi.ESUCCESS)
	})
}

func TestWASMFileQuotaNamespace(t *testing.T) {
	ctx := context.Background()
	t.Run("hard links and directory count", func(t *testing.T) {
		s, quota, root := wasmQuotaTestSystem(t, 8, 8, 3)
		fd := wasmQuotaTestOpen(t, s, root, "a")
		wasmQuotaTestWrite(t, s, fd, "12345678", wasi.ESUCCESS)
		if errno := s.PathLink(ctx, root, 0, "a", root, "b"); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		if quota.bytes != 8 || quota.files != 2 {
			t.Fatal("hard links must share bytes but consume separate entries")
		}
		if errno := s.PathCreateDirectory(ctx, root, "dir"); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		if errno := s.PathCreateDirectory(ctx, root, "dir/child"); errno != wasi.EDQUOT {
			t.Fatalf("directory quota bypass: %v", errno)
		}
		if _, errno := s.PathOpen(ctx, root, 0, "dir/file", wasi.OpenCreate, wasi.FileRights, 0, 0); errno != wasi.EDQUOT {
			t.Fatalf("file count quota bypass: %v", errno)
		}
		if errno := s.PathLink(ctx, root, 0, "a", root, "c"); errno != wasi.EDQUOT {
			t.Fatalf("hard link quota bypass: %v", errno)
		}
		if errno := s.PathRemoveDirectory(ctx, root, "dir"); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		if errno := s.PathRename(ctx, root, "b", root, "c"); errno != wasi.ESUCCESS || quota.files != 2 {
			t.Fatalf("rename changed entry charge: %v", errno)
		}
	})

	t.Run("open unlinked file", func(t *testing.T) {
		s, quota, root := wasmQuotaTestSystem(t, 12, 8, 2)
		a := wasmQuotaTestOpen(t, s, root, "a")
		wasmQuotaTestWrite(t, s, a, "12345678", wasi.ESUCCESS)
		if errno := s.PathUnlinkFile(ctx, root, "a"); errno != wasi.ESUCCESS || quota.bytes != 8 || quota.files != 1 {
			t.Fatalf("unlink released a live file: %v", errno)
		}
		b := wasmQuotaTestOpen(t, s, root, "b")
		wasmQuotaTestWrite(t, s, b, "12345", wasi.EDQUOT)
		if _, errno := s.PathOpen(ctx, root, 0, "c", wasi.OpenCreate, wasi.FileRights, 0, 0); errno != wasi.EDQUOT {
			t.Fatalf("unlinked file count was released: %v", errno)
		}
		if errno := s.FDRenumber(ctx, a, 255); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		if errno := s.FDRenumber(ctx, 255, 1<<30); errno != wasi.EBADF {
			t.Fatalf("unbounded descriptor table: %v", errno)
		}
		wasmQuotaTestWrite(t, s, 255, "x", wasi.EDQUOT)
		if errno := s.FDClose(ctx, 255); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		wasmQuotaTestWrite(t, s, b, "12345678", wasi.ESUCCESS)
	})

	t.Run("rename over open file", func(t *testing.T) {
		s, quota, root := wasmQuotaTestSystem(t, 12, 8, 3)
		a := wasmQuotaTestOpen(t, s, root, "a")
		b := wasmQuotaTestOpen(t, s, root, "b")
		wasmQuotaTestWrite(t, s, a, "1234", wasi.ESUCCESS)
		wasmQuotaTestWrite(t, s, b, "12345678", wasi.ESUCCESS)
		if errno := s.PathRename(ctx, root, "a", root, "b"); errno != wasi.ESUCCESS || quota.bytes != 12 {
			t.Fatalf("rename released live destination: %v", errno)
		}
		wasmQuotaTestWrite(t, s, a, "x", wasi.EDQUOT)
		if errno := s.FDClose(ctx, b); errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}
		wasmQuotaTestWrite(t, s, a, "5678", wasi.ESUCCESS)
	})
}

func TestWASMWorkspaceStageBoundary(t *testing.T) {
	_, quota, _ := wasmQuotaTestSystem(t, 12, 8, 2)
	if err := writeWASMWorkspaceFile(quota.dir, "main.c", []byte("12345678")); err != nil {
		t.Fatal(err)
	}
	if err := writeWASMWorkspaceFile(quota.dir, "helper", []byte("1234")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"extra", "main.c"} {
		if err := writeWASMWorkspaceFile(quota.dir, name, []byte("x")); err == nil {
			t.Fatal("host write bypassed the workspace quota")
		}
	}
	if err := os.WriteFile(filepath.Join(quota.dir, "oversized"), []byte("123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := newWASMQuotaSystem(quota.dir); err == nil {
		t.Fatal("next stage accepted an oversized workspace")
	}
}
