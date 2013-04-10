#include "poller.h"
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
	int sigfd;
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

	esize = p->pool_size + 2;
	p->epollfd = epoll_create (esize);
	if (p->epollfd < 0)
		ERRNO_ERR("epoll_create")

	if (fcntl (p->input_fd, F_SETFD, O_NONBLOCK) == -1)
		ERRNO_ERR("fcntl")

	ev.events = EPOLLIN | EPOLLPRI | EPOLLERR | EPOLLHUP;
	ev.data.fd = p->input_fd;
	if (epoll_ctl (p->epollfd, EPOLL_CTL_ADD, p->input_fd, &ev) < 0)
		ERRNO_ERR("epoll_ctl");

	g_assert (sigemptyset (&sset) == 0);
	g_assert (sigaddset (&sset, SIGCHLD) == 0);
	p->sigfd = signalfd (-1, &sset, SFD_NONBLOCK);
	if (p->sigfd < 0)
		ERRNO_ERR("signalfd")

	ev.events = EPOLLIN | EPOLLPRI | EPOLLERR | EPOLLHUP;
	ev.data.fd = p->sigfd;
	if (epoll_ctl (p->epollfd, EPOLL_CTL_ADD, p->sigfd, &ev) < 0)
		ERRNO_ERR("epoll_ctl");

	// TODO

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
	if (p->sigfd != 0)
		close (p->sigfd);
	if (p->epollfd != 0)
		close (p->epollfd);

	g_free (p);
}

gboolean
hammy_poller_run (hammy_poller_t p)
{
	return TRUE;
}
