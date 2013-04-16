#include "../poller.h"

#include <gtest/gtest.h>
#include "null_asserts.hh"

#include <unistd.h>

TEST (Poller, CreateDestroy)
{
	int pipefd[2];
	struct hammy_poller_cfg cfg;
	hammy_poller_t poller = NULL;
	GError *err = NULL;

	memset (&cfg, 0, sizeof (cfg));
	ASSERT_EQ (0, pipe2 (pipefd, 0));
	cfg.input_fd = pipefd[0];
	cfg.pool_size = 1;
	g_time_val_add (&cfg.timeout, G_USEC_PER_SEC); // 1 second

	poller = hammy_poller_new (&cfg, &err);
	ASSERT_NOT_NULL (poller);
	ASSERT_NULL (err);

	hammy_poller_free(poller);
	close (pipefd[0]); close (pipefd[1]);
}
