#pragma once

#include <map>
#include "TileGrid.hpp"
#include "EntityGrid.hpp"

using namespace std;

class World {
	public: 
		~World() {
			delete(this->tileGrid);
			delete(this->entityGrid);
		}

		TileGrid *tileGrid = new TileGrid();
		EntityGrid *entityGrid = new EntityGrid();
};
