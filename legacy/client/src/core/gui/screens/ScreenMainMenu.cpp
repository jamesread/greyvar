#include <iostream>

#include "ScreenMainMenu.hpp"
#include <boleas/gui/components/Label.hpp>
#include <boleas/gui/Gui.hpp>
#include <boleas/boleas.hpp>

using namespace std;

string ScreenMainMenu::getRandomSubtitle() {
	cout << "is motd set? " << cvarIsset("motd") << endl;

	if (cvarIsset("motd")) {
		return cvarGet("motd");
	}

	std::vector<string> subtitles;
	subtitles.push_back(move("Hello there!"));
	subtitles.push_back(move("No Java here, buddy."));
	subtitles.push_back(move("I like to code it code it"));
	subtitles.push_back(move("mooorawr"));

	int c = rand() % subtitles.size();

	return subtitles.at(c);
}

void quitMenuAction() {
	boleasQuitEngine();
}

void aboutMenuAction() {
	Gui::get().setScreen("about");
}

void settingsMenuAction() {
	Gui::get().setScreen("settings");
}

void serversMenuAction() {
	Gui::get().setScreen("servers");
}

void ScreenMainMenu::setupComponents() {
	auto cons = new LayoutConstraints();

	auto lblTitle = new Label("Greyvar", 48);

	cons->colWeight = 0;
	cons->rowWeight = 0;
	this->add(lblTitle, cons);

	auto lblSubtitle = new Label("^3" + this->getRandomSubtitle());

	cons->row++;
	cons->rowWeight = 0;
	this->add(lblSubtitle, cons);

	cons->row++;
	this->add(new Button("Play", &serversMenuAction), cons);
	
	cons->row++;
	this->add(new Button("Settings", &settingsMenuAction), cons);

	cons->row++;
	this->add(new Button("About", &aboutMenuAction), cons);

	cons->row++;
	this->add(new Button("Quit", &quitMenuAction), cons);

}


