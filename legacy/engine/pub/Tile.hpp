#pragma once

#include <iostream>
#include <SDL2/SDL.h>

using namespace std;

class Tile {
	public:
		Tile(const string& textureName, bool traversable);

		string textureName;

		bool texHFlip = false;
		bool texVFlip = false;
		int texRot = 0;
	private:
		bool traversable;

};
