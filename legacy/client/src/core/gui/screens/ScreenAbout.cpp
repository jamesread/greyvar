#include "ScreenAbout.hpp"

#include <boleas/gui/components/Label.hpp>
#include <boleas/gui/components/Button.hpp>
#include <boleas/gui/Gui.hpp>

void ScreenAbout::setupComponents() {
	auto cons = new LayoutConstraints();
	cons->colWeight = 1;
	cons->rowWeight = 0;

	auto lbl = new Label("Greyvar \u00BB About ");
	lbl->fontSize = 48;
	this->add(lbl, cons);

	auto lblDev = new Label("Greyvar is developed entirely by James Read. Yay.");
	cons->row++;
	cons->rowWeight = 1;
	cons->colWeight = 1;
	this->add(lblDev, cons);

	cons->row++;
	cons->rowWeight = 0;
	cons->colWeight = 0;
	this->add(new Button("Back", []() {
		Gui::get().setScreen("main");
	}), cons);
}

