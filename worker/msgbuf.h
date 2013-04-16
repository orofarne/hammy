#pragma once

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif

// Reads binary buffer and returns size of first object in buffer.
// Returns 0 if message is incomplete or on error.
gsize
hammy_msg_buf_read (gpointer buf, gsize size, GError **error);

#ifdef __cplusplus
} // extern "C"
#endif