package hammy

import (
	. "launchpad.net/gocheck"
)

type PluginTestSuite struct{}

var _ = Suite(&PluginTestSuite{})

func (s *PluginTestSuite) TestPluginAddGet(c *C) {
	p := &Plugin{}
	AddPlugin("test_plugin", p)
	p2 := GetPlugin("test_plugin")
	c.Check(p2, Equals, p)
}

func (s *PluginTestSuite) TestPluginAddPanic(c *C) {
	p1 := &Plugin{}
	p2 := &Plugin{}
	AddPlugin("test_plugin_1", p1)
	AddPlugin("test_plugin_2", p2)

	c.Assert(
		func() {
			p1 := &Plugin{}
			p2 := &Plugin{}
			AddPlugin("test_plugin_3", p1)
			AddPlugin("test_plugin_3", p2)
		}, PanicMatches, ".*already exists")
}

func (s *PluginTestSuite) TestPluginGetPanic(c *C) {
	p := &Plugin{}
	AddPlugin("test_plugin_7", p)
	p2 := GetPlugin("test_plugin_7")
	c.Assert(p2, Equals, p)

	c.Assert(
		func() {
			_ = GetPlugin("test_plugin_8")
		}, PanicMatches, ".*does not exist")
}
