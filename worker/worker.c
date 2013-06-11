#include "worker.h"

#include "child.h"

#include "glib_defines.h"

#include <unistd.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/socket.h>

G_DEFINE_QUARK (hammy-worker-error, hammy_worker_error)
#define E_DOMAIN hammy_worker_error_quark()

struct hammy_worker_priv
{
	gboolean busy;
	pid_t pid;
	int to_pfd[2];
	int from_pfd[2];
	struct ev_loop *loop;
	ev_io to_io;
	ev_io from_io;
	gpointer input_data;
	gsize input_data_size;
	gpointer output_data;
	gsize output_data_size;
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

static void
hammy_worker_from_cb (struct ev_loop *loop, ev_io *w, int revents)
{
	ssize_t rc;
	hammy_worker_t self = (hammy_worker_t)w->data;

	// TODO
	//rc = recv (self->from_pfd[1], ... , MSG_DONTWAIT);

	// All data read
	// Stop libev for write events
	ev_io_stop(self->loop, &self->from_io);
}

static void
hammy_worker_to_cb (struct ev_loop *loop, ev_io *w, int revents)
{
	ssize_t rc;
	hammy_worker_t self = (hammy_worker_t)w->data;

	g_assert (self->input_data_size > 0);

	// TODO
	// Try to write data immediately
	rc = send (self->to_pfd[0], self->input_data, self->input_data_size, MSG_DONTWAIT);
	if (rc < 0 && (errno == EAGAIN || errno == EWOULDBLOCK)) {
		// We can't write data immediately
		return;
	}

	if (rc < 0) {
		// Error
		// TODO
		g_error ("hammy_worker_to_cb");
	}

	if (rc < self->input_data_size) {
		self->input_data = self->input_data + rc;
		self->input_data_size = self->input_data_size - rc;
		return;
	}

	// All data wrote
	// Stop libev for write events
	ev_io_stop(self->loop, &self->to_io);
	// Set up libev for new child read events
	self->from_io.data = self;
	ev_io_init (&self->from_io, hammy_worker_from_cb, self->from_pfd[1], EV_READ);
	ev_io_start (self->loop, &self->from_io);
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
		g_warning ("pid = %d", pid); // DEBUG

		self->pid = pid;
		close (self->to_pfd[1]);
		close (self->from_pfd[0]);
	}

	FUNC_END()
}

hammy_worker_t
hammy_worker_new (struct ev_loop *loop, GError **error)
{
	GError *lerr = NULL;

	struct hammy_worker_priv *self = g_new0 (struct hammy_worker_priv, 1);

	g_warning ("new worker"); // DEBUG

	g_assert (loop != NULL);
	self->loop = loop;

	if (pipe (self->to_pfd) < 0)
		ERRNO_ERR ("pipe (to)");
	if (pipe (self->from_pfd) < 0)
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
hammy_worker_task (hammy_worker_t self, gpointer data, gsize data_size, GError **error)
{
	FUNC_BEGIN()

	ssize_t rc;

	g_assert (!hammy_worker_is_busy (self));
	self->busy = TRUE;

	// Try to write data immediately
	rc = send (self->to_pfd[0], data, data_size, MSG_DONTWAIT);
	if (rc < 0 && (errno == EAGAIN || errno == EWOULDBLOCK)) {
		// We can't write data immediately
		self->input_data = data;
		self->input_data_size = data_size;
		// Set up libev for new child write events
		self->to_io.data = self;
		ev_io_init (&self->to_io, hammy_worker_to_cb, self->to_pfd[0], EV_WRITE);
		ev_io_start (self->loop, &self->to_io);
		GOTO_END;
	}

	if (rc < 0) {
		ERRNO_ERR ("hammy_worker send");
	}

	if (rc < data_size) {
		// Not all data sent
		self->input_data = data + rc;
		self->input_data_size = data_size - rc;
		// Set up libev for new child write events
		self->to_io.data = self;
		ev_io_init (&self->to_io, hammy_worker_to_cb, self->to_pfd[0], EV_WRITE);
		ev_io_start(self->loop, &self->to_io);
		GOTO_END;
	}

	// All data wrote
	// Set up libev for new child read events
	self->from_io.data = self;
	ev_io_init (&self->from_io, hammy_worker_from_cb, self->from_pfd[1], EV_READ);
	ev_io_start (self->loop, &self->from_io);

	FUNC_END()
}

gboolean
hammy_worker_is_busy (hammy_worker_t self)
{
	return self->busy;
}
