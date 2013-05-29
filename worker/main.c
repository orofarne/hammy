#include "router.h"

#include <glib.h>
#include "glib_defines.h"

#include <string.h>

G_DEFINE_QUARK(hammy-main-error, hammy_main_error)
#define E_DOMAIN hammy_main_error_quark()

static gchar *cfg_file = NULL;

static GOptionEntry entries[] =
{
	{ "config", 'c', 0, G_OPTION_ARG_STRING, &cfg_file, "Config file path", "FILE" },
	{ NULL }
};

gboolean
start_router (GKeyFile *cfg_kv, _H_AERR)
{
	FUNC_BEGIN()
	hammy_router_t router = NULL;
	struct hammy_router_cfg cfg;

	memset(&cfg, sizeof(cfg), 0);

	cfg.sock_path = g_key_file_get_string (cfg_kv, "worker", "sock_path", ERR_RETURN);
	H_ASSERT_ERROR
	cfg.sock_backlog = g_key_file_get_uint64 (cfg_kv, "worker", "sock_backlog", ERR_RETURN);
	H_ASSERT_ERROR
	cfg.max_workers = g_key_file_get_uint64 (cfg_kv, "worker", "max_workers", ERR_RETURN);
	H_ASSERT_ERROR

	router = hammy_router_new (&cfg, ERR_RETURN);
	if (router == NULL)
		GOTO_END

	if (!hammy_router_run (router, ERR_RETURN))
		GOTO_END

	FUNC_END(
		if (cfg.sock_path != NULL)
			g_free (cfg.sock_path);
		if (router != NULL)
			hammy_router_free (router);
	)
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

	if (!start_router(cfg_kv, &error))
	{
		g_print ("failed to start router: %s\n", error->message);
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

