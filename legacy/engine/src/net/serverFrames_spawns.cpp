#include "GameState.hpp"

#include "net.hpp"

using std::cout;
using std::endl; 

#define TILE_SIZE 64

void processServerFrameEntitySpawns(greyvarproto::EntitySpawn spawn) {
	auto ent = new Entity();
	ent->textureName = spawn.texture();
	ent->pos = new SDL_Rect();
	ent->pos->x = (int) spawn.x();
	ent->pos->y = (int) spawn.y();
	ent->pos->w = TILE_SIZE;
	ent->pos->h = TILE_SIZE;
	ent->primaryColor = spawn.primarycolor();

	GameState::get().world->entityGrid->add(spawn.entityid(), ent);

	cout << "ent spawn " << spawn.texture() << ", id: " << spawn.entityid() << endl;
}
