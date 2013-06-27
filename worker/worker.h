#pragma once

#include <glib.h>
// Libev
#include <ev.h>

#ifdef __cplusplus
extern "C" {
#endif

struct hammy_worker_priv;
typedef struct hammy_worker_priv *hammy_worker_t;

typedef gboolean (* hammy_worker_task_cb) (gpointer callback_private, gpointer data, gsize data_size, GError **error);

hammy_worker_t
hammy_worker_new (struct ev_loop *loop, GError **error);

void
hammy_worker_free (hammy_worker_t self);

void
hammy_worker_free_ptr (gpointer ptr);

gboolean
hammy_worker_task (hammy_worker_t self, gpointer data, gsize data_size, hammy_worker_task_cb callback, gpointer callback_private, GError **error);

gboolean
hammy_worker_is_busy (hammy_worker_t self);

#ifdef __cplusplus
} // extern "C"
#endif
