#pragma once

#include "asserts.hh"

#include <stdint.h>

#include <stdexcept>
#include <sstream>
#include <limits>

#include <js/jsapi.h>
#include <js/jsvalue.h>

namespace hammy {

/********************************************************************************************************************
 * JS type   * JS::Value type tests * Constructor                   * Accessors         * Mutators                  *
 ********************************************************************************************************************
 * null      * val.isNull()         * NullValue()                   * N/A               * val.setNull()             *
 ********************************************************************************************************************
 * undefined * val.isUndefined()    * UndefinedValue()              * N/A               * val.setUndefined()        *
 ********************************************************************************************************************
 * boolean   * val.isBoolean(),     * BooleanValue(bool)            * val.toBoolean()   * val.setBoolean(bool)      *
 *           * val.isTrue(),        *                               *                   *                           *
 *           * val.isFalse()        *                               *                   *                           *
 ********************************************************************************************************************
 * number    * val.isInt32(),       * NumberValue(any number type), * val.toInt32(),    * val.setInt32(int32_t),    *
 *           * val.isDouble(),      * Int32Value(int32_t),          * value.toDouble(), * val.setNumber(uint32_t),  *
 *           * val.isNumber()       * DoubleValue(double)           * val.toNumber()    * val.setNumber(double)     *
 ********************************************************************************************************************
 * string    * val.isString()       * StringValue(JSString *)       * val.toString()    * val.setString(JSString *) *
 ********************************************************************************************************************
 * object    * val.isObject()       * ObjectValue(JSObject &),      *                   *                           *
 *           *                      * ObjectOrNullValue(JSObject *) * val.toObject()    * val.setObject(JSObject &) *
 ********************************************************************************************************************/

/*
packer< Stream > & 	pack_uint8 (uint8_t d)
packer< Stream > & 	pack_uint16 (uint16_t d)
packer< Stream > & 	pack_uint32 (uint32_t d)
packer< Stream > & 	pack_uint64 (uint64_t d)
packer< Stream > & 	pack_int8 (int8_t d)
packer< Stream > & 	pack_int16 (int16_t d)
packer< Stream > & 	pack_int32 (int32_t d)
packer< Stream > & 	pack_int64 (int64_t d)
packer< Stream > & 	pack_fix_uint8 (uint8_t d)
packer< Stream > & 	pack_fix_uint16 (uint16_t d)
packer< Stream > & 	pack_fix_uint32 (uint32_t d)
packer< Stream > & 	pack_fix_uint64 (uint64_t d)
packer< Stream > & 	pack_fix_int8 (int8_t d)
packer< Stream > & 	pack_fix_int16 (int16_t d)
packer< Stream > & 	pack_fix_int32 (int32_t d)
packer< Stream > & 	pack_fix_int64 (int64_t d)
packer< Stream > & 	pack_short (short d)
packer< Stream > & 	pack_int (int d)
packer< Stream > & 	pack_long (long d)
packer< Stream > & 	pack_long_long (long long d)
packer< Stream > & 	pack_unsigned_short (unsigned short d)
packer< Stream > & 	pack_unsigned_int (unsigned int d)
packer< Stream > & 	pack_unsigned_long (unsigned long d)
packer< Stream > & 	pack_unsigned_long_long (unsigned long long d)
packer< Stream > & 	pack_float (float d)
packer< Stream > & 	pack_double (double d)
packer< Stream > & 	pack_nil ()
packer< Stream > & 	pack_true ()
packer< Stream > & 	pack_false ()
packer< Stream > & 	pack_array (unsigned int n)
packer< Stream > & 	pack_map (unsigned int n)
packer< Stream > & 	pack_raw (size_t l)
packer< Stream > & 	pack_raw_body (const char *b, size_t l)
*/

template<typename P>
void pack_jsval(JSContext *cx, P *packer, js::Value const &val) {
	if(val.isNull() || val.isUndefined()) {
		packer->pack_nil();
	} else if(val.isTrue()) {
		packer->pack_true();
	} else if(val.isFalse()) {
		packer->pack_false();
	} else if(val.isInt32()) {
		packer->pack_int32(val.toInt32());
	} else if(val.isDouble()) {
		packer->pack_double(val.toDouble());
	} else if(val.isString()) {
		JSString *str = val.toString();
		size_t N = JS_GetStringLength(str);
		packer->pack_raw(N);
		packer->pack_raw_body(JS_EncodeString(cx, str), N);
	} else if(val.isObject() && JS_TRUE == JS_ObjectIsDate(cx, &val.toObject())) {
		jsval rval;
		if(!JS_CallFunctionName(cx, &val.toObject(), "getTime", 0, NULL, &rval))
			throw std::runtime_error("JS_CallFunctionName");
		ASSERTPP(val.isDouble());
		packer->pack_double(val.toDouble() / 1000.0);
	} else {
		throw std::runtime_error("Can't pack argument");
	}
}



/* enum object_type {
 *     NIL                     = MSGPACK_OBJECT_NIL,
 *     BOOLEAN                 = MSGPACK_OBJECT_BOOLEAN,
 *     POSITIVE_INTEGER        = MSGPACK_OBJECT_POSITIVE_INTEGER,
 *     NEGATIVE_INTEGER        = MSGPACK_OBJECT_NEGATIVE_INTEGER,
 *     DOUBLE                  = MSGPACK_OBJECT_DOUBLE,
 *     RAW                     = MSGPACK_OBJECT_RAW,
 *     ARRAY                   = MSGPACK_OBJECT_ARRAY,
 *     MAP                     = MSGPACK_OBJECT_MAP,
 * };
 *
 * union union_type {
 *     bool boolean;
 *     uint64_t u64;
 *     int64_t  i64;
 *     double   dec;
 *     object_array array;
 *     object_map map;
 *     object_raw raw;
 *     object_raw ref;  // obsolete
 * };
 */

void unpack_jsval(JSContext *cx, js::Value &val, msgpack::object &obj) {
	switch(obj.type) {
		case msgpack::type::NIL:
			val.setNull();
			break;
		case msgpack::type::POSITIVE_INTEGER:
			if (obj.via.u64 < std::numeric_limits<int32_t>::max())
				val.setInt32(obj.via.u64);
			else
				val.setNumber(double(obj.via.u64));
			break;
		case msgpack::type::NEGATIVE_INTEGER:
			if (obj.via.i64 > std::numeric_limits<int32_t>::min()
				&& obj.via.i64 < std::numeric_limits<int32_t>::max())
				val.setInt32(obj.via.i64);
			else
				val.setNumber(double(obj.via.i64));
			break;
		case msgpack::type::DOUBLE:
			val.setNumber(obj.via.dec);
			break;
		case msgpack::type::RAW:
			{
				JSString *str = JS_NewStringCopyN(cx, obj.via.raw.ptr, obj.via.raw.size);
				val.setString(str);
			}
			break;
		default:
			{
				std::ostringstream msg;
				msg << "Object as argument (type " << obj.type << ')';
				throw std::runtime_error(msg.str());
			}
	}
}

}
