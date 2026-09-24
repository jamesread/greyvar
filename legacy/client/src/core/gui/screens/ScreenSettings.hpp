#pragma once

#include <boleas/gui/screens/Screen.hpp>
#include <boleas/gui/components/Menu.hpp>

class ScreenSettings : public Screen {
	public:
		ScreenSettings() {
			this->setupComponents();
		}

		void setupComponents(); 

		void selectNextItem() override;
		void selectPrevItem() override;
		void executeCurrentItem() override;

	private:
		Menu* menu;
};
