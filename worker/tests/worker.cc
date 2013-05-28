#include "../worker.h"

#include <gtest/gtest.h>
#include "null_asserts.hh"

TEST (Worker, Test1)
{
	GError *err = NULL;

	hammy_worker_t w = hammy_worker_new (&err);
	ASSERT_NOT_NULL (w);
	EXPECT_NULL (err);
	hammy_worker_free (w);
}
