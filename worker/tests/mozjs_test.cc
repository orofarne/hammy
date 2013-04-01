#include <gtest/gtest.h>

#include "../eval.hh"

TEST(MozJsTest, SimpleScript) {
	hammy::MozJSEval eval;
	EXPECT_EQ(eval.init(), 0);
	const char *script = "var x = 2 + 2;\n";
	EXPECT_EQ(eval.eval(script), 0);
}

TEST(MozJsTest, SimpleBadScript) {
	hammy::MozJSEval eval;
	EXPECT_EQ(eval.init(), 0);
	const char *script = "var x 2 + 2;\n";
	EXPECT_NE(eval.eval(script), 0);
	EXPECT_NE(eval.last_error(), "");
}

TEST(MozJsTest, BadAndGood) {
	hammy::MozJSEval eval;
	EXPECT_EQ(eval.init(), 0);
	const char *script = "var x 2 + 2;\n";
	EXPECT_NE(eval.eval(script), 0);
	EXPECT_NE(eval.last_error(), "");

	const char *script2 = "var x = 2 + 2;\n";
	EXPECT_EQ(eval.eval(script2), 0);
	EXPECT_EQ(eval.last_error(), "");
}