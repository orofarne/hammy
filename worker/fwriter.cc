#include "fwriter.hh"

#include <errno.h>
#include <string.h>

#include <stdexcept>

namespace hammy {

FWriter::FWriter(int fd)
	: m_fp( fdopen(fd, "wb") )
{
}

void FWriter::write(const char* buf, size_t buflen) {
	size_t count = fwrite(buf, buflen, 1, m_fp);
	if(count < 1) {
		throw std::runtime_error(strerror(errno));
	}
}

void FWriter::flush() {
	fflush(m_fp);
}

void FWriter::close() {
	fclose(m_fp);
}

}