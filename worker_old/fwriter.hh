#pragma once

#include <stdio.h>

namespace hammy {

class FWriter {
	public:
		FWriter(int fd);

		void write(const char* buf, size_t buflen);
		void flush();
		void close();

	private:
		FILE* m_fp;
};

}