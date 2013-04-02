#include <time.h>
#include <gtest/gtest.h>

#include "../eval.hh"

#include <js/jsvalue.h>

#define TIME_X 1364931158

TEST(MozJsTest, SimpleScript) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
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
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
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
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
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
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
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
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
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
	/* TODO
	const char *script2 = "host = 'foo';\n";
	EXPECT_FALSE(
		0 == eval.compile(script, strlen(script))
		&& 0 == eval.exec()
	);
	*/
}

TEST(MozJsTest, Key) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
	eval.set_state(&s);
	const char *script = "cmd('log', {'message': key});\n";
	EXPECT_EQ(0, eval.compile(script, strlen(script)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
	hammy::CmdBuf &cmdb = eval.get_cmdbuf();
	ASSERT_EQ(1, cmdb.size());
	EXPECT_EQ("log", cmdb[0].cmd);
	ASSERT_EQ(1, cmdb[0].opts.size());
	EXPECT_TRUE(cmdb[0].opts["message"].isString());

	ASSERT_EQ(0, eval.set_key("k2", 2));
	/* TODO
	const char *script2 = "key = 'bar';\n";
	EXPECT_FALSE(
		0 == eval.compile(script, strlen(script))
		&& 0 == eval.exec()
	);
	*/
}

TEST(MozJsTest, Timestamp) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
	eval.set_state(&s);
	const char *script = "cmd('log', {'message': timestamp.toString()});\n";
	EXPECT_EQ(0, eval.compile(script, strlen(script)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
	hammy::CmdBuf &cmdb = eval.get_cmdbuf();
	ASSERT_EQ(1, cmdb.size());
	EXPECT_EQ("log", cmdb[0].cmd);
	ASSERT_EQ(1, cmdb[0].opts.size());
	EXPECT_TRUE(cmdb[0].opts["message"].isString());

	ASSERT_EQ(0, eval.set_timestamp(TIME_X + 10));
	/* TODO
	const char *script2 = "key = 'bar';\n";
	EXPECT_FALSE(
		0 == eval.compile(script, strlen(script))
		&& 0 == eval.exec()
	);
	*/
}

TEST(MozJsTest, Value) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
	eval.set_state(&s);
	const char *script = "cmd('log', {'message': value});\n";
	EXPECT_EQ(0, eval.compile(script, strlen(script)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());
	hammy::CmdBuf &cmdb = eval.get_cmdbuf();
	ASSERT_EQ(1, cmdb.size());
	EXPECT_EQ("log", cmdb[0].cmd);
	ASSERT_EQ(1, cmdb[0].opts.size());
	EXPECT_TRUE(cmdb[0].opts["message"].isDouble());

	ASSERT_EQ(0, eval.set_timestamp(TIME_X + 10));
	/* TODO
	const char *script2 = "key = 'bar';\n";
	EXPECT_FALSE(
		0 == eval.compile(script, strlen(script))
		&& 0 == eval.exec()
	);
	*/
}

TEST(MozJsTest, StateSetGet) {
	hammy::MozJSEval eval;
	hammy::State s;
	EXPECT_EQ(0, eval.init());
	ASSERT_EQ(0, eval.set_hostname("test", 4));
	ASSERT_EQ(0, eval.set_key("mykey", 5));
	ASSERT_EQ(0, eval.set_timestamp(TIME_X));
	ASSERT_EQ(0, eval.set_value(js::DoubleValue(3.1415)));
	eval.set_state(&s);

	const char *script =
		"var x = 10;\n"
		"set_state('x', x);\n"
		"if(x === get_state('x')) {\n"
		"    cmd('log', {'message': 'x_ok'});\n"
		"}\n"
		"var y = 'hello';\n"
		"set_state('y', y);\n"
		"if(y === get_state('y')) {\n"
		"    cmd('log', {'message': 'y_ok'});\n"
		"}\n"
		"var z = 3.1415;"
		"if(z === get_state('z')) {\n"
		"    cmd('log', {'message': 'z_ok'});\n"
		"}\n";

	EXPECT_EQ(0, eval.compile(script, strlen(script)));
	EXPECT_EQ(0, eval.exec());
	EXPECT_EQ("", eval.last_error());

	hammy::CmdBuf &cmdb = eval.get_cmdbuf();
	EXPECT_EQ(2, cmdb.size());
	EXPECT_EQ(2, s.size());
}
