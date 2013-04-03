#pragma once

#include <string>
#include <stdexcept>

#define ASSERTPP(expr)                                       \
	if (!(expr)) {                                           \
		std::string msg = "Assertion failed: ";              \
		msg += #expr;                                        \
		throw std::runtime_error(msg);                       \
	}
