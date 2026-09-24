#include "Renderer.hpp"

#include <ft2build.h>
#include FT_FREETYPE_H

#include <SDL2/SDL.h>

void initFreetype() {
	auto error = FT_Init_FreeType(Renderer::get().freetypeLib);

	if (error) {
		cout << "Freetype init failed: " << error << endl;
	} else {
		cout << "Freetype init complete: " << error << endl;
	}
}

// https://github.com/wutipong/drawtext-sdl2-freetype2-harfbuzz/blob/master/sdl-ft-1/main.cpp
// No license given. Need to check.
SDL_Texture* CreateTextureFromFT_Bitmap(SDL_Renderer* renderer, const FT_Bitmap& bitmap, const SDL_Color& color) {
	SDL_Texture* output = SDL_CreateTexture(renderer, SDL_PIXELFORMAT_RGBA8888, SDL_TEXTUREACCESS_STREAMING, bitmap.width, bitmap.rows);

	void *buffer;
	int pitch;
	
	if (SDL_LockTexture(output, NULL, &buffer, &pitch) != 0) {
		// cout << "Texture lock failed creating texture from FT bitmap" << endl; FIXME
	}

	unsigned char *src_pixels = bitmap.buffer;
	unsigned int *target_pixels = reinterpret_cast<unsigned int*>(buffer);

	SDL_PixelFormat* pixel_format = SDL_AllocFormat(SDL_PIXELFORMAT_RGBA8888);

	for (unsigned int y = 0; y < bitmap.rows; y++)
	{
		for (unsigned int x = 0; x < bitmap.width; x++)
		{
			int index = (y * bitmap.width) + x;

			unsigned int alpha = src_pixels[index];
			unsigned int pixel_value =
				SDL_MapRGBA(pixel_format, color.r, color.g, color.b, alpha);

			target_pixels[index] = pixel_value;
		}
	}

	SDL_FreeFormat(pixel_format);
	SDL_UnlockTexture(output);

	return output;
}

