#include <string>

#include "gui/Gui.hpp"
#include "util.hpp"
#include "cvars.hpp"
#include "Startup.hpp"

using namespace std;

Gui::Gui() {
	this->currentScreen = this->defaultScreen;
	this->layoutManager->onChanged(this->currentScreen);
}

void Gui::toggleConsole() {
	playSound("interface/interface3.wav", UI);

	if (this->scene == CONSOLE) {
		this->scene = PLAY;
	} else {
		this->scene = CONSOLE;
	}
}

void Gui::refreshPlayers() {
	std::cout << "FIXME refreshPlayers" << std::endl;
//	this->screenPlayerSetup->setupComponents();
//	this->layoutManager->onChanged(this->screenPlayerSetup);
//	this->layoutManager->doLayout(this->screenPlayerSetup);
}

void Gui::onMouseMoved(const int x, const int y) const {
	if (this->scene == MENU) {
		this->currentScreen->onMouseMoved(x, y);
	}
}

void Gui::quit() {
	std::cout << "Quitting screens" << std::endl;

	for (auto screen : this->screens) {
		delete(screen.second);
	}
}

