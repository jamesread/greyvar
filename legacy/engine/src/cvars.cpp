#include <iostream>
#include <map>

#include "cvars.hpp"

using std::cout, std::endl, std::to_string;


map<string, string> cvars;

bool cvarIsset(const string& key) {
	return ::cvars.count(key) == 1;
}

void cvarSet(const string& key, const string& val, string source) {
	if (!source.empty()) {
		source = "(from " + source + ") ";
	}

	cout << "Set cvar " << source << key << " = " << val << endl;

	::cvars[key] = val;
}

void cvarSet(const string& key, const string& val) {
	cvarSet(key, val, "");
}

void cvarSetb(const string& key, const bool& val, const string& source) {
	if (val) {
		cvarSet(key, "1", source);
	} else {
		cvarSet(key, "0", source);
	}
}

void cvarSetb(const string& key, const bool& val) {
	cvarSetb(key, val, "");
}

string cvarGet(const string& key) {
	if (cvarIsset(key)) {
		return ::cvars[key];
	} else {
		return "";
	}
}

wstring cvarGetW(const string& key) {
	if (cvarIsset(key)) {
		auto s = ::cvars[key];

		wstring wstmp(s.begin(), s.end());

		return wstmp;
	} else {
		return L"";
	}
}

bool cvarGetb(const string& key) {
	string s = cvarGet(key);

	if (s.empty()) {
		return false;
	}

	try {
		return (bool) stoi(s);
	} catch (...) {
		cout << "cvar, number expected, got >" << s << "<" << endl;
		return false;
	}
}

int cvarGeti(const string& key) {
	return cvarGeti(key, 0);
}

int cvarGeti(const string& key, const int& def) {
	string i = cvarGet(key);

	if (i.empty()) {
		return def;
	} else {
		return stoi(cvarGet(key));
	}
}

void cvarSeti(const string& key, const int& val) {
	cvarSeti(key, val, "");
}

void cvarSeti(const string& key, const int& val, const string& from) {
	cvarSet(key, to_string(val), from);
}

void cvarInit() {
	string def = "defaults";

	cvarSet("nickname", "unamed.player", def);
	cvarSetb("bind_keyboard", true, def);
	cvarSetb("r_scale_grid", true, def);
	cvarSeti("snd_channel_ui_volume", 100, def);
	cvarSet("lp_1_username", "Player 1", def);
}
