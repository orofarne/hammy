#include "worker.h"

#include "child.h"

#include "glib_defines.h"

#include <unistd.h>

G_DEFINE_QUARK (hammy-worker-error, hammy_worker_error)
#define E_DOMAIN hammy_worker_error_quark()

struct hammy_worker_priv
{
	gboolean busy;
	pid_t pid;
	int to_pfd[2];
	int from_pfd[2];
};

static gboolean
hammy_worker_child (hammy_worker_t self, _H_AERR)
{
	FUNC_BEGIN()

	hammy_child_t ch = NULL;

	ch = hammy_child_new (self->to_pfd[1], self->from_pfd[0], ERR_RETURN);
	if (!ch)
		H_ASSERT_ERROR

	if (!hammy_child_run (ch, ERR_RETURN))
		H_ASSERT_ERROR

	FUNC_END(hammy_child_free(ch))
}

static gboolean
hammy_worker_fork (hammy_worker_t self, _H_AERR)
{
	FUNC_BEGIN()

	pid_t pid = fork();

	if (pid < 0)
		ERRNO_ERR ("fork");

	if (pid == 0)
	{
		close (self->to_pfd[0]);
		close (self->from_pfd[1]);

		if (!hammy_worker_child (self, ERR_RETURN))
		{
			g_assert (lerr != NULL);
			g_error ("[WORKER] Child error: %s", lerr->message);
		}
	}
	else
	{
		self->pid = pid;
		close (self->to_pfd[1]);
		close (self->from_pfd[0]);
		// TODO
		g_warning ("pid = %d", pid);
	}

	FUNC_END()
}

hammy_worker_t
hammy_worker_new (GError **error)
{
	GError *lerr = NULL;

	struct hammy_worker_priv *self = g_new0 (struct hammy_worker_priv, 1);

	g_warning("new worker");

	if (pipe(self->to_pfd) < 0)
		ERRNO_ERR ("pipe (to)");
	if (pipe(self->from_pfd) < 0)
		ERRNO_ERR ("pipe (from)");

	if (!hammy_worker_fork (self, ERR_RETURN))
		H_ASSERT_ERROR

END:
	if (lerr != NULL)
	{
		close (self->to_pfd[0]); close (self->to_pfd[1]);
		close (self->from_pfd[0]); close (self->from_pfd[1]);
		g_free (self);
		g_propagate_error (error, lerr);
		return NULL;
	}
	return self;
}

void
hammy_worker_free (hammy_worker_t self)
{
	close (self->to_pfd[0]);
	close (self->from_pfd[1]);
	g_free (self);
}

void
hammy_worker_free_ptr (gpointer ptr)
{
	hammy_worker_free ((hammy_worker_t)ptr);
}

gboolean
hammy_worker_is_busy (hammy_worker_t self)
{
	return self->busy;
}
