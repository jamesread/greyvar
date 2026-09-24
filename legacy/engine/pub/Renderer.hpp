#pragma once

#include <SDL2/SDL.h>
#include "ResCache.hpp"
#include "World.hpp"
#include "gui/utils/TextAlignment.hpp"
#include "gui/layout/ResolvedPanelPosition.hpp"
#include "cvars.hpp"

#define TILE_SIZE 64

class Renderer {	
	public: 
		static Renderer& get() {
			static Renderer instance;

			return instance;
		}

		static void set(SDL_Window* win, SDL_Renderer* renderer) {
			get().sdlRen = renderer;
			get().win = win;
		}

		void createWindow() {
			SDL_Window *win = SDL_CreateWindow("Greyvar", cvarGeti("win_x", SDL_WINDOWPOS_CENTERED), cvarGeti("win_y", SDL_WINDOWPOS_CENTERED), TILE_SIZE * GRID_SIZE, 780, SDL_WINDOW_SHOWN);
			SDL_SetWindowResizable(win, SDL_TRUE);
			SDL_SetWindowMinimumSize(win, 640, 480);

			SDL_SetWindowIcon(win, this->resCache->loadImageUncached("logo.png", 0));

			Renderer::set(win, SDL_CreateRenderer(win, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC));
		}

		void destroyWindow() {
			SDL_DestroyWindow(this->win);
		}

		SDL_Window* getWindow() {
			return this->win;
		}

		ResCache* resCache = new ResCache();
		SDL_Renderer* sdlRen;

		Renderer(Renderer const&) = delete;
		void operator=(Renderer const&) = delete;

		void renderFrame();

		int window_w = 0;
		int window_h = 0;

		FT_Library* freetypeLib = new FT_Library;

		int averageFps = 0;
	private:
		Renderer() {}

		SDL_Window* win;

};

using std::wstring;

void renderRect(SDL_Color color, int x, int y, int w, int h);
void renderText(const wstring& text, int x, int y, SDL_Color color, bool canChangeColor, int size, TextAlignment align);
void renderText(const wstring& text, ResolvedPanelPosition& pos, SDL_Color color, bool canChangeColor, int size, TextAlignment alignment);
void renderTextShadow(const wstring& text, int x, int y, int size);
void renderTextShadow(const wstring& text, int x, int y, int w, int h, int size, TextAlignment alignment, SDL_Color color);
void renderTextShadowWithBackground(const wstring& text, int x, int y, int size, TextAlignment alignment, SDL_Color bgColor, SDL_Color textColor, int offsetX);
void renderBackgroundSolidColor(SDL_Color color);
SDL_Color rgbaToSdlColor(int rgba);

void renderHud();
void renderPanel();

extern SDL_Color colorHighlight;
extern SDL_Color colorInactive;
extern SDL_Color colorText;
extern SDL_Color colorTextHover;
extern SDL_Color colorTextSubtle;
