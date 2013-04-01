#include "eval.hh"
#include "asserts.hh"

#include <stdio.h>
#include <assert.h>

namespace hammy {

JSClass MozJSEval::m_global_class = {
	"global", JSCLASS_GLOBAL_FLAGS,
	JS_PropertyStub, JS_PropertyStub, JS_PropertyStub, JS_StrictPropertyStub,
	JS_EnumerateStub, JS_ResolveStub, JS_ConvertStub, JS_FinalizeStub,
	JSCLASS_NO_OPTIONAL_MEMBERS
};

JSFunctionSpec MozJSEval::m_global_functions[] = {
	JS_FS("cmd",	MozJSEval::cmd,			2,	0),
	JS_FS_END
};

MozJSEval *MozJSEval::m_instance = NULL;

MozJSEval::MozJSEval() {
	ASSERTPP(m_instance == NULL);
	m_instance = this;
}

MozJSEval::~MozJSEval() {
	JS_DestroyContext(m_cx);
	JS_DestroyRuntime(m_rt);
	JS_ShutDown();
	m_instance = NULL;
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
	JS_SetOptions(m_cx, JSOPTION_VAROBJFIX | JSOPTION_JIT | JSOPTION_METHODJIT);
	JS_SetVersion(m_cx, JSVERSION_LATEST);
	JS_SetErrorReporter(m_cx, MozJSEval::reportError);

	// Create the global object in a new compartment.
	m_global = JS_NewCompartmentAndGlobalObject(m_cx, &MozJSEval::m_global_class, NULL);
	if (m_global == NULL)
		return 1;

	// Populate the global object with the standard globals, like Object and Array.
	if (!JS_InitStandardClasses(m_cx, m_global))
		return 1;

	if (!JS_DefineFunctions(m_cx, m_global, m_global_functions))
		return 2;

	return 0;
}

int MozJSEval::eval(const char *script) {
	jsval rval;
	JSString *str;
	JSBool ok;
	const char *filename = "noname";
	uintN lineno = 0;

	m_cmdbuf.clear();
	m_error.str( std::string() );
	m_error.clear();

	ok = JS_EvaluateScript(m_cx, m_global, script, strlen(script), filename, lineno, &rval);
	if (!ok)
		return 1;

	return 0;
}

CmdBuf *MozJSEval::get_cmdbuf() {
	return &m_cmdbuf;
}

// The error reporter callback.
void MozJSEval::reportError(JSContext *cx, const char *message, JSErrorReport *report) {
	m_instance->m_error
		<< (report->filename ? report->filename : "<no filename=\"filename\">") << ':'
		<< report->lineno << ": " << message << "\n"
		;
}

std::string MozJSEval::last_error() {
	return m_error.str();
}

/************************************************************************
 *                           Worker API                                 *
 ************************************************************************/

JSBool MozJSEval::cmd(JSContext *cx, uintN argc, jsval *vp) {
	if(argc == 0)
		return JS_FALSE;

	JSString* name_raw;
	JSObject* opts_raw;

	char *name;

	if(!JS_ConvertArguments(cx, argc, JS_ARGV(cx, vp), "S/o", &name_raw, &opts_raw))
		return JS_FALSE;

	name = JS_EncodeString(cx, name_raw);

	if(opts_raw != NULL) {
		js::Value opts;
		opts.setObject(*opts_raw);

		// TODO
	}

	JS_free(cx, name);

	JS_SET_RVAL(cx, vp, JSVAL_VOID); // return undefined
	return JS_TRUE;
}

}