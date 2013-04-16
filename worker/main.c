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

	// TODO

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

