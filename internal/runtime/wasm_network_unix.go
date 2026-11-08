//go:build unix

// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/stealthrocket/wasi-go"
	wasmunix "github.com/stealthrocket/wasi-go/systems/unix"
	"golang.org/x/sys/unix"
)

type wasmNetworkSystem struct{ *wasmunix.System }

// Keep native sockets nonblocking and let Go interrupt waits without closing a live WASI file table.
func (s *wasmNetworkSystem) socketCall(ctx context.Context, fd wasi.FD, write bool, call func() wasi.Errno) wasi.Errno {
	native, stat, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	if ctx.Err() != nil {
		return wasi.ECANCELED
	}

	if err := unix.SetNonblock(int(native), true); err != nil {
		return wasi.MakeErrno(err)
	}
	if stat.Flags.Has(wasi.NonBlock) {
		return call()
	}

	duplicate, err := unix.Dup(int(native))
	if err != nil {
		return wasi.MakeErrno(err)
	}
	unix.CloseOnExec(duplicate)
	socket := os.NewFile(uintptr(duplicate), "wasm-socket")
	defer func() { _ = socket.Close() }()
	raw, err := socket.SyscallConn()
	if err != nil {
		return wasi.MakeErrno(err)
	}

	option := wasi.RecvTimeout
	if write {
		option = wasi.SendTimeout
	}
	if value, errno := s.SockGetOpt(ctx, fd, option); errno == wasi.ESUCCESS {
		if timeout, ok := value.(wasi.TimeValue); ok && timeout > 0 {
			if err := socket.SetDeadline(time.Now().Add(time.Duration(timeout))); err != nil {
				return wasi.MakeErrno(err)
			}
		}
	}

	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		_ = socket.SetDeadline(time.Now())
		close(done)
	})
	defer func() {
		if !stop() {
			<-done
		}
	}()

	ready := func(uintptr) bool {
		if ctx.Err() != nil {
			errno = wasi.ECANCELED
			return true
		}

		errno = call()
		return errno != wasi.EAGAIN && errno != wasi.EINPROGRESS
	}
	if write {
		err = raw.Write(ready)
	} else {
		err = raw.Read(ready)
	}

	switch {
	case ctx.Err() != nil:
		return wasi.ECANCELED
	case errors.Is(err, os.ErrDeadlineExceeded):
		return wasi.EAGAIN
	case err != nil:
		return wasi.MakeErrno(err)
	default:
		return errno
	}
}

func (s *wasmNetworkSystem) SockConnect(ctx context.Context, fd wasi.FD, peer wasi.SocketAddress) (address wasi.SocketAddress, errno wasi.Errno) {
	pending := false
	errno = s.socketCall(ctx, fd, true, func() wasi.Errno {
		if !pending {
			address, errno = s.System.SockConnect(ctx, fd, peer)
			pending = errno == wasi.EINPROGRESS
			return errno
		}

		value, errno := s.SockGetOpt(ctx, fd, wasi.QuerySocketError)
		if errno != wasi.ESUCCESS {
			return errno
		}
		if errno = wasi.Errno(value.(wasi.IntValue)); errno != wasi.ESUCCESS {
			return errno
		}

		address, errno = s.SockLocalAddress(ctx, fd)
		return errno
	})

	return address, errno
}

func (s *wasmNetworkSystem) SockAccept(ctx context.Context, fd wasi.FD, flags wasi.FDFlags) (accepted wasi.FD, peer, address wasi.SocketAddress, errno wasi.Errno) {
	accepted = -1
	errno = s.socketCall(ctx, fd, false, func() wasi.Errno {
		accepted, peer, address, errno = s.System.SockAccept(ctx, fd, flags)
		return errno
	})

	return accepted, peer, address, errno
}

func (s *wasmNetworkSystem) SockRecv(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec, flags wasi.RIFlags) (count wasi.Size, resultFlags wasi.ROFlags, errno wasi.Errno) {
	_, stat, _ := s.LookupSocketFD(fd, 0)
	waitAll := flags.Has(wasi.RecvWaitAll) && !flags.Has(wasi.RecvPeek) && !stat.Flags.Has(wasi.NonBlock) && stat.FileType == wasi.SocketStreamType
	remaining := buffers
	if waitAll {
		remaining = append([]wasi.IOVec(nil), buffers...)
	}

	errno = s.socketCall(ctx, fd, false, func() wasi.Errno {
		for {
			n, receivedFlags, errno := s.System.SockRecv(ctx, fd, remaining, flags)
			if errno != wasi.ESUCCESS {
				return errno
			}

			count += n
			resultFlags |= receivedFlags
			if !waitAll || n == 0 {
				return wasi.ESUCCESS
			}

			for len(remaining) > 0 && int(n) >= len(remaining[0]) {
				n -= wasi.Size(len(remaining[0]))
				remaining = remaining[1:]
			}
			if len(remaining) == 0 {
				return wasi.ESUCCESS
			}
			remaining[0] = remaining[0][n:]
			if ctx.Err() != nil {
				return wasi.ECANCELED
			}
		}
	})

	if count > 0 && errno != wasi.ECANCELED {
		errno = wasi.ESUCCESS
	}

	return count, resultFlags, errno
}

func (s *wasmNetworkSystem) SockSend(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec, flags wasi.SIFlags) (count wasi.Size, errno wasi.Errno) {
	errno = s.socketCall(ctx, fd, true, func() wasi.Errno {
		count, errno = s.System.SockSend(ctx, fd, buffers, flags)
		return errno
	})

	return count, errno
}

func (s *wasmNetworkSystem) SockRecvFrom(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec, flags wasi.RIFlags) (count wasi.Size, resultFlags wasi.ROFlags, address wasi.SocketAddress, errno wasi.Errno) {
	errno = s.socketCall(ctx, fd, false, func() wasi.Errno {
		count, resultFlags, address, errno = s.System.SockRecvFrom(ctx, fd, buffers, flags)
		return errno
	})

	return count, resultFlags, address, errno
}

func (s *wasmNetworkSystem) SockSendTo(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec, flags wasi.SIFlags, address wasi.SocketAddress) (count wasi.Size, errno wasi.Errno) {
	errno = s.socketCall(ctx, fd, true, func() wasi.Errno {
		count, errno = s.System.SockSendTo(ctx, fd, buffers, flags, address)
		return errno
	})

	return count, errno
}

func (s *wasmNetworkSystem) FDRead(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec) (wasi.Size, wasi.Errno) {
	if _, _, errno := s.LookupSocketFD(fd, 0); errno == wasi.ESUCCESS {
		count, _, errno := s.SockRecv(ctx, fd, buffers, 0)
		return count, errno
	}

	return s.System.FDRead(ctx, fd, buffers)
}

func (s *wasmNetworkSystem) FDWrite(ctx context.Context, fd wasi.FD, buffers []wasi.IOVec) (wasi.Size, wasi.Errno) {
	if _, _, errno := s.LookupSocketFD(fd, 0); errno == wasi.ESUCCESS {
		return s.SockSend(ctx, fd, buffers, 0)
	}

	return s.System.FDWrite(ctx, fd, buffers)
}
