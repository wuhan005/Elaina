// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"net/netip"
	"slices"

	"github.com/stealthrocket/wasi-go"

	"github.com/wuhan005/Elaina/internal/config"
)

var wasmGlobalIPv6 = netip.MustParsePrefix("2000::/3")

var wasmDeniedDestinations = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("168.63.129.16/32"), // Azure WireServer uses a public IP internally.
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/3"), // Multicast and reserved space, including broadcast.
	netip.MustParsePrefix("2001::/23"),   // Protocol assignments, including Teredo.
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"), // 6to4 can tunnel to IPv4 destinations.
	netip.MustParsePrefix("3fff::/20"),
}

func wasmDestinationAllowed(peer wasi.SocketAddress) bool {
	var ip netip.Addr
	switch peer := peer.(type) {
	case *wasi.Inet4Address:
		ip = netip.AddrFrom4(peer.Addr)
	case *wasi.Inet6Address:
		ip = netip.AddrFrom16(peer.Addr).Unmap()
	default:
		return false
	}

	// IPv6 outside the global-unicast allocation includes local and well-known NAT64 addresses.
	if ip.Is6() && !wasmGlobalIPv6.Contains(ip) {
		return false
	}
	containsIP := func(prefix netip.Prefix) bool { return prefix.Contains(ip) }
	return !slices.ContainsFunc(wasmDeniedDestinations, containsIP) &&
		!slices.ContainsFunc(config.App.WASMDeniedCIDRs, containsIP)
}
