#pragma once

#include <map>
#include <SDL2/SDL.h>

#include <ft2build.h>
#include FT_FREETYPE_H

#include <string>

using std::string;
using std::map;

class ResCache {
	public:
		SDL_Texture* loadTexture(const string& filename);
		SDL_Texture* loadTexture(const string& filename, uint32_t primaryColor);
		SDL_Texture* loadEntity(const string& filename);
		SDL_Texture* loadEntity(const string& filename, uint32_t primaryColor);
		SDL_Texture* loadHud(const string& filename);
		SDL_Texture* loadTile(const string& filename);

		SDL_Surface* loadImageUncached(const string& filename, uint32_t primaryColor);

		FT_Face* loadFont(const string& filename, uint16_t size);

		void loadStartup();

	private:
		map<string, SDL_Texture*> textureCache = {};

		map<string, FT_Face*> fontCache;
};
