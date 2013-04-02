#include <gtest/gtest.h>

#include "../eval.hh"

TEST(MozJsTest, SimpleScript) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	eval.set_state(&s);
	const char *script = "var x = 2 + 2;\n";
	EXPECT_EQ(0, eval.compile(script, strlen(script)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
}

TEST(MozJsTest, SimpleBadScript) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	eval.set_state(&s);
	EXPECT_EQ("", eval.last_error());
	const char *script = "var x 2 + 2;\n";
	EXPECT_NE(0, eval.compile(script, strlen(script)));
	EXPECT_NE("", eval.last_error());
}

TEST(MozJsTest, BadAndGood) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	eval.set_state(&s);

	const char *script = "var x 2 + 2;\n";
	EXPECT_NE(0, eval.compile(script, strlen(script)));
	EXPECT_NE("", eval.last_error());

	const char *script2 = "var x = 2 + 2;\n";
	EXPECT_EQ(0, eval.compile(script2, strlen(script2)));
	EXPECT_EQ("", eval.last_error());
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
}

TEST(MozJsTest, Cmd) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	eval.set_state(&s);

	const char *script = "cmd('log');\n";
	EXPECT_EQ(0, eval.compile(script, strlen(script)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
	hammy::CmdBuf &cmdb = eval.get_cmdbuf();
	ASSERT_EQ(1, cmdb.size());
	EXPECT_EQ("log", cmdb[0].cmd);
	EXPECT_EQ(0, cmdb[0].opts.size());

	const char *script2 = "cmd('log', {'message': 'hello'});\n";
	EXPECT_EQ(0, eval.compile(script2, strlen(script2)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
	hammy::CmdBuf &cmdb2 = eval.get_cmdbuf();
	ASSERT_EQ(1, cmdb2.size());
	EXPECT_EQ("log", cmdb2[0].cmd);
	ASSERT_EQ(1, cmdb2[0].opts.size());
	EXPECT_TRUE(cmdb2[0].opts["message"].isString());
}

TEST(MozJsTest, Hostname) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	eval.set_state(&s);
	const char *script = "cmd('log', {'message': host});\n";
	EXPECT_EQ(0, eval.compile(script, strlen(script)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
	hammy::CmdBuf &cmdb = eval.get_cmdbuf();
	ASSERT_EQ(1, cmdb.size());
	EXPECT_EQ("log", cmdb[0].cmd);
	ASSERT_EQ(1, cmdb[0].opts.size());
	EXPECT_TRUE(cmdb[0].opts["message"].isString());

	ASSERT_EQ(0, eval.set_hostname("test2", 5));
}
