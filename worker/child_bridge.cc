#include "child.h"
#include "child.hh"

#include <exception>

#include "glib_defines.h"

G_DEFINE_QUARK (hammy-child-error, hammy_child_error)
#define E_DOMAIN hammy_child_error_quark()

struct hammy_child_priv
{
	hammy::Worker *w;
};

extern "C"
hammy_child_t
hammy_child_new (struct ev_loop *loop, int in_sock, int out_sock, GError **error)
{
	GError *lerr = NULL;
	try
	{
		hammy_child_t self  = new struct hammy_child_priv ();
		self->w = new hammy::Worker (loop, in_sock, out_sock);
		return self;
	}
	catch (std::exception const &e)
	{
		g_set_error (&lerr, E_DOMAIN, 1, "Exception: %s", e.what());
	}
	catch (...)
	{
		g_set_error (&lerr, E_DOMAIN, 1, "Exception: <unknown>");
	}
	if (lerr != NULL)
		g_propagate_error (error, lerr);
	return NULL;
}

extern "C"
gboolean
hammy_child_run (hammy_child_t self, GError **error)
{
	GError *lerr = NULL;
	try
	{
		self->w->run ();
		return TRUE;
	}
	catch (std::exception const &e)
	{
		g_set_error (&lerr, E_DOMAIN, 1, "Exception: %s", e.what());
	}
	catch (...)
	{
		g_set_error (&lerr, E_DOMAIN, 1, "Exception: <unknown>");
	}
	if (lerr != NULL)
		g_propagate_error (error, lerr);
	return FALSE;
}

extern "C"
void
hammy_child_free (hammy_child_t self)
{
	if (self == NULL)
		return;

	try {
		if (self->w)
			delete self->w;
		delete self;
	}
	catch (...)
	{
		g_assert (0);
	}
}
