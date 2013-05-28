#include "../router.h"

#include <gtest/gtest.h>
#include "null_asserts.hh"

#include <string.h>
#include <errno.h>

// Network Stuff
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <fcntl.h>
#include <unistd.h>

gpointer
h_router_thread_func (gpointer data)
{
	GError *err = NULL;
	gboolean rc;
	hammy_router_t r = (hammy_router_t)data;

	rc = hammy_router_run (r, &err);
	g_assert ((rc ? 1 : 0) == (err ? 0 : 1));

	return err;
}

TEST (Router, Test1)
{
	GError *error = NULL;
	struct hammy_router_cfg cfg;
	GThread *th = NULL;
	int sock, msgsock, rval;
	struct sockaddr_un server;

	char buf1[] = {'\xa5', 'H', 'e', 'l', 'l', 'o'};

	memset (&cfg, sizeof(cfg), 0);
	cfg.sock_path = g_strdup ("/tmp/hammy_worker_test.socket");
	cfg.sock_backlog = 100;
	cfg.max_workers = 1;

	hammy_router_t r = hammy_router_new (&cfg, &error);
	ASSERT_NOT_NULL (r);
	ASSERT_NULL (error);

	th = g_thread_new ("router_thread", h_router_thread_func, r);
	g_usleep (1000);

	sock = socket (AF_UNIX, SOCK_STREAM, 0);
	if (sock < 0)
		g_error ("opening stream socket");
	server.sun_family = AF_UNIX;
	strcpy(server.sun_path, cfg.sock_path);
	if (connect (sock, (struct sockaddr *) &server, sizeof (struct sockaddr_un)) != 0)
	{
		close (sock);
		g_error ("connecting stream socket: %s", strerror (errno));
	}

	// TODO
	EXPECT_EQ (sizeof (buf1), write (sock, buf1, sizeof (buf1)));

	close (sock);

	g_usleep (100000); // FIXME

	ASSERT_EQ (TRUE, hammy_router_stop (r, &error));
	ASSERT_NULL (error);

	// ======= bad
	sock = socket (AF_UNIX, SOCK_STREAM, 0);
	if (sock < 0)
		g_error ("opening stream socket");
	if (connect (sock, (struct sockaddr *) &server, sizeof (struct sockaddr_un)) != 0)
	{
		close (sock);
		g_error ("connecting stream socket: %s", strerror(errno));
	}
	close (sock);
	/// =======

	error = (GError *)g_thread_join (th);
	ASSERT_NULL (error);

	g_free (cfg.sock_path);
	hammy_router_free (r);
}
