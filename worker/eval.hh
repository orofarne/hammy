#pragma once

#include <js/jsapi.h>

namespace hammy {

class MozJSEval {
	public:
		MozJSEval();
		~MozJSEval();

		int init();

	private:
		// The class of the global object.
		static JSClass m_global_class;
		// The error reporter callback.
		static void reportError(JSContext *cx, const char *message, JSErrorReport *report);

	private:
		// JS variables.
		JSRuntime *m_rt;
		JSContext *m_cx;
		JSObject *m_global;
};

}