#pragma once

#include <glib.h>

#ifdef __cplusplus
extern "C" {
#endif

struct hammy_poller_cfg
{
	int input_fd;
	guint pool_size;
	GTimeVal timeout;
};

struct hammy_poller_priv;
typedef struct hammy_poller_priv *hammy_poller_t;

hammy_poller_t
hammy_poller_new (struct hammy_poller_cfg *cfg, GError **error);

void
hammy_poller_free (hammy_poller_t p);

gboolean
hammy_poller_run (hammy_poller_t p);

#ifdef __cplusplus
} // extern "C"
#endif
