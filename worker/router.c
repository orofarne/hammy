#include "router.h"

#include "msgbuf.h"
#include "worker.h"
#include "glib_defines.h"

#include <string.h>
#include <sys/param.h> // MAXPATHLEN

// Network Stuff
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <fcntl.h>
#include <unistd.h>

// Libev
#include <ev.h>

G_DEFINE_QUARK (hammy-router-error, hammy_router_error)
#define E_DOMAIN hammy_router_error_quark()

#define HAMMY_ROUTER_DEF_DATA_SIZE 1024

struct hammy_router_client
{
	ev_io io;
	int fd;

	gpointer buffer;
	gsize buffer_size;
	gsize buffer_capacity;

	struct hammy_router_priv* server;
};

struct hammy_router_task
{
	gpointer data;
	gsize data_size;

	struct hammy_router_client *client;
};

struct hammy_router_priv
{
	struct ev_loop *loop;

	// Config
	gchar* sock_path;
	gint sock_backlog;

	// In socket
	struct sockaddr_un socket_un;
	gsize socket_un_len;
	int in_fd;
	ev_io in_io;

	// Clients
	GSList *clients;
	GQueue *tasks;

	// Workers
	guint max_workers;
	GPtrArray *workers;

	// Error
	GError *error;
};

static void
hammy_router_client_free (gpointer ptr)
{
	struct hammy_router_client *client = (struct hammy_router_client *)ptr;

	if (client == NULL)
		return;

	if (client->buffer != NULL)
		g_free (client->buffer);
	// FIXME
	ev_io_stop(client->server->loop, &client->io);
	close (client->fd);
	g_free (client);
}

static void
hammy_router_task_free (gpointer ptr)
{
	struct hammy_router_task *task = (struct hammy_router_task *)ptr;

	if (task->data != NULL)
		g_free (task->data);
	g_free (task);
}

// Simply adds O_NONBLOCK to the file descriptor of choice
static gboolean
hammy_router_setnonblock(int fd, _H_AERR)
{
	FUNC_BEGIN()

	int flags;

	flags = fcntl (fd, F_GETFL);
	flags |= O_NONBLOCK;
	if (fcntl (fd, F_SETFL, flags) < 0)
		ERRNO_ERR ("fcntl")

	FUNC_END()
}

// We have a new task for worker
static gboolean
hammy_router_touch_worker (hammy_router_t self, _H_AERR)
{
	GError *lerr = NULL;
	guint i;

	g_warning ("touch...");

	for (i = 0; i < self->workers->len; ++i)
	{
		if (!hammy_worker_is_busy ((hammy_worker_t)self->workers->pdata[i]))
		{
			// TODO
			g_warning ("TODO %d", __LINE__);
			return TRUE;
		}
	}

	if (lerr != NULL)
		g_propagate_error (error, lerr);
	return FALSE;
}

static gboolean
hammy_router_touch_workers (hammy_router_t self, _H_AERR)
{
	FUNC_BEGIN()

	struct hammy_router_task *task;

	for (;;)
	{
		task = (struct hammy_router_task *)g_queue_pop_head (self->tasks);
		if (task == NULL)
			break;

		if (!hammy_router_touch_worker (self, ERR_RETURN))
		{
			H_ASSERT_ERROR

			if (self->workers->len < self->max_workers)
			{
				// Create new worker
				hammy_worker_t w = hammy_worker_new (ERR_RETURN);
				g_ptr_array_add (self->workers, w);
				// Return task to queue
				g_queue_push_head (self->tasks, task);
				continue;
			}
			else
			{
				// Return task to queue
				g_queue_push_head (self->tasks, task);
				break;
			}
		}
	}

	FUNC_END()
}

// This callback is called when client data is readable on the unix socket.
static void
hammy_router_client_cb (struct ev_loop *loop, ev_io *w, int revents)
{
	struct hammy_router_client *client = (struct hammy_router_client *)w->data;
	int n;
	gsize m;
	struct hammy_router_task *task;
	GError *err = NULL;

	n = recv(client->fd, client->buffer, client->buffer_capacity, MSG_DONTWAIT);
	if (n < 0)
	{
		// TODO: handle error
		g_error ("n = %d, err: %s", n, strerror(errno));
	}
	if (n == 0)
	{
		// an orderly disconnect
		g_warning ("disconnected"); // FIXME

		client->server->clients = g_slist_remove (client->server->clients, client);
		hammy_router_client_free (client);
		return;
	}
	client->buffer_size += n;

	m = hammy_msg_buf_read (client->buffer, client->buffer_size, &err);
	if (err != NULL)
	{
		// TODO
		g_error ("hammy_msg_buf_read: %s", err->message);
	}
	if (m > 0)
	{
		task = g_new0 (struct hammy_router_task, 1);
		task->data = g_memdup (client->buffer, m);
		task->data_size = m;
		g_queue_push_tail (client->server->tasks, task);
		if (m != client->buffer_size)
		{
			memmove (client->buffer, (char *)client->buffer + m, client->buffer_size - m);
		}
		client->buffer_size -= m;

		if (!hammy_router_touch_workers (client->server, &err))
		{
			// TODO
		}
	}
	else
	{
		if (n == client->buffer_capacity)
		{
			client->buffer_capacity *= 2;
			client->buffer = g_realloc (client->buffer, client->buffer_capacity);
			return; // ?...
		}
	}
}

// This callback is called when data is readable on the unix socket.
static void
hammy_router_accept_cb (struct ev_loop *loop, ev_io *w, int revents)
{
	hammy_router_t self = (hammy_router_t)w->data;
	struct hammy_router_client *client;
	GError *err = NULL;

	g_assert (revents & EV_READ);

	// New client

	client = g_new0 (struct hammy_router_client, 1);

	client->fd = accept (self->in_fd, NULL, NULL);
	if (client->fd < 0)
	{
		E_SET_ERRNO (&self->error, "accept");
		ev_loop_destroy (EV_DEFAULT_UC); // ???
		g_free (client);
		return; // break;
	}

	hammy_router_setnonblock (client->fd, &err);
	if (err != NULL)
	{
		g_propagate_error (&self->error, err);
		ev_loop_destroy (EV_DEFAULT_UC); // ???
		g_free (client);
		return; // break;
	}

	client->buffer_capacity = HAMMY_ROUTER_DEF_DATA_SIZE;
	client->buffer = g_malloc0 (client->buffer_capacity);
	client->server = self;
	client->io.data = client;
	ev_io_init (&client->io, hammy_router_client_cb, client->fd, EV_READ);
	ev_io_start (self->loop, &client->io);

	self->clients = g_slist_prepend (self->clients, client);
	client = NULL;
}

static gboolean
hammy_router_unix_socket_init (hammy_router_t self, _H_AERR)
{
	FUNC_BEGIN()
	int fd;

	if (unlink (self->sock_path) < 0 && errno != ENOENT)
		ERRNO_ERR ("unlink")

	// Setup a unix socket listener.
	fd = socket (AF_UNIX, SOCK_STREAM, 0);
	if (-1 == fd)
		ERRNO_ERR ("filed to create socket")

	// Set it non-blocking
	H_TRY (hammy_router_setnonblock (fd, ERR_RETURN))

	g_assert (strlen (self->sock_path) <=  MAXPATHLEN);

	// Set it as unix socket
	self->socket_un.sun_family = AF_UNIX;
	strcpy (self->socket_un.sun_path, self->sock_path);

	self->socket_un_len = sizeof (self->socket_un.sun_family) + strlen (self->socket_un.sun_path);
	self->in_fd = fd;

	FUNC_END()
}

static gboolean
hammy_router_listen (hammy_router_t self, _H_AERR)
{
	FUNC_BEGIN()

	if (bind (self->in_fd, (struct sockaddr*)&self->socket_un, self->socket_un_len) < 0)
		ERRNO_ERR ("bind")

	if (listen (self->in_fd, self->sock_backlog) < 0)
		ERRNO_ERR ("listen")

	// Get notified whenever the socket is ready to read
	self->in_io.data = self;
	ev_io_init (&self->in_io, hammy_router_accept_cb, self->in_fd, EV_READ);
	ev_io_start (self->loop, &self->in_io);
	ev_run (self->loop, 0);

	FUNC_END()
}

hammy_router_t
hammy_router_new (struct hammy_router_cfg *cfg, GError **error)
{
	hammy_router_t self;

	self = g_new0 (struct hammy_router_priv, 1);

	self->sock_path = g_strdup (cfg->sock_path);
	self->sock_backlog = cfg->sock_backlog;

	self->clients = g_slist_alloc ();
	self->tasks = g_queue_new ();

	self->workers = g_ptr_array_new ();
	g_ptr_array_set_free_func (self->workers, hammy_worker_free_ptr);
	self->max_workers = cfg->max_workers;

	self->loop = ev_default_loop (EVFLAG_AUTO);

	return self;
}

void
hammy_router_free (hammy_router_t self)
{
	if (self->sock_path != NULL)
		g_free (self->sock_path);
	if (self->tasks != NULL)
		g_queue_free_full (self->tasks, hammy_router_task_free);
	if (self->clients != NULL)
		g_slist_free_full (self->clients, hammy_router_client_free);
	if (self->workers != NULL)
		g_ptr_array_free (self->workers, TRUE);
	if (self->loop != NULL)
		ev_loop_destroy (self->loop);
	g_free (self);
}

gboolean
hammy_router_run (hammy_router_t self, GError **error)
{
	FUNC_BEGIN()

	H_TRY (hammy_router_unix_socket_init (self, ERR_RETURN));

	H_TRY (hammy_router_listen (self, ERR_RETURN));

	if (self->error != NULL)
	{
		g_propagate_error (ERR_RETURN, self->error);
		return FALSE;
	}

	FUNC_END()
}

gboolean
hammy_router_stop (hammy_router_t self, GError **error)
{
	FUNC_BEGIN()

	// TODO

	if (self->tasks != NULL)
	{
		g_queue_free_full (self->tasks, hammy_router_task_free);
		self->tasks = NULL;
	}

	if (self->clients != NULL)
	{
		g_slist_free_full (self->clients, hammy_router_client_free);
		self->clients = NULL;
	}

	ev_io_stop (self->loop, &self->in_io);
	ev_break (self->loop, EVBREAK_ALL);

	FUNC_END()
}
