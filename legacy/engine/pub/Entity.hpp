#pragma once

#include <SDL2/SDL.h>
#include <iostream>

using namespace std;

class Entity {
	public:
		Entity();

		string textureName;
		uint32_t primaryColor;

		SDL_Rect* pos = new SDL_Rect;
};

