#include "cvars.hpp"
#include "Renderer.hpp"
#include "ResCache.hpp"

#include <SDL2/SDL_image.h>
#include <iostream>

using namespace std;

void putpixel(SDL_Surface *surface, int x, int y, Uint32 pixel);
Uint32 getpixel(SDL_Surface *surface, int x, int y);

void replaceColors(SDL_Surface* img, uint32_t primaryColor) {
	uint32_t replaceColor = SDL_MapRGB(img->format, 230, 23, 230);

	SDL_LockSurface(img);

	for (int x = 0; x < img->w; x++) {
		for (int y = 0; y < img->h; y++) {
			if (getpixel(img, x, y) == replaceColor) {
				putpixel(img, x, y, primaryColor);
			}
		}
	}

	SDL_UnlockSurface(img);
} 

SDL_Texture* ResCache::loadTexture(const string& filename) {
	return this->loadTexture(filename, 0);
}

SDL_Texture* ResCache::loadTexture(const string& filename, uint32_t primaryColor) {
	string textureNameKey = filename + ":" + to_string(primaryColor);

	if (this->textureCache.count(textureNameKey) == 0) {
		auto surf = this->loadImageUncached("textures/" + filename, primaryColor);

		SDL_Texture *tex = SDL_CreateTextureFromSurface(Renderer::get().sdlRen, surf);

		this->textureCache[textureNameKey] = tex;

		SDL_FreeSurface(surf);
	}

	return this->textureCache[textureNameKey];
}

SDL_Surface* ResCache::loadImageUncached(const string& filename, uint32_t primaryColor) {
	SDL_Surface *img = IMG_Load(string(string("res/img/") + filename).c_str());

	if (img == nullptr) {
		std::cout << "loadImageUncached error: " << SDL_GetError() << ". filename: " << filename << std::endl;
	} else {
		if (cvarGetb("debug_textures")) {
			std::cout << "texture loaded: " << filename << std::endl;
		}

		if (primaryColor != 0) {
			cout << "Generating replacement color for tex: " << filename << endl;
			replaceColors(img, primaryColor);
		}
	}

	return img;
}

SDL_Texture* ResCache::loadEntity(const string& filename) {
	return this->loadEntity(filename, 0);
}

SDL_Texture* ResCache::loadEntity(const string& filename, uint32_t primaryColor) {
	return this->loadTexture(string("entities/" + filename), primaryColor);
}

SDL_Texture* ResCache::loadHud(const string& filename) {
	return this->loadTexture(string("hud/" + filename), 0);
}

SDL_Texture* ResCache::loadTile(const string& filename) {
	return this->loadTexture(string("tiles/" + filename));
}

FT_Face* ResCache::loadFont(const string& filename, uint16_t size) {
	const string tag = filename + ":" + std::to_string(size);

	if (this->fontCache.count(tag) == 0) {
		this->fontCache[tag] = new FT_Face;

		cout << "Loading font with freetype " << tag << endl;

		auto loadFontResult = FT_New_Face(*Renderer::get().freetypeLib, string("res/ttf/").append(filename).c_str(), 0, this->fontCache[tag]);

		cout << "Load font freetype result: " << std::hex << loadFontResult << endl;

		FT_Set_Pixel_Sizes(*this->fontCache[tag], size, size);
		FT_Select_Charmap(*this->fontCache[tag], FT_ENCODING_UNICODE);

		if (this->fontCache[tag] == nullptr) {
			cout << "Failed to load font: " << tag << endl;
		} else {
			cout << "Cached font: " << tag << endl;
		}
	}

	return this->fontCache[tag];
}

void ResCache::loadStartup() {
//	this->loadFont("DejaVuSans.ttf");
}

