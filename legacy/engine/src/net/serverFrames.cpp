#include "net.hpp"
#include "GameState.hpp"

using std::cout;
using std::endl;

using greyvarproto::ServerFrameResponse; 

void processServerFrame(ServerFrameResponse sframe) {
    cout << "Frame: " << endl;

    if (sframe.has_grid()) {
        processServerFrameGrid(sframe.grid());
    }

	for (auto entspawn : sframe.entityspawns()) {
		processServerFrameEntitySpawns(entspawn);
	}

	for (auto entpos : sframe.entitypositions()) {
		auto ent = GameState::get().world->entityGrid->get(entpos.entityid());
		ent->pos->x = entpos.x();
		ent->pos->y = entpos.y();
	}

	if (sframe.has_playerjoined()) {
		auto rp = new RemotePlayer();
		rp->username = sframe.playerjoined().username();

		GameState::get().onPlayerJoin(rp);
	}
}
