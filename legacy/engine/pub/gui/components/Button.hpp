#pragma once

#include "GuiComponent.hpp"

using std::string;

typedef void (*ButtonAction)(void);

class Button: public GuiComponent {
	public:
		SDL_Color color{};

		Button(const string& text, ButtonAction action) : Button(text, {0,0,0, 255}, action) {
		}

		Button(const string& text, SDL_Color color, ButtonAction action) {
			this->text = text;
			this->color = color;
			this->action = action;

			this->rendererFunc = "button";
			this->minimumHeight = 50;
		}

		string getText() const {
			return this->text;
		}

		void onClick() override {
			this->action();
		}

		string toString() const override {
			return "Button {" + this->text + "}";
		}

		ButtonAction action;
	private:
		string text{};
};
