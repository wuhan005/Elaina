/* Copyright 2026 E99p1ant. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

#ifndef ELAINA_SOCKET_H
#define ELAINA_SOCKET_H

#include_next <sys/socket.h>

int socket(int, int, int);
int connect(int, const struct sockaddr *, socklen_t);

#endif
