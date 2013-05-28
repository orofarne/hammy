#include "fwriter.hh"

#include <errno.h>
#include <string.h>

#include <stdexcept>
#include <sstream>

namespace hammy {

FWriter::FWriter(int fd)
	: m_fp( fdopen(fd, "w") )
{
}

void FWriter::write(const char* buf, size_t buflen) {
	if(buflen == 0)
		return;
	ssize_t count = fwrite(buf, buflen, 1, m_fp);
	if(count == 0) {
		std::ostringstream msg;
		msg << "connection closed ("
		    << strerror(errno) << "), bytes to write: " << buflen;
		throw std::runtime_error(msg.str());
	}
	if(count < 0) {
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
