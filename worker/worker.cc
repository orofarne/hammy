#include "worker.hh"
#include "asserts.hh"
#include "converters.hh"
#include "types.hh"

#include <errno.h>
#include <stdint.h>

#include <iostream>
#include <stdexcept>

namespace hammy {

Worker::Worker(int in_sock, int out_sock)
	: m_in_sock(in_sock)
	, m_out_sock(out_sock)
	, fw(m_out_sock)
	, m_pack(fw)
{
	ASSERTPP(m_evl.init() == 0);
}

Worker::~Worker() {
}

void Worker::socket_readable() {
	m_unpack.reserve_buffer(1024);

	ssize_t count = read(m_in_sock, m_unpack.buffer(), m_unpack.buffer_capacity());

	std::cerr << count << std::endl; // FIXME

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
	msgpack::object *Key = NULL;
	msgpack::object *Trigger = NULL;
	msgpack::object *State = NULL;
	msgpack::object *IData = NULL;

	ASSERTPP(msg.type == msgpack::type::MAP);
	for(uint32_t i = 0; i < msg.via.map.size; ++i) {
		msgpack::object_kv &kv = msg.via.map.ptr[i];
		ASSERTPP(kv.key.type == msgpack::type::RAW);

		if(0 == strncmp(kv.key.via.raw.ptr, "Key", 3)) {
			ASSERTPP(kv.val.type == msgpack::type::RAW);
			Key = &kv.val;
		} else if(0 == strncmp(kv.key.via.raw.ptr, "Trigger", 7)) {
			ASSERTPP(kv.val.type == msgpack::type::RAW);
			Trigger = &kv.val;
		} else if (0 == strncmp(kv.key.via.raw.ptr, "State", 5)) {
			State = &kv.val;
		} else if (0 == strncmp(kv.key.via.raw.ptr, "IData", 5)) {
			IData = &kv.val;
		}
	}
	ASSERTPP(Key && Trigger && State && IData);

	std::string key(Key->via.raw.ptr, Key->via.raw.size);
	std::string trigger(Trigger->via.raw.ptr, Trigger->via.raw.size);

	int r = m_evl.eval(trigger.c_str());
	if (r != 0)
		throw std::runtime_error(m_evl.last_error());

	m_pack.pack_map(2);

	// CmdBuffer
	CmdBuf &cmdb = m_evl.get_cmdbuf();
	m_pack.pack(std::string("CmdBuffer"));
	m_pack.pack_map(cmdb.size());
	for(CmdBuf::const_iterator it = cmdb.begin(); it != cmdb.end(); ++it) {
		m_pack.pack(it->cmd);
		m_pack.pack_map(it->opts.size());
		for(Hash::const_iterator jt = it->opts.begin(); jt != it->opts.end(); ++jt) {
			m_pack.pack(jt->first);
			pack_jsval(&m_pack, jt->second);
		}
	}

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