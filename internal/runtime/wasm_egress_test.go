// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"net/netip"
	"slices"
	"testing"

	"github.com/stealthrocket/wasi-go"
	"github.com/wuhan005/Elaina/internal/config"
)

func allowWASMTestLoopback(t *testing.T) {
	t.Helper()
	denied := wasmDeniedDestinations
	t.Cleanup(func() { wasmDeniedDestinations = denied })
	wasmDeniedDestinations = slices.DeleteFunc(slices.Clone(denied), func(prefix netip.Prefix) bool {
		return prefix == netip.MustParsePrefix("127.0.0.0/8")
	})
}

func TestWASMDestinationPolicy(t *testing.T) {
	denied := config.App.WASMDeniedCIDRs
	t.Cleanup(func() { config.App.WASMDeniedCIDRs = denied })
	config.App.WASMDeniedCIDRs = []netip.Prefix{netip.MustParsePrefix("8.8.4.0/24"), netip.MustParsePrefix("2606:4700:4700::/48")}

	for _, test := range []struct {
		ip      string
		allowed bool
	}{
		{"0.0.0.0", false}, {"0.1.2.3", false}, {"127.5.0.1", false},
		{"10.0.0.1", false}, {"172.31.255.255", false}, {"192.168.0.1", false},
		{"100.64.0.1", false}, {"100.100.100.200", false}, {"169.254.169.254", false}, {"168.63.129.16", false},
		{"192.0.2.1", false}, {"198.18.0.1", false}, {"224.0.0.1", false}, {"255.255.255.255", false},
		{"::", false}, {"::1", false}, {"fd00:ec2::254", false}, {"fe80::1", false}, {"fec0::1", false}, {"ff02::1", false},
		{"64:ff9b::a9fe:a9fe", false}, {"2001::1", false}, {"2002:7f00:1::1", false}, {"2001:db8::1", false},
		{"::ffff:127.0.0.1", false}, {"::ffff:169.254.169.254", false},
		{"8.8.4.4", false}, {"::ffff:8.8.4.4", false}, {"2606:4700:4700::1111", false},
		{"8.8.8.8", true}, {"::ffff:8.8.8.8", true}, {"2001:4860:4860::8888", true},
	} {
		t.Run(test.ip, func(t *testing.T) {
			if allowed := wasmDestinationAllowed(wasmTestAddress(test.ip)); allowed != test.allowed {
				t.Fatalf("allowed = %t, want %t", allowed, test.allowed)
			}
		})
	}
}

func wasmTestAddress(address string) wasi.SocketAddress {
	ip := netip.MustParseAddr(address)
	if ip.Is4() {
		return &wasi.Inet4Address{Addr: ip.As4(), Port: 80}
	}
	return &wasi.Inet6Address{Addr: ip.As16(), Port: 80}
}

func TestWASMEgressSocketCalls(t *testing.T) {
	ctx := context.Background()
	host := &wasmTestEgressSystem{}
	for _, test := range []struct {
		internet bool
		peer     wasi.SocketAddress
		want     wasi.Errno
	}{
		{true, wasmTestAddress("127.0.0.1"), wasi.EPERM},
		{true, wasmTestAddress("::ffff:127.0.0.1"), wasi.EPERM},
		{true, &wasi.UnixAddress{Name: "/socket"}, wasi.EPERM},
		{false, wasmTestAddress("8.8.8.8"), wasi.EPERM},
		{true, wasmTestAddress("8.8.8.8"), wasi.ESUCCESS},
		{true, wasmTestAddress("2001:4860:4860::8888"), wasi.ESUCCESS},
	} {
		s := &wasmSystem{System: host, internet: test.internet}
		before := host.calls
		if peer, errno := s.SockConnect(ctx, 1, test.peer); errno != test.want || (errno == wasi.ESUCCESS && peer != test.peer) {
			t.Fatalf("connect to %v: %v %v, want %v", test.peer, peer, errno, test.want)
		}
		if size, errno := s.SockSendTo(ctx, 1, nil, 0, test.peer); errno != test.want || (errno == wasi.ESUCCESS && size != 1) {
			t.Fatalf("sendto %v: %v %v, want %v", test.peer, size, errno, test.want)
		}
		wantCalls := before
		if test.want == wasi.ESUCCESS {
			wantCalls += 2
		}
		if host.calls != wantCalls {
			t.Fatalf("host calls for %v: %d, want %d", test.peer, host.calls, wantCalls)
		}
	}
}

type wasmTestEgressSystem struct {
	wasi.System
	calls int
}

func (s *wasmTestEgressSystem) SockConnect(_ context.Context, _ wasi.FD, peer wasi.SocketAddress) (wasi.SocketAddress, wasi.Errno) {
	s.calls++
	return peer, wasi.ESUCCESS
}

func (s *wasmTestEgressSystem) SockSendTo(context.Context, wasi.FD, []wasi.IOVec, wasi.SIFlags, wasi.SocketAddress) (wasi.Size, wasi.Errno) {
	s.calls++
	return 1, wasi.ESUCCESS
}
