#pragma once

#include <boleas/gui/screens/Screen.hpp>
#include <boleas/LocalPlayer.hpp>

class ScreenPlayerSetup : public Screen {
	public:
		ScreenPlayerSetup() {
			this->setupComponents();
		}

		void setupComponents();
		void setupPlayerComponents(LocalPlayer* lp, LayoutConstraints* cons);
};
