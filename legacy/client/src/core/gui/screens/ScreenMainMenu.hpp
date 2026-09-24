#pragma once

#include <boleas/gui/components/Menu.hpp>
#include <boleas/cvars.hpp>
#include <boleas/gui/screens/Screen.hpp>

#include <iostream>

using namespace std;

class ScreenMainMenu: public Screen {
	public: 
		ScreenMainMenu() {
			setupComponents();
		}

		void setupComponents();

		string getRandomSubtitle();
		
	private:
		string subtitle; 
};
