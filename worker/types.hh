#pragma once

#include <stdint.h>

#include <string>
#include <map>
#include <vector>

#include <js/jsapi.h>
#include <js/jsvalue.h>

namespace hammy {

typedef std::map< std::string, js::Value > Hash;

struct Cmd {
	std::string cmd;
	Hash opts;
};

typedef std::vector< Cmd > CmdBuf;

struct StateElem {
	uint64_t LastUpdate;
	js::Value Value;
};

typedef std::map< std::string, StateElem > State;

}