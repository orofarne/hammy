#pragma once

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif

struct hammy_worker_priv;
typedef struct hammy_worker_priv *hammy_worker_t;

hammy_worker_t
hammy_worker_new (GError **error);

void
hammy_worker_free (hammy_worker_t self);

void
hammy_worker_free_ptr (gpointer ptr);

gboolean
hammy_worker_is_busy (hammy_worker_t self);

#ifdef __cplusplus
} // extern "C"
#endif
