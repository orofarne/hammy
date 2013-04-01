#include "fwriter.hh"

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

	private:
		int m_in_sock;
		int m_out_sock;
		msgpack::unpacker m_unpack;
		FWriter fw;
		msgpack::packer<FWriter> m_pack;
};

}