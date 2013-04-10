#include "poller.h"

#include <stdlib.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <math.h>

#include <glib.h>
#include "glib_extra.h"

G_DEFINE_QUARK(hammy-main-error, hammy_main_error)
#define E_DOMAIN hammy_main_error_quark()

static gchar *cfg_file = NULL;

static GOptionEntry entries[] =
{
	{ "config", 'c', 0, G_OPTION_ARG_STRING, &cfg_file, "Config file path", "FILE" },
	{ NULL }
};

gboolean
init_poller (GKeyFile *cfg, gint pipe_fd, GError **error)
{
	struct hammy_poller_cfg p_conf;
	GError *lerr;
	gdouble timeout, t_sec, t_usec;
	hammy_poller_t poller = NULL;

	p_conf.pool_size = g_key_file_get_uint64 (cfg, "worker", "pool_size", &lerr);
	if (lerr != NULL)
		goto END;

	timeout = g_key_file_get_double (cfg, "worker", "timeout", &lerr);
	if (lerr != NULL)
		goto END;
	if (timeout < 0)
	{
		g_set_error (&lerr, E_DOMAIN, EINVAL, "Negative timeout");
		goto END;
	}

	t_usec = G_USEC_PER_SEC * modf(timeout, &t_sec);
	p_conf.timeout.tv_sec = t_sec;
	p_conf.timeout.tv_usec = t_usec;

	poller = hammy_poller_new(&p_conf, &lerr);
	if (poller == NULL)
		goto END;

END:
	if (poller != NULL)
		hammy_poller_free(poller);
	if (lerr != NULL)
	{
		g_propagate_error (error, lerr);
		return FALSE;
	}
	else
	{
		return TRUE;
	}
}

gboolean
hammy_start (GKeyFile *cfg, GError **error)
{
	GError *lerr = NULL;
	gchar *pf_name = NULL;
	gint fd = 0;

	pf_name = g_key_file_get_string (cfg, "worker", "pipe", &lerr);
	if (lerr != NULL)
		goto END;

	if (mkfifo (pf_name, S_IWUSR | S_IRUSR | S_IRGRP | S_IROTH) != 0)
	{
		g_set_error (&lerr,
			E_DOMAIN,
			errno,
			"Failed to create pipe: %s",
			g_strerror (errno));
		goto END;
	}

	fd = open(pf_name, 0);
	if (fd < 0)
	{
		g_set_error (&lerr,
			E_DOMAIN,
			errno,
			"Failed to open pipe: %s",
			g_strerror (errno));
		goto END;
	}

	if (!init_poller (cfg, fd, &lerr))
		goto END;

END:
	if (pf_name != NULL)
		g_free (pf_name);
	if (lerr != NULL)
	{
		g_propagate_error (error, lerr);
		return FALSE;
	}
	else
	{
		return TRUE;
	}
}

int
main (int argc, char *argv[])
{
	GError *error = NULL;
	GOptionContext *context;
	GKeyFile *cfg_kv = NULL;
	gint rc = 1;

	context = g_option_context_new (NULL);
	g_option_context_add_main_entries (context, entries, NULL);
	if (!g_option_context_parse (context, &argc, &argv, &error))
	{
		g_print ("option parsing failed: %s\n", error->message);
		goto END;
	}

	if (cfg_file == NULL) {
		gchar *help_str = g_option_context_get_help (context, TRUE, NULL);
		g_print ("%s", help_str);
		g_free (help_str);
		goto END;
	}

	cfg_kv = g_key_file_new ();
	if (!g_key_file_load_from_file (cfg_kv, cfg_file, 0, &error))
	{
		g_print ("failed to read config: %s\n", error->message);
		goto END;
	}

	if (!hammy_start(cfg_kv, &error)) {
		g_print ("failed to start hammy: %s\n", error->message);
		goto END;
	}

	rc = 0;
END:
	if (error != NULL)
		g_error_free (error);
	if (context != NULL)
		g_option_context_free (context);
	if (cfg_kv != NULL)
		g_key_file_free (cfg_kv);
	return rc;
}

