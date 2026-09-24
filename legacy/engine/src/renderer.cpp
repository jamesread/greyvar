#include <chrono>

#include <SDL2/SDL_image.h>

#include <ft2build.h>
#include FT_FREETYPE_H

#include "GameState.hpp"
#include "Renderer.hpp"
#include "gui/Gui.hpp"
#include "cvars.hpp"

SDL_Color colorHighlight {100, 255, 255, 255};
SDL_Color colorInactive {100, 100, 100, 255};
SDL_Color colorText {255, 255, 255, 255};
SDL_Color colorTextHover {0, 0, 0, 255};
SDL_Color colorTextSubtle {100, 100, 100, 255};

void renderGridTiles(World* world) {
	int windowWidth = 2 * floor(Renderer::get().window_w / 2);
	int windowHeight = 2 * floor(Renderer::get().window_h / 2);

	int tileRenderLength;

	if (cvarGetb("r_scale_grid")) {
		tileRenderLength = std::min(windowWidth, windowHeight) / 16;
	} else {
		 tileRenderLength = TILE_SIZE;
	}

	int gridTilesWidth = world->tileGrid->w * tileRenderLength;
	int gridTilesHeight = world->tileGrid->h * tileRenderLength;

	int padx = (windowWidth - gridTilesWidth) / 2;
	int pady = (windowHeight - gridTilesHeight) / 2;

	for (int row = 0; row < world->tileGrid->h; row++) {
		for (int col = 0; col < world->tileGrid->w; col++) {
			Tile* tile = world->tileGrid->get(row, col);

			SDL_Rect pos;


			pos.x = padx + (col * tileRenderLength);
			pos.y = pady + (row * tileRenderLength);
			pos.w = tileRenderLength;
			pos.h = tileRenderLength;

			int rot = tile->texRot;

			int flip = SDL_FLIP_NONE;
			if (tile->texVFlip && tile->texHFlip) {
				flip = SDL_FLIP_HORIZONTAL | SDL_FLIP_VERTICAL;
			} else if (tile->texVFlip) {
				flip = SDL_FLIP_VERTICAL;
			} else if (tile->texHFlip) {
				flip = SDL_FLIP_HORIZONTAL;
			}

			SDL_RenderCopyEx(Renderer::get().sdlRen, Renderer::get().resCache->loadTile(tile->textureName), nullptr, &pos, rot, nullptr, (SDL_RendererFlip)flip);
		}
	}
}

void renderGridEntities(World* world) {
	int windowWidth = 2 * floor(Renderer::get().window_w / 2);
	int windowHeight = 2 * floor(Renderer::get().window_h / 2);



	int tileRenderLength = TILE_SIZE;

	int gridTilesWidth = world->tileGrid->w * tileRenderLength;
	int gridTilesHeight = world->tileGrid->h * tileRenderLength;

	int padx = (windowWidth - gridTilesWidth) / 2;
	int pady = (windowHeight - gridTilesHeight) / 2;

	for (auto p : world->entityGrid->entities) {
		auto e = p.second;
		SDL_Rect r; 
		r.x = padx + (e->pos->x * 4);
		r.y = pady + (e->pos->y * 4);
		r.w = e->pos->w;
		r.h = e->pos->h;

		SDL_Texture* tex = Renderer::get().resCache->loadEntity(e->textureName, e->primaryColor);
		
		SDL_RenderCopy(Renderer::get().sdlRen, tex, nullptr, &r);

	}
}


void renderConsole() {
	renderBackgroundSolidColor({66, 66, 66, 255});

/**
	for (unsigned int i = 0; i < GameState::get().gui->consoleHistory.size(); i++) {
		renderTextShadow(GameState::get().gui->consoleHistory.at(i), 50, 50 + (i * 20), 24);
	}
*/
}

void renderUiMessages() {
	for (auto it : Gui::get().messages) {
//		renderTextShadow(L"FIXME it.second", 50, (Renderer::get().window_h - 100) - (1 * 50), 24);
	}
}

void renderFps() {
	if (cvarGetb("r_fps")) {
		renderText(L"FPS: " + to_wstring(Renderer::get().averageFps), Renderer::get().window_w - 0, 20, {230, 230, 230, 255}, false, 16, CENTER);
	}
}

void Renderer::renderFrame() {
	//std::cout << "render frame ----------------------" << std::endl;

	SDL_Renderer* ren = Renderer::get().sdlRen;
	SDL_RenderClear(ren);

	World* world = GameState::get().world;

	if (Renderer::get().window_h < 480 || Renderer::get().window_w < 640) {
		renderTextShadow(L"RESOLUTION TOO LOW!", 20, 20, 16);
		renderTextShadow(L"Needs to be 640x480 or bigger", 20, 40, 16);
		renderTextShadow(to_wstring(Renderer::get().window_w) + L"x" + to_wstring(Renderer::get().window_h), 20, 60, 16);
		SDL_RenderPresent(ren);
		return;
	}

	switch (Gui::get().scene) {
		case MENU:
			renderPanel();
			break;
		case CONSOLE:
			renderConsole();
			break;
		case PLAY:
			renderBackgroundSolidColor({0, 0, 0, 255});
			renderGridTiles(world);
			renderGridEntities(world);
			renderHud();
			break;
		case NONE:
			renderBackgroundSolidColor({30, 30, 30, 255});
	}

	renderFps();	
	renderUiMessages();

	SDL_RenderPresent(ren);
}


