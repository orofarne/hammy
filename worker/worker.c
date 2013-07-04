#include "worker.h"

#include "child.h"

#include "glib_defines.h"

#include <unistd.h>
#include <errno.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/socket.h>

G_DEFINE_QUARK (hammy-worker-error, hammy_worker_error)
#define E_DOMAIN hammy_worker_error_quark()

#define HAMMY_WORKER_DEF_BUFF_SIZE 1024

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
	gsize output_data_capacity;
	hammy_worker_task_cb callback;
	gpointer callback_private;
};

static gboolean
hammy_worker_child (hammy_worker_t self, _H_AERR)
{
	FUNC_BEGIN()

	hammy_child_t ch = NULL;

	ch = hammy_child_new (self->loop, self->to_pfd[0], self->from_pfd[1], ERR_RETURN);
	if (!ch)
		H_ASSERT_ERROR

	if (!hammy_child_run (ch, ERR_RETURN))
		H_ASSERT_ERROR

	FUNC_END(hammy_child_free(ch))
}

static gboolean
hammy_worker_task_done (hammy_worker_t self, _H_AERR)
{
	FUNC_BEGIN()

	gpointer buf =  g_memdup (self->output_data, self->output_data_size);
	gsize size = self->output_data_size;

	(*self->callback)(self->callback_private, buf, size, ERR_RETURN);

	self->output_data_size = 0;
	self->callback = NULL;
	self->callback_private = NULL;
	self->busy = FALSE;

	FUNC_END()
}

static void
hammy_worker_from_cb (struct ev_loop *loop, ev_io *w, int revents)
{
	ssize_t rc;
	gsize m_rc;
	GError *err = NULL;
	hammy_worker_t self = (hammy_worker_t)w->data;

	if (self->output_data_size == self->output_data_capacity) {
		self->output_data_capacity = self->output_data_capacity * 2 + HAMMY_WORKER_DEF_BUFF_SIZE;
		self->output_data = g_realloc (self->output_data, self->output_data_capacity);
	}

	rc = read (self->from_pfd[0], self->output_data + self->output_data_size, self->output_data_capacity);
	if (rc < 0)
		g_error ("read: %s", g_strerror (errno));

	if (rc == 0) {
		g_warning ("EOF");
		goto END;
	}

	self->output_data_size += rc;

	m_rc = hammy_msg_buf_read (self->output_data, self->output_data_size, &err);
	if (m_rc == 0) {
		if (err != NULL)
			g_error ("hammy_msg_buf_read: %s", err->message);

		return; // Wait for more data
	}

	if (m_rc != self->output_data_size) {
		g_warning ("Garbage in channel");
		self->output_data_size = m_rc;
	}

	// Process answer
	if (!hammy_worker_task_done (self, &err))
		g_error ("hammy_worker_task_done: %s", err->message);

END:
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
	rc = write (self->to_pfd[1], self->input_data, self->input_data_size);
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
	ev_io_init (&self->from_io, hammy_worker_from_cb, self->from_pfd[0], EV_READ);
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
		ev_loop_fork (self->loop);
		ev_loop_destroy (self->loop);
		self->loop = ev_default_loop (EVFLAG_AUTO);

		if (!hammy_worker_child (self, ERR_RETURN))
		{
			g_assert (lerr != NULL);
			g_error ("[WORKER] Child error: %s", lerr->message);
		}
	}
	else
	{
		self->pid = pid;
	}

	FUNC_END()
}

hammy_worker_t
hammy_worker_new (struct ev_loop *loop, GError **error)
{
	GError *lerr = NULL;

	struct hammy_worker_priv *self = g_new0 (struct hammy_worker_priv, 1);

	g_assert (loop != NULL);
	self->loop = loop;

	if (pipe (self->to_pfd) < 0)
		ERRNO_ERR ("pipe (to)");
	if (pipe (self->from_pfd) < 0)
		ERRNO_ERR ("pipe (from)");

	// Set nonblock
	if (fcntl(self->to_pfd[0], F_SETFL, O_NONBLOCK) < 0)
		ERRNO_ERR ("fcntl");
	if (fcntl(self->to_pfd[1], F_SETFL, O_NONBLOCK) < 0)
		ERRNO_ERR ("fcntl");
	if (fcntl(self->from_pfd[0], F_SETFL, O_NONBLOCK) < 0)
		ERRNO_ERR ("fcntl");
	if (fcntl(self->from_pfd[1], F_SETFL, O_NONBLOCK) < 0)
		ERRNO_ERR ("fcntl");

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
	close (self->to_pfd[0]); close (self->to_pfd[1]);
	close (self->from_pfd[0]); close (self->from_pfd[1]);
	g_free (self);
}

void
hammy_worker_free_ptr (gpointer ptr)
{
	hammy_worker_free ((hammy_worker_t)ptr);
}

gboolean
hammy_worker_task (hammy_worker_t self, gpointer data, gsize data_size, hammy_worker_task_cb callback, gpointer callback_private, GError **error)
{
	FUNC_BEGIN()

	ssize_t rc;

	g_assert (!hammy_worker_is_busy (self));
	self->busy = TRUE;

	self->callback = callback;
	self->callback_private = callback_private;

	// Try to write data immediately
	rc = write (self->to_pfd[1], data, data_size);
	if (rc < 0 && (errno == EAGAIN || errno == EWOULDBLOCK)) {
		// We can't write data immediately
		self->input_data = data;
		self->input_data_size = data_size;
		// Set up libev for new child write events
		self->to_io.data = self;
		ev_io_init (&self->to_io, hammy_worker_to_cb, self->to_pfd[1], EV_WRITE);
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
		ev_io_init (&self->to_io, hammy_worker_to_cb, self->to_pfd[1], EV_WRITE);
		ev_io_start(self->loop, &self->to_io);
		GOTO_END;
	}

	// All data wrote
	// Set up libev for new child read events
	self->from_io.data = self;
	ev_io_init (&self->from_io, hammy_worker_from_cb, self->from_pfd[0], EV_READ);
	ev_io_start (self->loop, &self->from_io);

	FUNC_END()
}

gboolean
hammy_worker_is_busy (hammy_worker_t self)
{
	return self->busy;
}
