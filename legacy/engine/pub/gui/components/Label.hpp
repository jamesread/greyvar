#pragma once

#include <string>

#include "GuiComponent.hpp"

using namespace std;

class Label: public GuiComponent {
	public:
		Label(const string& text) : Label(text, 24) {}

		Label(const string& text, int fontSize, bool bold, bool underlined, bool background) : Label(text, fontSize) {
			this->bold = bold;
			this->underlined = underlined;
			this->background = background;
		}

		Label(const string& text, int fontSize) {
			this->rendererFunc = "label";
			this->text = text;

			this->minimumHeight = this->fontSize * 2.6;
			this->minimumWidth = this->text.size() * 18;
			this->fontSize = fontSize;
		}

		string getText() {
			return this->text;
		}

		string toString() const override {
			return "Label {" + this->text + "}";
		}

		int fontSize = 24;
		bool bold = false;
		bool underlined = false;
		bool background = true;

	private: 
		string text;

};
