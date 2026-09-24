#include "ScreenPlayerSetup.hpp"

#include <boleas/GameState.hpp>

#include <boleas/gui/components/Label.hpp>
#include <boleas/gui/components/Button.hpp>
#include <boleas/gui/components/TextureViewer.hpp>
#include <boleas/gui/Gui.hpp>
#include <boleas/net/NetClient.hpp>

void actionPlay() {
//	NetClient::get().playerSetup();
	Gui::get().scene = PLAY;
}

void actionPlayerSetupBack() {
	Gui::get().setScreen("servers");
}

const string getInputDeviceIcon(HidInputDeviceType type) {
	string inputDeviceIcon = "hud/question.png";

	switch (type) {
		case KEYBOARD_AND_POINTER:
			inputDeviceIcon = "hud/keyboard.png";
			break;

		case GAMEPAD:
			inputDeviceIcon = "hud/gamepad.png";
			break;
	}

	return inputDeviceIcon;
}

void ScreenPlayerSetup::setupPlayerComponents(LocalPlayer* localPlayer, LayoutConstraints* cons) {
	cons->row = 1;
	cons->rowWeight = 0;
	this->add(new Label("Username: " + localPlayer->username), cons);

	cons->row++;
	this->add(new TextureViewer("entities/playerBob.png", 128), cons);

	cons->row++;
	cons->rowWeight = 1;
	this->add(new TextureViewer(getInputDeviceIcon(localPlayer->inputDevice.type)), cons);

	cons->row++;
	cons->rowWeight = 0;
	this->add(new Button("Play", &actionPlay), cons);

	cons->col++;

}

void ScreenPlayerSetup::setupComponents() {
	auto cons = new LayoutConstraints();

	this->components.clear();

	this->add(new Label("Greyvar \u00BB Servers \u00BB Player Setup", 48), cons);

	for (auto localPlayer : GameState::get().getLocalPlayers()) {
		this->setupPlayerComponents(localPlayer, cons);
	}

	cons->row++;
	cons->col = 0;
	this->add(new Button("Back", &actionPlayerSetupBack), cons);
}
