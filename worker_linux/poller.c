#include "poller.h"
#include "msgbuf.h"

#include "glib_extra.h"

#include <errno.h>
#include <signal.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/epoll.h>
#include <sys/signalfd.h>

G_DEFINE_QUARK (hammy-poller-error, hammy_poller_error)
#define E_DOMAIN hammy_poller_error_quark()

#define ERRNO_ERR(name) { \
	g_set_error (&lerr, \
		E_DOMAIN, \
		errno, \
		"[%s:%d] " name ": %s", \
		__FILE__, __LINE__, \
		g_strerror (errno)); \
	goto END; \
}


struct hammy_poller_priv
{
	guint pool_size;
	GTimeVal timeout;
	int epollfd;

	int input_fd;
	gpointer input_buf;
	gsize input_buf_size;
	GQueue *input_queue;

	int sig_fd;
};

hammy_poller_t
hammy_poller_new (struct hammy_poller_cfg *cfg, GError **error)
{
	GError *lerr = NULL;
	hammy_poller_t p;
	int esize;
	sigset_t sset;
	struct epoll_event ev;

	p = g_new0 (struct hammy_poller_priv, 1);
	g_assert(p != NULL);

	p->input_fd = cfg->input_fd;
	g_assert(p->input_fd > 0);
	p->pool_size = cfg->pool_size;
	g_assert(p->pool_size > 0);

	p->input_queue = g_queue_new();

	esize = p->pool_size + 2;
	p->epollfd = epoll_create (esize);
	if (p->epollfd < 0)
		ERRNO_ERR("epoll_create")

	if (fcntl (p->input_fd, F_SETFD, O_NONBLOCK) == -1)
		ERRNO_ERR("fcntl")

	ev.events = EPOLLIN | EPOLLERR | EPOLLHUP;
	ev.data.fd = p->input_fd;
	if (epoll_ctl (p->epollfd, EPOLL_CTL_ADD, p->input_fd, &ev) < 0)
		ERRNO_ERR("epoll_ctl");

	g_assert (sigemptyset (&sset) == 0);
	g_assert (sigaddset (&sset, SIGCHLD) == 0);
	p->sig_fd = signalfd (-1, &sset, SFD_NONBLOCK);
	if (p->sig_fd < 0)
		ERRNO_ERR("signalfd")

	ev.events = EPOLLIN | EPOLLERR | EPOLLHUP;
	ev.data.fd = p->sig_fd;
	if (epoll_ctl (p->epollfd, EPOLL_CTL_ADD, p->sig_fd, &ev) < 0)
		ERRNO_ERR("epoll_ctl");

END:
	if (lerr != NULL)
	{
		if (p != NULL)
			hammy_poller_free (p);
		g_propagate_error (error, lerr);
		return NULL;
	}
	else
	{
		return p;
	}
}

void
hammy_poller_free (hammy_poller_t p)
{
	if (p->sig_fd != 0)
		close (p->sig_fd);
	if (p->epollfd != 0)
		close (p->epollfd);

	if (p->input_buf != NULL)
		g_free (p->input_buf);
	if (p->input_queue != NULL)
		g_queue_free_full (p->input_queue, &g_free);

	g_free (p);
}

static gboolean
hammy_poller_handle_input (hammy_poller_t p, guint32 events, GError **error)
{
	GError *lerr = NULL;

	if (events & EPOLLERR)
	{
		lerr = g_error_new (E_DOMAIN, 0, "input_fd EPOLLERR");
		goto END;
	}
	if (events & EPOLLHUP)
	{
		lerr = g_error_new (E_DOMAIN, 0, "input_fd EPOLLHUP");
		goto END;
	}
	if (events & EPOLLIN)
	{

	}

END:
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

static gboolean
hammy_poller_handle_sig (hammy_poller_t p, guint32 events, GError **error)
{
	return TRUE;
}

static gboolean
hammy_poller_handle_fd (hammy_poller_t p, int fd, guint32 events, GError **error)
{
	return TRUE;
}

static gboolean
hammy_poller_handle (hammy_poller_t p, int fd, guint32 events, GError **error)
{
	GError *lerr = NULL;

	if (fd == p->input_fd)
	{
		if (!hammy_poller_handle_input (p, events, &lerr))
			goto END;
	}
	else if (fd == p->sig_fd)
	{
		if (!hammy_poller_handle_sig (p, events, &lerr))
			goto END;
	}
	else
	{
		if (!hammy_poller_handle_fd (p, fd, events, &lerr))
			goto END;
	}

END:
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
hammy_poller_run (hammy_poller_t p, GError **error)
{
	GError *lerr = NULL;
	int esize = p->pool_size + 2;
	struct epoll_event *events = NULL;
	int nfds, etimeout;
	int i;

	etimeout = 1000 * p->timeout.tv_sec + p->timeout.tv_usec / 1000;
	events = g_new(struct epoll_event, esize);

	for (;;)
	{
		nfds = epoll_wait(p->epollfd, events, esize, etimeout);
		if (nfds < 0)
			ERRNO_ERR("epoll_wait")
		// for each ready socket
		for(i = 0; i < nfds; i++)
		{
			// handle io on socket fd
			if (!hammy_poller_handle (p,
						events[i].data.fd,
						events[i].events,
						&lerr))
				goto END;
		}
	}

END:
	if (events != NULL)
		g_free(events);
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
