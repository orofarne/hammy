#include "../msgbuf.h"

#include <gtest/gtest.h>
#include "null_asserts.hh"

#include <unistd.h>

extern "C" int _binary_cases_mpac_start;
extern "C" int _binary_cases_mpac_size;

TEST (MsgBuf, Test1)
{
	GError *error = NULL;

	char buf[] = {'\xa5', 'H', 'e', 'l', 'l', 'o'};
	gsize n = 6;

	EXPECT_EQ (n, hammy_msg_buf_read (buf, n, &error));
	EXPECT_NULL (error);
}

TEST (MsgBuf, CasesMpac)
{
	GError *error = NULL;

	EXPECT_EQ ((size_t)&_binary_cases_mpac_size,
		hammy_msg_buf_read ((char *)&_binary_cases_mpac_start, (size_t)&_binary_cases_mpac_size, &error));

	EXPECT_NULL (error);
}

TEST (MsgBuf, CasesMpacPart)
{
	GError *error = NULL;

	size_t n = (size_t)&_binary_cases_mpac_size;
	size_t n1 = n / 2;

	EXPECT_EQ (0,
		hammy_msg_buf_read ((char *)&_binary_cases_mpac_start, n1, &error));
	EXPECT_NULL (error);

	EXPECT_EQ (n,
		hammy_msg_buf_read ((char *)&_binary_cases_mpac_start, n, &error));
	EXPECT_NULL (error);
}

TEST (MsgBuf, BadData)
{
	GError *error = NULL;

	char buf[] = {'\xd7', 'H', 'e', 'l', 'l', 'o'};
	gsize n = 6;

	EXPECT_EQ (0, hammy_msg_buf_read (buf, n, &error));
	EXPECT_NOT_NULL (error);
}