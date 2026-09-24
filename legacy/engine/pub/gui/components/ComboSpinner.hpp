#pragma once

#include <vector>

#include "GuiComponent.hpp"

using std::string;
using std::vector;

class ComboSpinner : public GuiComponent {
	public:
		ComboSpinner(const string& title) {
			this->rendererFunc = "combospinner";
			this->title = title;

			this->add("default");
		}

		void add(const string& item);

		const string getTitle() const {
			return this->title;
		}

		const string getValue() const {
			return this->current;
		}

	private:
		string title{};
		string current{};

		vector<string> values{};
};
