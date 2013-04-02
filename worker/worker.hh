#pragma once

#include "fwriter.hh"
#include "eval.hh"

#include <msgpack.hpp>

namespace hammy {

class Worker {
	public:
		Worker(int in_sock, int out_sock);
		~Worker();

		void run();

	private:
		typedef std::auto_ptr<msgpack::zone> auto_zone;

		void socket_readable();
		void process_message(msgpack::object msg, auto_zone& life);
		void read_state(msgpack::object *obj);
		void write_state();
		void write_cmdbuf();

	private:
		int m_in_sock;
		int m_out_sock;
		msgpack::unpacker m_unpack;
		FWriter fw;
		msgpack::packer<FWriter> m_pack;
		MozJSEval m_evl;

		// Request
		State m_state;
};

}