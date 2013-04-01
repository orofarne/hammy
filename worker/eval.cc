#include "eval.hh"

#include <stdio.h>

namespace hammy {

JSClass MozJSEval::m_global_class = {
	"global",
	JSCLASS_GLOBAL_FLAGS,
	JS_PropertyStub,
	JS_PropertyStub,
	JS_PropertyStub,
	JS_StrictPropertyStub,
	JS_EnumerateStub,
	JS_ResolveStub,
	JS_ConvertStub,
	NULL,
	JSCLASS_NO_OPTIONAL_MEMBERS
};

MozJSEval::MozJSEval() {

}

MozJSEval::~MozJSEval() {
	JS_DestroyContext(m_cx);
	JS_DestroyRuntime(m_rt);
	JS_ShutDown();
}

int MozJSEval::init() {
	// Create a JS runtime.
	m_rt = JS_NewRuntime(8L * 1024L * 1024L);
	if (m_rt == NULL)
		return 1;

	// Create a context.
	m_cx = JS_NewContext(m_rt, 8192);
	if (m_cx == NULL)
		return 1;
	JS_SetOptions(m_cx, JSOPTION_VAROBJFIX | JSOPTION_METHODJIT);
	JS_SetVersion(m_cx, JSVERSION_LATEST);
	JS_SetErrorReporter(m_cx, MozJSEval::reportError);

	// Create the global object in a new compartment.
	m_global = JS_NewGlobalObject(m_cx, &MozJSEval::m_global_class);
	if (m_global == NULL)
		return 1;

	// Populate the global object with the standard globals, like Object and Array.
	if (!JS_InitStandardClasses(m_cx, m_global))
		return 1;

	return 0;
}

// The error reporter callback.
void MozJSEval::reportError(JSContext *cx, const char *message, JSErrorReport *report) {
	fprintf(stderr, "%s:%u:%s\n",
		report->filename ? report->filename : "<no filename="">",
		(unsigned int) report->lineno,
		message);
}

}