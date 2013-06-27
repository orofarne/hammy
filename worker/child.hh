#pragma once

#include "fwriter.hh"
#include "eval.hh"

#include <stdint.h>
#include <msgpack.hpp>
#include <ev.h>

namespace hammy {

class Worker {
	public:
		Worker(struct ev_loop *loop, int in_sock, int out_sock);
		~Worker();

		void run();

	private:
		typedef std::auto_ptr<msgpack::zone> auto_zone;

		void socket_readable();
		void process_message(msgpack::object msg, auto_zone& life);
		void read_state(msgpack::object *obj);
		void write_state();
		void write_cmdbuf();
		void process_data(msgpack::object *obj);

		static void read_cb (struct ev_loop *loop, ev_io *w, int revents);

	private:
		struct ev_loop *m_loop;
		ev_io m_io;
		int m_in_sock;
		int m_out_sock;
		uint64_t m_id;
		msgpack::unpacker m_unpack;
		FWriter fw;
		msgpack::packer<FWriter> m_pack;
		MozJSEval m_evl;

		// Request
		State m_state;
};

}
