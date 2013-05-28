#include "child.h"
#include "child.hh"

#include <exception>


struct hammy_child_priv
{
	hammy::Worker *w;
};

extern "C"
hammy_child_t
hammy_child_new (int in_sock, int out_sock, GError **error)
{
	try
	{
		hammy_child_t self  = new struct hammy_child_priv ();
		self->w = new hammy::Worker (in_sock, out_sock);
		return self;
	}
	catch (std::exception const &e)
	{
		g_assert (0); // TODO
	}
	catch (...)
	{
		g_assert (0); // TODO
	}

	return NULL;
}

extern "C"
gboolean
hammy_child_run (hammy_child_t self, GError **error)
{
	try
	{
		self->w->run ();
	}
	catch (std::exception const &e)
	{
		g_assert (0); // TODO
	}
	catch (...)
	{
		g_assert (0); // TODO
	}
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
