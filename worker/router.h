#pragma once

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif

struct hammy_router_cfg
{
	gchar* sock_path;
	gint sock_backlog;
	guint max_workers;
};

struct hammy_router_priv;
typedef struct hammy_router_priv *hammy_router_t;

hammy_router_t
hammy_router_new (struct hammy_router_cfg *cfg, GError **error);

void
hammy_router_free (hammy_router_t self);

gboolean
hammy_router_run (hammy_router_t self, GError **error);

gboolean
hammy_router_stop (hammy_router_t self, GError **error);

#ifdef __cplusplus
} // extern "C"
#endif
