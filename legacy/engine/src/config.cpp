#include "platforms/platform.hpp"

#include "cvars.hpp"
#include "YamlNode.hpp"

#include <fstream>
#include <iostream>
#include <dirent.h>

using std::string;

string getHomeDirectory() {
	return getenv("HOME");
}

void loadConfigurationFile(const string& filename) {
	ifstream ifs(filename);

	string content ((istreambuf_iterator<char>(ifs) ), 
					 (istreambuf_iterator<char>()));

	auto config = YamlNode::fromString(content);

	for (auto p : config->attributes) {
		cvarSet(p.first, p.second, string("config file: ").append(filename));
	}

	delete(config);
}

void loadHomedirConfigurationFiles() {
	const string configFileDirectory = getHomeDirectory().append("/.greyvar/");

	auto dir = opendir(configFileDirectory.c_str());

	if (dir == nullptr) {
		cout << "Could not find Greyvar home directory." << endl;
		return;
	}

	vector<string> configFiles{};

	for (auto ent = readdir(dir); ent != nullptr; ent = readdir(dir)) {
		if (ent->d_type == DT_REG) {
			configFiles.push_back(configFileDirectory + ent->d_name);
		}
	}

	free(dir);

	for (const auto& configFile : configFiles) {
		loadConfigurationFile(configFile);
	}
}
