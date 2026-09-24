#pragma once

#include <boleas/gui/screens/Screen.hpp>

class ScreenServerBrowser: public Screen {
	public:
		ScreenServerBrowser() {
			this->setupComponents();
		}

		void setupComponents();
		void executeCurrentItem() override;
};
