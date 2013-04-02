#pragma once

#include "types.hh"

#include <string>
#include <sstream>

#include <js/jsapi.h>

namespace hammy {

class MozJSEval {
	public:
		MozJSEval();
		~MozJSEval();

		int init();
		int set_hostname(const char *name, size_t len);
		void set_state(State *state);

		int compile(const char *script, size_t len);
		int exec();
		std::string last_error();
		CmdBuf &get_cmdbuf();

	private:
		// The error reporter callback.
		static void reportError(JSContext *cx, const char *message, JSErrorReport *report);

		static JSBool cmd(JSContext *cx, uintN argc, jsval *vp);

	private:
		// JS variables.
		JSRuntime *m_rt;
		JSContext *m_cx;
		JSObject *m_global;
		JSObject *m_script;

		std::ostringstream m_error;
		CmdBuf m_cmdbuf;
		State *m_state;

		// The class of the global object.
		static JSClass m_global_class;
		static JSFunctionSpec m_global_functions[];
		static MozJSEval *m_instance;
};

}