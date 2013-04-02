#pragma once

#include "asserts.hh"

#include <stdexcept>

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
	} else if(val.isObject()) {
		throw std::runtime_error("Object as argument");
	}
}

}