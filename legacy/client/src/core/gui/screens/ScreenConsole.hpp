#pragma once 
 
#include <vector>

#include <boleas/gui/screens/Screen.hpp>

class ScreenConsole: public Screen {
	private:
		std::vector<string> consoleHistory; 
};
