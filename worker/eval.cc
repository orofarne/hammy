#include "eval.hh"
#include "asserts.hh"

#include <stdio.h>
#include <assert.h>
#include <time.h>

namespace hammy {

JSClass MozJSEval::m_global_class = {
	"global", JSCLASS_GLOBAL_FLAGS,
	JS_PropertyStub, JS_PropertyStub, JS_PropertyStub, JS_StrictPropertyStub,
	JS_EnumerateStub, JS_ResolveStub, JS_ConvertStub, JS_FinalizeStub,
	JSCLASS_NO_OPTIONAL_MEMBERS
};

JSFunctionSpec MozJSEval::m_global_functions[] = {
	JS_FS("cmd",		MozJSEval::cmd,			2,	0),
	JS_FS("get_state",	MozJSEval::get_state,	1,	0),
	JS_FS("set_state",	MozJSEval::set_state,	2,	0),
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
		return 1;

	return 0;
}

int MozJSEval::set_hostname(const char *name, size_t len) {
	JSString *str = JS_NewStringCopyN(m_cx, name , len);
	if (!JS_DefineProperty(m_cx, m_global, "host", STRING_TO_JSVAL(str), NULL, NULL, JSPROP_READONLY))
		return 1;
	return 0;
}

int MozJSEval::set_key(const char *key, size_t len) {
	JSString *str = JS_NewStringCopyN(m_cx, key , len);
	if (!JS_DefineProperty(m_cx, m_global, "key", STRING_TO_JSVAL(str), NULL, NULL, JSPROP_READONLY))
		return 1;
	return 0;
}

int MozJSEval::set_timestamp(uint64_t ts) {
	double msec = ts / 1000.0;
	JSObject *date = JS_NewDateObjectMsec(m_cx, msec);
	if (!JS_DefineProperty(m_cx, m_global, "timestamp", OBJECT_TO_JSVAL(date), NULL, NULL, JSPROP_READONLY))
		return 1;
	return 0;
}

int MozJSEval::set_value(js::Value val) {
	if (!JS_DefineProperty(m_cx, m_global, "value", js::Jsvalify(val), NULL, NULL, JSPROP_READONLY))
		return 1;
	return 0;
}

void MozJSEval::set_state(State *state) {
	m_state = state;
}

int MozJSEval::compile(const char *script, size_t len) {
	jsval rval;
	JSString *str;
	JSBool ok;
	const char *filename = "trigger";
	uintN lineno = 0;

	m_cmdbuf.clear();
	m_error.str( std::string() );
	m_error.clear();

	m_script = JS_CompileScript(m_cx, m_global, script, len, filename, lineno);
	if (m_script == NULL)
		return 1;

	return 0;
}

int MozJSEval::exec() {
	jsval result;

	if (!JS_ExecuteScript(m_cx, JS_GetGlobalObject(m_cx), m_script, &result))
		return 1;

	JS_MaybeGC(m_cx);

	return 0;
}

CmdBuf &MozJSEval::get_cmdbuf() {
	return m_cmdbuf;
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
	JSString* name_raw;
	JSObject* opts_raw;

	char *name;

	if(!JS_ConvertArguments(cx, argc, JS_ARGV(cx, vp), "S/o", &name_raw, &opts_raw))
		return JS_FALSE;

	name = JS_EncodeString(cx, name_raw);

	m_instance->m_cmdbuf.resize(m_instance->m_cmdbuf.size() + 1);
	Cmd &cmd = m_instance->m_cmdbuf.back();
	cmd.cmd = name;

	if(opts_raw != NULL) {
		JSIdArray *ids = JS_Enumerate(cx, opts_raw);
		for (jsint i = 0; i < ids->length; i++) {
			jsid id = ids->vector[i];
			jsval id_val;
			assert(JS_TRUE == JS_IdToValue(cx, id, &id_val));
			JSString *id_raw = JS_ValueToString(cx , id_val);
			char *id_str = JS_EncodeString(cx, id_raw);
			jsval val_val;
			assert(JS_TRUE == JS_GetPropertyById(cx, opts_raw, id, &val_val));

			cmd.opts[id_str] = js::Valueify(val_val);

			JS_free(cx, id_str);
		}
	}

	JS_free(cx, name);

	JS_SET_RVAL(cx, vp, JSVAL_VOID); // return undefined
	return JS_TRUE;
}

JSBool MozJSEval::get_state(JSContext *cx, uintN argc, jsval *vp) {
	JSString* key_raw;

	if(!JS_ConvertArguments(cx, argc, JS_ARGV(cx, vp), "S", &key_raw))
		return JS_FALSE;

	char *key_str = JS_EncodeString(cx, key_raw);

	State::const_iterator it = m_instance->m_state->find(key_str);
	if (it == m_instance->m_state->end()) {
		JS_SET_RVAL(cx, vp, JSVAL_NULL);
	} else {
		JS_SET_RVAL(cx, vp, js::Jsvalify(it->second.Value));
	}

	return JS_TRUE;
}

JSBool MozJSEval::set_state(JSContext *cx, uintN argc, jsval *vp) {
	JSString* key_raw;

	if(!JS_ConvertArguments(cx, argc, JS_ARGV(cx, vp), "S*", &key_raw))
		return JS_FALSE;

	char *key_str = JS_EncodeString(cx, key_raw);
	jsval val = vp[3];

	StateElem se;
	se.LastUpdate = time(NULL);
	se.Value = js::Valueify(val);
	(*m_instance->m_state)[key_str] = se;

	JS_SET_RVAL(cx, vp, JSVAL_VOID); // return undefined
	return JS_TRUE;
}

}