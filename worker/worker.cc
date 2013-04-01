#include "worker.hh"

#include <errno.h>

#include <iostream>

namespace hammy {

Worker::Worker(int in_sock, int out_sock)
	: m_in_sock(in_sock)
	, m_out_sock(out_sock)
	, fw(m_out_sock)
	, m_pack(fw)
{
	m_evl.init();
}

Worker::~Worker() {
}

void Worker::socket_readable() {
	m_unpack.reserve_buffer(1024);

	ssize_t count = read(m_in_sock, m_unpack.buffer(), m_unpack.buffer_capacity());

	if(count <= 0) {
		if(count == 0) {
			throw std::runtime_error("connection closed");
		}
		if(errno == EAGAIN || errno == EINTR) {
			return;
		}
		throw std::runtime_error(strerror(errno));
	}

	m_unpack.buffer_consumed(count);

	while(m_unpack.execute()) {
		msgpack::object msg = m_unpack.data();

		auto_zone life( m_unpack.release_zone() );

		m_unpack.reset();

		process_message(msg, life);
	}

	if(m_unpack.message_size() > 10*1024*1024) {
		throw std::runtime_error("message is too large");
	}
}

void Worker::process_message(msgpack::object msg, auto_zone& life) {
	std::cerr << "message reached: " << msg << std::endl;

	m_pack.pack_map(2);
	// CmdBuffer
	m_pack.pack(std::string("CmdBuffer"));
	m_pack.pack_nil();
	// State
	m_pack.pack(std::string("State"));
	m_pack.pack_nil();

	fw.flush();
}

void Worker::run() {
	while(true) {
		socket_readable();
	}
}

}