#pragma once

#include <map>

#include "Grid.hpp"
#include "Tile.hpp"

class TileGrid : public Grid {
	public:
		~TileGrid() {
			for (int col = 0; col < this->w; col++) {
				for (int row = 0; row < this->h; row++) {
					delete(this->tiles[row][col]);
				}
			}
		}

		TileGrid() : Grid(GRID_SIZE, GRID_SIZE) {
			for (int col = 0; col < this->w; col++) {
				for (int row = 0; row < this->h; row++) {
					this->tiles[row][col] = new Tile(string("grass.png"), false);
				}
			}
		}

		Tile* get(int row, int col) {
			return this->tiles[row][col];
		}

	private:
		map<int, map<int, Tile*>> tiles;
};

