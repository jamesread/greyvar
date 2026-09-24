#pragma once

#include <map>
#include <string>

using std::map;
using std::string;
using std::wstring;

void cvarInit();

void cvarSet(const string& key, const string& val, string source);
void cvarSet(const string& key, const string& val);
void cvarSetb(const string& key, const bool& val);
void cvarSeti(const string& key, const int& val);
void cvarSeti(const string& key, const int& val, const string& from);

string cvarGet(const string& key);
wstring cvarGetW(const wstring& key);
bool cvarGetb(const string& key);
int cvarGeti(const string& key);
int cvarGeti(const string& key, const int& def);

bool cvarIsset(const string& key);

