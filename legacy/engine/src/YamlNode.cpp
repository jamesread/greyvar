#include "YamlNode.hpp"

#include <sstream>
#include <iostream>

#include "util.hpp"

using std::string;

YamlNode* YamlNode::attr(const string& name, const int value) {
	this->attributes[name] = to_string(value);

	return this;
}

YamlNode* YamlNode::attr(const string& name, const string& value) {
	if (this->items.size() > 0) {
		throw std::invalid_argument("Cannot add attributes, as this node is already a list");
	}

	this->attributes[name] = value;

	return this;
}

string YamlNode::attr(const string& name) {
	return this->attributes[name];
}

uint32_t YamlNode::attri(const string& name) {
	string s = this->attributes[name];

	try {
		return stoul(s.c_str());
	} catch (...) {
		cout << "Yaml, number expected, got >" << s << "<" << endl;

		return 0;
	}
}

bool YamlNode::attrb(const string& name) {
	if (this->attributes[name] == "true") {
		return true;
	} else if (this->attributes[name] == "false") {
		return false;
	} else {
		cout << "wtf" << name << endl;
		return false;
	}
}

YamlNode* YamlNode::child(const string& name) {
	if (this->items.size() > 0) {
		throw std::invalid_argument("Cannot add child, as this node is already a list");
	}

	auto child = new YamlNode();
	child->parent = this;

	this->children[name] = child;

	return child;
}

string YamlNode::toString() {
	return this->toString(0);
}

YamlNode* YamlNode::list(const string& title) {
	auto i = new YamlNode();

	this->children[title] = i;

	return i;
}

YamlNode* YamlNode::listitem() {
	if (this->attributes.size() > 0) {
		throw std::invalid_argument("Cannot add listitems, as this node already has attributes");
	}

	auto i = new YamlNode();

	this->items.push_back(i);

	return i;
}

string YamlNode::toString(const int level) {
	return this->toString(level, false);
}

string YamlNode::toString(const int level, const bool skipLeading) {
	stringstream out;
	bool outputStarted = false;

	for (auto pair : this->attributes) {
		if (outputStarted || !skipLeading) {
			out << string(level * 2, ' ');
		}
		
		out << pair.first << ": " << pair.second << "\n";

		outputStarted = true;
	}

	for (auto pair : this->children) {
		out << string(level * 2, ' ') << pair.first << ": \n";
		out << pair.second->toString(level + 1);
	}

	for (auto i : this->items) {
		out << string(level * 2, ' ') << "- " << i->toString(level + 1, true);
	}

	return out.str();
}

struct YamlParsedLine {
	int indentation = 0;
	string key;
	string val;
	bool startOfList = false;
};

YamlParsedLine* parseLine(string line) {
	auto ret = new YamlParsedLine();

	bool contentStarted = false;
	bool keyFound = false;

	for (auto c : line) {
		switch (c) {
			case ' ':
				if (!contentStarted) {
					ret->indentation++;
				} else {
					ret->val += ' ';
				}

				break;
			case '-':
				if (!contentStarted) {
					ret->startOfList = true;
					contentStarted = true;
				}

				break;
			case ':':
				keyFound = true;

				break;
			default:
				if (!contentStarted) {
					contentStarted = true;
				}

				if (keyFound) {
					ret->val += c;
				} else {
					ret->key += c;
				}
		}
	}
	
	ret->indentation /= 2;
	trim(ret->key);
	trim(ret->val);

	return ret;
}

YamlNode* YamlNode::fromStringstream(stringstream content) {
	string line;

	content.clear();
	content.seekg(0, ios::beg);

	//cout << "parse" << endl;

	YamlNode* root = new YamlNode;
	auto current = root;

	int currentIndentation = 0;
	bool inList = false;

	YamlParsedLine* pline;

	while (getline(content, line)) {
		pline = parseLine(line);

		//cout << "parsed. indentation: " << pline->indentation << "  k: " << pline->key << " v: " << pline->val << endl;

		if (pline->indentation != currentIndentation) {
			inList = false;
		}

		if (pline->indentation > currentIndentation) {
			if (pline->startOfList) {
				inList = true;
			}

		} else if (pline->indentation < currentIndentation) {
			for (int i = 0; i < (currentIndentation - pline->indentation); i++) {
				current = current->parent;	
				//cout << "back" << endl;
			}
		}

		if (inList) {
			YamlNode* li = current->listitem();

			if (pline->val.size() == 0) {
				li->child(pline->key);
			} else {
				li->attr(pline->key, pline->val);
			}
		} else {
			if (pline->val.size() == 0) {
				current = current->child(pline->key);
			} else {
				current->attr(pline->key, pline->val);
			}
		}

		currentIndentation = pline->indentation;
	
		delete(pline);
	}


	//cout << "Finished parsing." << endl << "----------" << endl;

	return root;
}

YamlNode* YamlNode::fromString(const string& content) {
	return YamlNode::fromStringstream(stringstream(content));
}

