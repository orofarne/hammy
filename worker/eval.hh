#pragma once

#include <string>
#include <sstream>

#include <js/jsapi.h>

namespace hammy {

class MozJSEval {
	public:
		MozJSEval();
		~MozJSEval();

		int init();

		int eval(const char *script);
		std::string last_error();

	private:
		// The error reporter callback.
		static void reportError(JSContext *cx, const char *message, JSErrorReport *report);

	private:
		// JS variables.
		JSRuntime *m_rt;
		JSContext *m_cx;
		JSObject *m_global;

		// The class of the global object.
		static JSClass m_global_class;

		static std::ostringstream m_error;
		static bool m_constructed;
};

}