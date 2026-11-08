// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"strings"

	"github.com/stealthrocket/wasi-go"
)

const wasmReadOnlyRights = wasi.FDReadRight | wasi.FDSeekRight | wasi.FDTellRight | wasi.FDStatSetFlagsRight |
	wasi.FDAdviseRight | wasi.FDReadDirRight | wasi.FDFileStatGetRight | wasi.PollFDReadWriteRight |
	wasi.PathOpenRight | wasi.PathReadLinkRight | wasi.PathFileStatGetRight

// Workspaces contain no symlinks; compiler stages also mount a trusted, read-only toolchain.
// Reject symlink creation and parent traversal before delegating to the Unix host.
type wasmSystem struct {
	wasi.System
	internet bool
	outputs  map[wasi.FD]*wasmOutputBuffer
	preopens map[wasi.FD]string
}

func (s *wasmSystem) FDWrite(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec) (wasi.Size, wasi.Errno) {
	output, ok := s.outputs[fd]
	if !ok {
		return s.System.FDWrite(ctx, fd, buffers)
	}

	stat, errno := s.FDStatGet(ctx, fd)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	if !stat.RightsBase.Has(wasi.FDWriteRight) {
		return 0, wasi.ENOTCAPABLE
	}

	var count wasi.Size
	for _, buffer := range buffers {
		n, err := output.Write(buffer)
		count += wasi.Size(n)
		if err != nil {
			return count, wasi.EIO
		}
	}

	return count, wasi.ESUCCESS
}

func (s *wasmSystem) FDClose(ctx context.Context, fd wasi.FD) wasi.Errno {
	errno := s.System.FDClose(ctx, fd)
	if errno == wasi.ESUCCESS {
		delete(s.outputs, fd)
	}

	return errno
}

func validWASMPath(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") || strings.ContainsRune(path, 0) {
		return false
	}

	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return false
		}
	}

	return true
}

func (s *wasmSystem) FDPreStatGet(ctx context.Context, fd wasi.FD) (wasi.PreStat, wasi.Errno) {
	stat, errno := s.System.FDPreStatGet(ctx, fd)
	if name, ok := s.preopens[fd]; ok && errno == wasi.ESUCCESS {
		stat.PreStatDir.NameLength = wasi.Size(len(name))
	}

	return stat, errno
}

func (s *wasmSystem) FDPreStatDirName(ctx context.Context, fd wasi.FD) (string, wasi.Errno) {
	name, errno := s.System.FDPreStatDirName(ctx, fd)
	if path, ok := s.preopens[fd]; ok && errno == wasi.ESUCCESS {
		name = path
	}

	return name, errno
}

func (s *wasmSystem) SockOpen(ctx context.Context, family wasi.ProtocolFamily, kind wasi.SocketType, protocol wasi.Protocol, base, inheriting wasi.Rights) (wasi.FD, wasi.Errno) {
	if !s.internet || (family != wasi.InetFamily && family != wasi.Inet6Family) {
		return -1, wasi.EPERM
	}

	return s.System.SockOpen(ctx, family, kind, protocol, base, inheriting)
}

func (s *wasmSystem) SockAddressInfo(ctx context.Context, name, service string, hints wasi.AddressInfo, results []wasi.AddressInfo) (int, wasi.Errno) {
	if !s.internet {
		return 0, wasi.EPERM
	}

	return s.System.SockAddressInfo(ctx, name, service, hints, results)
}

func (s *wasmSystem) PathSymlink(context.Context, string, wasi.FD, string) wasi.Errno {
	return wasi.ENOTSUP
}

func (s *wasmSystem) PathReadLink(context.Context, wasi.FD, string, []byte) (int, wasi.Errno) {
	return 0, wasi.ENOTSUP
}

func (s *wasmSystem) PathOpen(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, path string, open wasi.OpenFlags, base, inheriting wasi.Rights, fdFlags wasi.FDFlags) (wasi.FD, wasi.Errno) {
	if !validWASMPath(path) {
		return -1, wasi.EPERM
	}

	stat, errno := s.FDStatGet(ctx, fd)
	if errno != wasi.ESUCCESS {
		return -1, errno
	}
	if !stat.RightsInheriting.Has(wasi.FDWriteRight) && base&wasi.WriteRights == 0 && open&(wasi.OpenCreate|wasi.OpenTruncate) == 0 {
		// Go requests writable inheriting rights even when opening a file read-only.
		base &= stat.RightsInheriting
		inheriting &= stat.RightsInheriting
	}

	return s.System.PathOpen(ctx, fd, flags, path, open, base, inheriting, fdFlags)
}

func (s *wasmSystem) PathCreateDirectory(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	if !validWASMPath(path) {
		return wasi.EPERM
	}

	return s.System.PathCreateDirectory(ctx, fd, path)
}

func (s *wasmSystem) PathFileStatGet(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, path string) (wasi.FileStat, wasi.Errno) {
	if !validWASMPath(path) {
		return wasi.FileStat{}, wasi.EPERM
	}

	return s.System.PathFileStatGet(ctx, fd, flags, path)
}

func (s *wasmSystem) PathFileStatSetTimes(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, path string, access, modify wasi.Timestamp, times wasi.FSTFlags) wasi.Errno {
	if !validWASMPath(path) {
		return wasi.EPERM
	}

	return s.System.PathFileStatSetTimes(ctx, fd, flags, path, access, modify, times)
}

func (s *wasmSystem) PathLink(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	if !validWASMPath(oldPath) || !validWASMPath(newPath) {
		return wasi.EPERM
	}

	return s.System.PathLink(ctx, fd, flags, oldPath, newFD, newPath)
}

func (s *wasmSystem) PathRename(ctx context.Context, fd wasi.FD, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	if !validWASMPath(oldPath) || !validWASMPath(newPath) {
		return wasi.EPERM
	}

	return s.System.PathRename(ctx, fd, oldPath, newFD, newPath)
}

func (s *wasmSystem) PathRemoveDirectory(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	if !validWASMPath(path) {
		return wasi.EPERM
	}

	return s.System.PathRemoveDirectory(ctx, fd, path)
}

func (s *wasmSystem) PathUnlinkFile(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	if !validWASMPath(path) {
		return wasi.EPERM
	}

	return s.System.PathUnlinkFile(ctx, fd, path)
}
