#include "ScreenSettings.hpp"
#include <boleas/gui/components/Button.hpp>
#include <boleas/gui/components/Label.hpp>
#include <boleas/gui/components/Menu.hpp>
#include <boleas/gui/components/ComboSpinner.hpp>
#include <boleas/gui/Gui.hpp>

void ScreenSettings::setupComponents() {
	auto cons = new LayoutConstraints();
	cons->rowWeight = 0;

	auto lbl = new Label("Greyvar \u00BB Settings", 48);
	this->add(lbl, cons);
	
	auto cboFullscreen = new ComboSpinner("Fullscreen");
	cons->row++;
	this->add(cboFullscreen, cons);

	auto cboSound = new ComboSpinner("Sound");
	cboSound->add("1");
	cboSound->add("2");
	cboSound->add("3");
	cboSound->add("4");
	cboSound->add("5");
	cboSound->add("6");
	cboSound->add("7");
	cboSound->add("8");
	cboSound->add("9");
	cboSound->add("10");
	cons->row++;
	cons->rowWeight = 1;
	this->add(cboSound, cons);

	cons->row++; this->add(new Label("foo"), cons);
	cons->row++; this->add(new Label("foo"), cons);

	auto btnBack = new Button("Back", []() {
		Gui::get().setScreen("main");	
	});

	cons->row++;
	cons->rowWeight = 0;
	this->add(btnBack, cons);
}

void ScreenSettings::selectNextItem() {
	this->menu->selectNextMenuItem();
}

void ScreenSettings::selectPrevItem() {
	this->menu->selectPrevMenuItem();
}

void ScreenSettings::executeCurrentItem() {
	this->menu->executeCurrentMenuItem();
}

