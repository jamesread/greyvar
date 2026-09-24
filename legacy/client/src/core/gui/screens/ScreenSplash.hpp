#pragma once

#include <boleas/gui/screens/Screen.hpp>

class ScreenSplash : public Screen {
	public: 
		ScreenSplash() {
			this->setupComponents();
		}

		void setupComponents();
};
