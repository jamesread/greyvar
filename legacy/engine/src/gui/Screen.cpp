#include "gui/screens/Screen.hpp"

using std::move;

void Screen::add(GuiComponent* comp, LayoutConstraints* lc) {
	comp->setConstraints(lc);

	this->components.push_back(move(comp));
}

void Screen::onClick(const int x, const int y) {
	GuiComponent* componentUnderClick = this->getComponentAt(x, y);

	if (componentUnderClick != nullptr) {
		componentUnderClick->onClick();
	}
}

void Screen::onMouseMoved(const int x, const int y) {
	if (this->focussedComponent != nullptr) {
		this->focussedComponent->hasFocus = false;
		this->focussedComponent = nullptr;
	}

	GuiComponent* componentUnderCursor = this->getComponentAt(x, y);

	if (componentUnderCursor != nullptr) {
		componentUnderCursor->hasFocus = true;

		this->focussedComponent = componentUnderCursor;
	}
}

GuiComponent* Screen::getComponentAt(const int x, const int y) const {
	for (auto comp : this->components) {
		if (x >= comp->pos.x && x <= comp->pos.xw()) {
			if (y >= comp->pos.y && y < comp->pos.yh()) {
				return comp;
			}
		}
	}

	return nullptr;
}

std::string Screen::getPreviousScreenName() { 
	return "noprev";
}


