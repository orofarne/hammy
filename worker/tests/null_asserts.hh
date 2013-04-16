#pragma once

#define ASSERT_NULL(p) ASSERT_EQ (NULL, (long)p)
#define ASSERT_NOT_NULL(p) ASSERT_NE (NULL, (long)p)

#define EXPECT_NULL(p) EXPECT_EQ (NULL, (long)p)
#define EXPECT_NOT_NULL(p) EXPECT_NE (NULL, (long)p)