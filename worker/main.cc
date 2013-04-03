#include "worker.hh"

#include <unistd.h>

#include <iostream>
#include <exception>

int main(void) {
	try {
		hammy::Worker w(STDIN_FILENO, STDOUT_FILENO);

		w.run();

		return 0;
	}
	catch(std::exception const &e) {
		std::cerr << "Exception: " << e.what() << std::endl;
	}
	catch(...) {
		std::cerr << "Unknown exception" << std::endl;
	}
	return 1;
}