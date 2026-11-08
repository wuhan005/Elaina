/* Copyright 2026 E99p1ant. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

#include <errno.h>
#include <netdb.h>
#include <netinet/in.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

/* The socket imports use wasi-go's WasmEdge preview1 ABI. */
#define WASI_IMPORT(name) __attribute__((import_module("wasi_snapshot_preview1"), import_name(name)))

struct wasi_address {
    const void *data;
    uint32_t length;
};

struct wasi_sockaddr {
    uint8_t family;
    uint32_t length;
    void *data;
};

struct wasi_addrinfo {
    uint16_t flags;
    uint8_t family;
    uint8_t type;
    uint32_t protocol;
    uint32_t length;
    struct wasi_sockaddr *address;
    char *canonical_name;
    uint32_t canonical_length;
    struct wasi_addrinfo *next;
};

WASI_IMPORT("sock_open") int wasi_sock_open(int, int, int *);
WASI_IMPORT("sock_connect") int wasi_sock_connect(int, const struct wasi_address *, uint32_t);
WASI_IMPORT("sock_getaddrinfo") int wasi_getaddrinfo(const char *, uint32_t, const char *, uint32_t,
    const struct wasi_addrinfo *, struct wasi_addrinfo **, uint32_t, uint32_t *);

int socket(int domain, int type, int protocol) {
    if (domain != AF_INET) {
        errno = EAFNOSUPPORT;
        return -1;
    }
    if (type != SOCK_STREAM || (protocol != 0 && protocol != IPPROTO_TCP)) {
        errno = EPROTONOSUPPORT;
        return -1;
    }
    int fd;
    int error = wasi_sock_open(1, 2, &fd);
    if (error) {
        errno = error;
        return -1;
    }
    return fd;
}

int connect(int fd, const struct sockaddr *address, socklen_t length) {
    if (!address || length < sizeof(struct sockaddr_in) || address->sa_family != AF_INET) {
        errno = EAFNOSUPPORT;
        return -1;
    }
    const struct sockaddr_in *ip = (const struct sockaddr_in *)address;
    struct wasi_address raw = { &ip->sin_addr, sizeof(ip->sin_addr) };
    int error = wasi_sock_connect(fd, &raw, ntohs(ip->sin_port));
    if (error) {
        errno = error;
        return -1;
    }
    return 0;
}

void freeaddrinfo(struct addrinfo *result) {
    while (result) {
        struct addrinfo *next = result->ai_next;
        free(result->ai_addr);
        free(result);
        result = next;
    }
}

int getaddrinfo(const char *host, const char *service, const struct addrinfo *hints, struct addrinfo **result) {
    *result = NULL;
    struct wasi_addrinfo hint = { .family = 1, .type = 2, .protocol = 1 };
    if (hints) {
        if (hints->ai_family != AF_UNSPEC && hints->ai_family != AF_INET) return EAI_FAMILY;
        if (hints->ai_socktype != 0 && hints->ai_socktype != SOCK_STREAM) return EAI_SOCKTYPE;
        if (hints->ai_protocol != 0 && hints->ai_protocol != IPPROTO_TCP) return EAI_SERVICE;
        if (hints->ai_flags & ~(AI_NUMERICHOST | AI_NUMERICSERV)) return EAI_BADFLAGS;
        if (hints->ai_flags & AI_NUMERICHOST) hint.flags |= 4;
        if (hints->ai_flags & AI_NUMERICSERV) hint.flags |= 8;
    }

    /* Bound both guest buffers and the host's DNS result allocation. */
    struct wasi_addrinfo raw[16] = {0};
    struct wasi_sockaddr addresses[16] = {0};
    uint8_t data[16][26] = {0};
    for (int i = 0; i < 16; i++) {
        addresses[i].data = data[i];
        addresses[i].length = sizeof(data[i]);
        raw[i].address = &addresses[i];
        raw[i].next = i < 15 ? &raw[i + 1] : NULL;
    }
    struct wasi_addrinfo *first = raw;
    uint32_t count = 0;
    int error = wasi_getaddrinfo(host, host ? strlen(host) : 0, service, service ? strlen(service) : 0,
        &hint, &first, 16, &count);
    if (error) {
        errno = error;
        return EAI_SYSTEM;
    }
    if (!count) return EAI_NONAME;
    struct addrinfo **next = result;
    for (uint32_t i = 0; i < count; i++) {
        struct addrinfo *entry = calloc(1, sizeof(*entry));
        struct sockaddr_in *ip = calloc(1, sizeof(*ip));
        if (!entry || !ip) {
            free(entry);
            free(ip);
            freeaddrinfo(*result);
            *result = NULL;
            return EAI_MEMORY;
        }
        ip->sin_family = AF_INET;
        memcpy(&ip->sin_port, data[i], 2);
        memcpy(&ip->sin_addr, data[i] + 2, 4);
        entry->ai_family = AF_INET;
        entry->ai_socktype = SOCK_STREAM;
        entry->ai_protocol = IPPROTO_TCP;
        entry->ai_addrlen = sizeof(*ip);
        entry->ai_addr = (struct sockaddr *)ip;
        *next = entry;
        next = &entry->ai_next;
    }
    return 0;
}

const char *gai_strerror(int error) {
    switch (error) {
    case 0: return "Success";
    case EAI_BADFLAGS: return "Unsupported address flags";
    case EAI_NONAME: return "No address found";
    case EAI_FAMILY: return "Only IPv4 is supported";
    case EAI_SOCKTYPE: return "Only stream sockets are supported";
    case EAI_SERVICE: return "Only TCP is supported";
    case EAI_MEMORY: return "Out of memory";
    case EAI_SYSTEM: return "System error; inspect errno";
    default: return "Address lookup failed";
    }
}
