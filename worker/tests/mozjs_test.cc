#include <gtest/gtest.h>

#include "../eval.hh"

TEST(MozJsTest, SimpleScript) {
	hammy::MozJSEval eval;
	EXPECT_EQ(0, eval.init());
	const char *script = "var x = 2 + 2;\n";
	EXPECT_EQ(0, eval.eval(script));
	EXPECT_EQ("", eval.last_error());
}

TEST(MozJsTest, SimpleBadScript) {
	hammy::MozJSEval eval;
	EXPECT_EQ(0, eval.init());
	EXPECT_EQ("", eval.last_error());
	const char *script = "var x 2 + 2;\n";
	EXPECT_NE(0, eval.eval(script));
	EXPECT_NE("", eval.last_error());
}

TEST(MozJsTest, BadAndGood) {
	hammy::MozJSEval eval;
	EXPECT_EQ(0, eval.init());
	const char *script = "var x 2 + 2;\n";
	EXPECT_NE(0, eval.eval(script));
	EXPECT_NE("", eval.last_error());

	const char *script2 = "var x = 2 + 2;\n";
	EXPECT_EQ(0, eval.eval(script2));
	EXPECT_EQ("", eval.last_error());
}

TEST(MozJsTest, Cmd) {
	hammy::MozJSEval eval;
	EXPECT_EQ(0, eval.init());
	const char *script = "cmd('log');\n";
	EXPECT_EQ(0, eval.eval(script));
	EXPECT_EQ("", eval.last_error());

	const char *script2 = "cmd('log', {'message': 'hello'});\n";
	EXPECT_EQ(0, eval.eval(script2));
	EXPECT_EQ("", eval.last_error());
}