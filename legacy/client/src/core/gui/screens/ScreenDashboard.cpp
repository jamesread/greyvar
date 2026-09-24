#include "ScreenDashboard.hpp"

#include <boleas/gui/components/Label.hpp>
#include <boleas/gui/components/Button.hpp>

void nullAction() {}

ScreenDashboard::ScreenDashboard() {
	auto cons = new LayoutConstraints();
	cons->colWeight = 1;

	this->add(new Label("Greyvar dashboard test", 42), cons);

	cons->row++;
	auto menu = new Menu();
	menu->add("foo");
	menu->add("bar");
	menu->add("one");
	menu->add("two");
	menu->add("three");
	menu->add("four");
	this->add(menu, cons);


	cons->row++;
	cons->rowWeight = 1;
	this->add(new Button("one", {255, 0, 0, 255}, &nullAction), cons);

	cons->row++;
	cons->rowWeight = 0;
	this->add(new Button("two", {0, 255, 0, 255}, &nullAction), cons);

	cons->row++;
	cons->rowWeight = 1;
	this->add(new Button("two", {0, 0, 255, 255}, &nullAction), cons);

	cons->col = 1;
	cons->row = 1;
	this->add(new Label("label 1"), cons);

	cons->row++;
	this->add(new Label("label 2"), cons);

	cons->row++;
	this->add(new Label("label 3"), cons);

}
