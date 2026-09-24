#pragma once

#include <boleas/gui/screens/Screen.hpp>

class ScreenAbout : public Screen {
	public:
		ScreenAbout() {
			this->setupComponents();
		}

		void setupComponents();
};
