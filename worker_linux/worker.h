#pragma once

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif

struct hammy_worker
{
	int fd;
	int pid;
};

#ifdef __cplusplus
} // extern "C"
#endif