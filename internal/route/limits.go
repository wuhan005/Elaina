// Copyright 2026 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"

	"github.com/flamego/flamego"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/context"
	"github.com/wuhan005/Elaina/internal/dto"
)

// bufferedRequestBytes is the body memory reserved by in-flight requests, released only after the request is handled.
var bufferedRequestBytes atomic.Int64

func reserveRequestBytes(n int64) bool {
	for {
		used := bufferedRequestBytes.Load()
		if used+n > config.App.WASMLimits.MaxBufferedRequestBytes {
			return false
		}
		if bufferedRequestBytes.CompareAndSwap(used, used+n) {
			return true
		}
	}
}

func limitRequestBody(ctx flamego.Context) {
	r := ctx.Request().Request
	if r.Body == nil {
		return
	}

	limit := config.App.WASMLimits.MaxRequestBytes
	if r.ContentLength > limit {
		_ = r.Body.Close()
		context.WriteJSON(ctx.ResponseWriter(), r, http.StatusRequestEntityTooLarge, dto.ErrorResponse{
			Msg: "Request body exceeds the configured size limit",
		})
		return
	}

	// Chunked bodies have unknown length, so reserve the full limit.
	reserved := r.ContentLength
	if reserved < 0 {
		reserved = limit
	}
	if !reserveRequestBytes(reserved) {
		_ = r.Body.Close()
		context.WriteJSON(ctx.ResponseWriter(), r, http.StatusServiceUnavailable, dto.ErrorResponse{
			Msg: "Server is busy; try again later",
		})
		return
	}
	defer bufferedRequestBytes.Add(-reserved)

	// Read before decoding so chunked bodies and trailing JSON cannot bypass the limit.
	data, err := io.ReadAll(http.MaxBytesReader(ctx.ResponseWriter(), r.Body, limit))
	_ = r.Body.Close()
	if err != nil {
		status, message := http.StatusBadRequest, "Failed to read request body"
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			status, message = http.StatusRequestEntityTooLarge, "Request body exceeds the configured size limit"
		}
		context.WriteJSON(ctx.ResponseWriter(), r, status, dto.ErrorResponse{
			Msg: message,
		})
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(data))
	ctx.Next()
}

// clientAddress returns the client key for rate limiting: behind trusted proxies it uses the rightmost untrusted
// X-Forwarded-For address, and IPv6 addresses are grouped by /64.
func clientAddress(r *http.Request) string {
	remote, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return "unknown"
	}

	addr := remote.Addr().Unmap()
	if trustedProxy(addr) {
		hops := strings.Split(strings.Join(r.Header.Values("X-Forwarded-For"), ","), ",")
		for i := len(hops) - 1; i >= 0; i-- {
			hop, err := netip.ParseAddr(strings.TrimSpace(hops[i]))
			if err != nil {
				break
			}
			addr = hop.Unmap()
			if !trustedProxy(addr) {
				break
			}
		}
	}

	if addr.Is6() {
		prefix, _ := addr.WithZone("").Prefix(64)
		return prefix.String()
	}
	return addr.String()
}

func trustedProxy(addr netip.Addr) bool {
	for _, prefix := range config.App.TrustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
