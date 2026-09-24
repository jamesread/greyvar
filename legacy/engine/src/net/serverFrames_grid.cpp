#include "GameState.hpp"

#include "net.hpp"

using std::cout;
using std::endl; 

void processServerFrameGrid(greyvarproto::Grid grid) {
	TileGrid* tileGrid = GameState::get().world->tileGrid; 

    for (int tileIndex = 0; tileIndex < grid.tiles_size(); tileIndex++) {
        auto netTile = grid.tiles(tileIndex); 
	   	auto memTile = tileGrid->get(netTile.col(), netTile.row()); 

		memTile->textureName = netTile.tex();	
		memTile->texRot = netTile.rot();
		memTile->texHFlip = netTile.fliph();
		memTile->texVFlip = netTile.flipv();
        
        cout << "Update grid tile: " << netTile.DebugString() << endl;               
    }
} 
