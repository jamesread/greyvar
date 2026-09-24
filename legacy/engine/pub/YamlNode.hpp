#pragma once

#include <string>
#include <sstream>
#include <map>
#include <vector>

using namespace std;

class YamlNode {
	public: 
		YamlNode* parent = NULL;

		map<string, string> attributes;
		map<string, YamlNode*> children;
		vector<YamlNode*> items;

		YamlNode* child(const string& name);
		YamlNode* attr(const string& name, const string& value);
		YamlNode* attr(const string& name, int value);

		string attr(const string& name);
		uint32_t attri(const string& name);
		bool attrb(const string& name);

		YamlNode* list(const string& title);
		YamlNode* listitem();

		static YamlNode* fromStringstream(stringstream content);
		static YamlNode* fromString(const string& content);
		string toString();
		string toString(int level);
		string toString(int level, bool skipLeading);
};

