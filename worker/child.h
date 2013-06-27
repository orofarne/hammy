#pragma once

#include <glib.h>
// Libev
#include <ev.h>

#ifdef __cplusplus
extern "C" {
#endif

struct hammy_child_priv;
typedef struct hammy_child_priv *hammy_child_t;

hammy_child_t
hammy_child_new(struct ev_loop *loop, int in_sock, int out_sock, GError **error);

gboolean
hammy_child_run (hammy_child_t self, GError **error);

void
hammy_child_free (hammy_child_t self);

#ifdef __cplusplus
} // extern "C"
#endif
