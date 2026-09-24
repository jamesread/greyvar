#pragma once 

#include "Button.hpp"

typedef void (*MenuItemAction)(void);

class MenuItem : public GuiComponent {
	public:
		explicit MenuItem(std::string text) : MenuItem(text, nullptr) {}

		MenuItem(std::string text, MenuItemAction action) {
			this->text = text;
			this->action = action;
		}

		std::string text;
		MenuItemAction action;

		string toString() const override {
			return "MenuItem {" + this->text + "}";
		}

};

