//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stealthrocket/wasi-go"
	wasmunix "github.com/stealthrocket/wasi-go/systems/unix"

	"github.com/wuhan005/Elaina/internal/config"
)

var errWASMFileQuota = errors.New("WASM workspace file quota exceeded")

type wasmFileID struct {
	device wasi.Device
	inode  wasi.INode
}

type wasmQuotaSystem struct {
	wasi.System
	host   *wasmunix.System
	dir    string
	limits config.WASMLimits
	files  int
	bytes  wasi.FileSize
	sizes  map[wasmFileID]wasi.FileSize
	opened map[wasi.FD]struct{}
}

func newWASMQuotaSystem(dir string) (*wasmQuotaSystem, error) {
	s := &wasmQuotaSystem{
		dir:    dir,
		limits: config.App.WASMLimits,
		opened: make(map[wasi.FD]struct{}),
	}
	if err := s.refresh(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

// Recount after namespace changes; ordinary writes update inode sizes in constant time.
func (s *wasmQuotaSystem) refresh(ctx context.Context) error {
	s.files, s.bytes = 0, 0
	s.sizes = make(map[wasmFileID]wasi.FileSize)
	seen := make(map[wasmFileID]bool)
	add := func(id wasmFileID, regular bool, size wasi.FileSize) error {
		s.files++
		if s.files > s.limits.MaxFiles {
			return errWASMFileQuota
		}
		if regular && !seen[id] {
			if size > wasi.FileSize(s.limits.MaxFileBytes) || size > wasi.FileSize(s.limits.MaxWorkspaceBytes)-s.bytes {
				return errWASMFileQuota
			}
			s.bytes += size
			s.sizes[id] = size
		}
		seen[id] = true
		return nil
	}

	if err := filepath.WalkDir(s.dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || path == s.dir {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() && !info.IsDir() {
			return errors.New("unsupported WASM workspace file type")
		}
		stat := info.Sys().(*syscall.Stat_t)
		return add(wasmFileID{
			device: wasi.Device(stat.Dev),
			inode:  wasi.INode(stat.Ino),
		}, info.Mode().IsRegular(), wasi.FileSize(info.Size()))
	}); err != nil {
		return err
	}

	// Unlink and rename must not free a file's quota while a guest descriptor still holds it.
	for fd := range s.opened {
		file, _, errno := s.host.LookupFD(fd, 0)
		if errno != wasi.ESUCCESS {
			return errno
		}
		stat, errno := file.FDFileStatGet(ctx)
		if errno != wasi.ESUCCESS {
			return errno
		}
		id := wasmFileID{
			device: stat.Device,
			inode:  stat.INode,
		}
		if stat.NLink == 0 && !seen[id] {
			if err := add(id, stat.FileType == wasi.RegularFileType, stat.Size); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *wasmQuotaSystem) reserve(stat wasi.FileStat, size wasi.FileSize) wasi.Errno {
	id := wasmFileID{
		device: stat.Device,
		inode:  stat.INode,
	}
	previous, ok := s.sizes[id]
	if !ok {
		return wasi.ENOTCAPABLE
	}
	if size > wasi.FileSize(s.limits.MaxFileBytes) || size > wasi.FileSize(s.limits.MaxWorkspaceBytes)-(s.bytes-previous) {
		return wasi.EDQUOT
	}
	s.bytes = s.bytes - previous + size
	s.sizes[id] = size
	return wasi.ESUCCESS
}

func (s *wasmQuotaSystem) finish(ctx context.Context, file wasmunix.FD) wasi.Errno {
	stat, errno := file.FDFileStatGet(ctx)
	if errno != wasi.ESUCCESS {
		return errno
	}
	return s.reserve(stat, stat.Size)
}

func (s *wasmQuotaSystem) write(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec, offset wasi.FileSize, positional bool) (wasi.Size, wasi.Errno) {
	file, flags, errno := s.host.LookupFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	stat, errno := file.FDFileStatGet(ctx)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	if stat.FileType == wasi.RegularFileType {
		var length wasi.FileSize
		for _, buffer := range buffers {
			if wasi.FileSize(len(buffer)) > wasi.FileSize(s.limits.MaxFileBytes)-length {
				return 0, wasi.EDQUOT
			}
			length += wasi.FileSize(len(buffer))
		}
		end := offset
		if flags.Flags.Has(wasi.Append) {
			// Linux also appends pwrite on O_APPEND descriptors; cover both platform behaviors.
			end = max(end, stat.Size)
		} else if !positional {
			end, errno = file.FDSeek(ctx, 0, wasi.SeekCurrent)
			if errno != wasi.ESUCCESS {
				return 0, errno
			}
		}
		if length != 0 {
			if end > wasi.FileSize(s.limits.MaxFileBytes)-length {
				return 0, wasi.EDQUOT
			}
			if errno := s.reserve(stat, max(stat.Size, end+length)); errno != wasi.ESUCCESS {
				return 0, errno
			}
		}
	}

	var n wasi.Size
	if positional {
		n, errno = s.System.FDPwrite(ctx, fd, buffers, offset)
	} else {
		n, errno = s.System.FDWrite(ctx, fd, buffers)
	}
	if stat.FileType == wasi.RegularFileType {
		if updateErr := s.finish(ctx, file); updateErr != wasi.ESUCCESS {
			return n, updateErr
		}
	}
	return n, errno
}

func (s *wasmQuotaSystem) FDWrite(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec) (wasi.Size, wasi.Errno) {
	return s.write(ctx, fd, buffers, 0, false)
}

func (s *wasmQuotaSystem) FDPwrite(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec, offset wasi.FileSize) (wasi.Size, wasi.Errno) {
	return s.write(ctx, fd, buffers, offset, true)
}

func (s *wasmQuotaSystem) resize(ctx context.Context, fd wasi.FD, offset, length wasi.FileSize, allocate bool) wasi.Errno {
	right := wasi.FDFileStatSetSizeRight
	if allocate {
		right = wasi.FDAllocateRight
	}
	file, _, errno := s.host.LookupFD(fd, right)
	if errno != wasi.ESUCCESS {
		return errno
	}
	stat, errno := file.FDFileStatGet(ctx)
	if errno != wasi.ESUCCESS {
		return errno
	}
	if stat.FileType != wasi.RegularFileType {
		return wasi.EINVAL
	}
	if length > wasi.FileSize(s.limits.MaxFileBytes) || offset > wasi.FileSize(s.limits.MaxFileBytes)-length {
		return wasi.EDQUOT
	}
	size := length
	if allocate {
		size = max(stat.Size, offset+length)
	}
	// Keep the old charge until a shrinking operation succeeds.
	if errno := s.reserve(stat, max(stat.Size, size)); errno != wasi.ESUCCESS {
		return errno
	}
	if allocate {
		errno = s.System.FDAllocate(ctx, fd, offset, length)
	} else {
		errno = s.System.FDFileStatSetSize(ctx, fd, size)
	}
	if updateErr := s.finish(ctx, file); updateErr != wasi.ESUCCESS {
		return updateErr
	}
	return errno
}

func (s *wasmQuotaSystem) FDFileStatSetSize(ctx context.Context, fd wasi.FD, size wasi.FileSize) wasi.Errno {
	return s.resize(ctx, fd, 0, size, false)
}

func (s *wasmQuotaSystem) FDAllocate(ctx context.Context, fd wasi.FD, offset, length wasi.FileSize) wasi.Errno {
	return s.resize(ctx, fd, offset, length, true)
}

func (s *wasmQuotaSystem) checkCreate(ctx context.Context, fd wasi.FD, path string, right wasi.Rights) wasi.Errno {
	file, _, errno := s.host.LookupFD(fd, right)
	if errno != wasi.ESUCCESS {
		return errno
	}
	_, errno = file.PathFileStatGet(ctx, 0, path)
	if errno == wasi.ENOENT && s.files >= s.limits.MaxFiles {
		return wasi.EDQUOT
	}
	return wasi.ESUCCESS
}

func (s *wasmQuotaSystem) recount(ctx context.Context, errno wasi.Errno) wasi.Errno {
	if errno == wasi.ESUCCESS {
		if err := s.refresh(ctx); err != nil {
			return wasi.EIO
		}
	}
	return errno
}

func (s *wasmQuotaSystem) PathOpen(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, path string, open wasi.OpenFlags, base, inheriting wasi.Rights, fdFlags wasi.FDFlags) (wasi.FD, wasi.Errno) {
	if open.Has(wasi.OpenCreate) {
		if errno := s.checkCreate(ctx, fd, path, wasi.PathCreateFileRight); errno != wasi.ESUCCESS {
			return -1, errno
		}
	}
	opened, errno := s.System.PathOpen(ctx, fd, flags, path, open, base, inheriting, fdFlags)
	if errno == wasi.ESUCCESS {
		s.opened[opened] = struct{}{}
		if open&(wasi.OpenCreate|wasi.OpenTruncate) != 0 {
			errno = s.recount(ctx, errno)
		}
	}
	return opened, errno
}

func (s *wasmQuotaSystem) PathCreateDirectory(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	if errno := s.checkCreate(ctx, fd, path, wasi.PathCreateDirectoryRight); errno != wasi.ESUCCESS {
		return errno
	}
	return s.recount(ctx, s.System.PathCreateDirectory(ctx, fd, path))
}

func (s *wasmQuotaSystem) PathLink(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	if errno := s.checkCreate(ctx, newFD, newPath, wasi.PathLinkTargetRight); errno != wasi.ESUCCESS {
		return errno
	}
	return s.recount(ctx, s.System.PathLink(ctx, fd, flags, oldPath, newFD, newPath))
}

func (s *wasmQuotaSystem) PathRename(ctx context.Context, fd wasi.FD, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	return s.recount(ctx, s.System.PathRename(ctx, fd, oldPath, newFD, newPath))
}

func (s *wasmQuotaSystem) PathUnlinkFile(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	return s.recount(ctx, s.System.PathUnlinkFile(ctx, fd, path))
}

func (s *wasmQuotaSystem) PathRemoveDirectory(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	return s.recount(ctx, s.System.PathRemoveDirectory(ctx, fd, path))
}

func (s *wasmQuotaSystem) FDClose(ctx context.Context, fd wasi.FD) wasi.Errno {
	errno := s.System.FDClose(ctx, fd)
	if errno == wasi.ESUCCESS {
		delete(s.opened, fd)
	}
	return s.recount(ctx, errno)
}

func (s *wasmQuotaSystem) FDRenumber(ctx context.Context, from, to wasi.FD) wasi.Errno {
	errno := s.System.FDRenumber(ctx, from, to)
	if errno == wasi.ESUCCESS {
		_, tracked := s.opened[from]
		delete(s.opened, from)
		delete(s.opened, to)
		if tracked {
			s.opened[to] = struct{}{}
		}
	}
	return s.recount(ctx, errno)
}

func writeWASMWorkspaceFile(dir, name string, data []byte) error {
	quota, err := newWASMQuotaSystem(dir)
	if err != nil {
		return err
	}
	if quota.files >= quota.limits.MaxFiles || int64(len(data)) > quota.limits.MaxFileBytes ||
		wasi.FileSize(len(data)) > wasi.FileSize(quota.limits.MaxWorkspaceBytes)-quota.bytes {
		return errWASMFileQuota
	}
	// Host-created helper files must not overwrite an existing guest file or hard link.
	file, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
