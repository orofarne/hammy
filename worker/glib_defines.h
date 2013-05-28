#pragma once

#include <errno.h>
#include <glib.h>
#include "glib_extra.h"

#ifdef __cplusplus
extern "C" {
#endif

#define ERR_RETURN &lerr
#define _H_AERR GError **error
#define GOTO_END goto END;

#define E_SET_ERRNO(error, name) \
	g_set_error (error, \
		E_DOMAIN, \
		errno, \
		"[%s:%d] " name ": %s", \
		__FILE__, __LINE__, \
		g_strerror (errno))

#define ERRNO_ERR(name) { \
	E_SET_ERRNO (ERR_RETURN, name); \
	GOTO_END \
}

#define FUNC_BEGIN() GError *lerr = NULL;

#define FUNC_END(cleanup) \
	goto END; \
END: \
	if (lerr != NULL) \
	{ \
		cleanup; \
		g_propagate_error (error, lerr); \
		return FALSE; \
	} \
	else \
	{ \
		return TRUE; \
	}

#define H_TRY(code) \
	if (!(code)) \
		GOTO_END

#define H_ASSERT_ERROR \
	if (lerr != NULL) \
		GOTO_END; \

#ifdef __cplusplus
} // extern "C"
#endif
