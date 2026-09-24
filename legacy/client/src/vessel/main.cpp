#include <iostream>
#include <sys/stat.h>

using std::cout;
using std::endl;

int mainGreyvarCore(int argc, char* argv[]);

bool directoryExists(const char* path) {
	bool exists = false;
	struct stat statResult{};

	if (path != nullptr) {
		if (stat(path, &statResult) == 0 && S_ISDIR(statResult.st_mode)) {
			exists = true;
		}
	}

	return exists;
}

int main(int argc, char* argv[]) {
	cout << "Greyvar (vessel) " << endl << "----------------" << endl;

	if (!directoryExists("res")) {
		cout << "res directory not found." << endl;
	} else {
		cout << "Vessel checks complete. Starting the core..." << endl << endl;
		mainGreyvarCore(argc, argv);
	}

	return 0;
}
