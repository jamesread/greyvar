#include "ScreenSplash.hpp"
#include <boleas/gui/components/SplashyThingy.hpp>

void ScreenSplash::setupComponents() {
	auto cons = new LayoutConstraints();

	this->add(new SplashyThingy(), cons);
}

