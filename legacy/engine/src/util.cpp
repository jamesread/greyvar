#include <fstream>
#include <sstream>
#include <iostream>
#include <unistd.h>
#include <uuid++.hh>

#include "cvars.hpp"

using std::string, std::stringstream, std::ifstream;
using std::cout;
using std::endl;

stringstream readFile(const string& filename) {
	string line;
	stringstream out; 	
	ifstream inFile;
	
	inFile.open(filename);

	while (getline(inFile, line)) {
		cout << "read line: " << line << endl;

		out << line << endl;
	}

	inFile.close();

	return out;
}

void parseArguments(int argc, char* argv[]) {
	enum {
		ANY,
		CVAR,
	} nextArgumentExpected = ANY;

	string last{};

	for (int i = 0; i < argc; i++) {
		string current = argv[i];

		switch (nextArgumentExpected) {
			case ANY:	
				if (current[0] == '-') {
					nextArgumentExpected = CVAR;
				}

				if (current[0] == '+') {
					cvarSet(current.substr(1), "1", "command line, +1 syntax");
					nextArgumentExpected = ANY;
				}

				break;
			case CVAR:
				cvarSet(last.substr(1), current, "command line");

				nextArgumentExpected = ANY;
				break;
		}

		last = current;
	}
}

string uuidGen() {
	uuid id;
	id.make(UUID_MAKE_V1);

	return id.string();
}
