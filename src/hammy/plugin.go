package hammy

import (
	"fmt"
	"encoding/json"
)

// Plugin factory struct
type Plugin struct {
	create func(json.RawMessage) interface{}
}

var plugin_map map[string]*Plugin

// Add plugin to global repository. Should be called from module's init()
// Panics on name collapse
func AddPlugin(name string, plugin *Plugin) {
	if plugin_map == nil {
		plugin_map = make(map[string]*Plugin)
	}

	_, ok := plugin_map[name]
	if ok {
		panic(fmt.Sprintf("plugin '%v' already exists", name))
	}
	plugin_map[name] = plugin
}

// Get plugin from global repository
// Panics if plugin does not exist
func GetPlugin(name string) *Plugin {
	v, ok := plugin_map[name]
	if !ok {
		panic(fmt.Sprintf("plugin '%v' does not exist", name))
	}
	return v
}
