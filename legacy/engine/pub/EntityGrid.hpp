#pragma once

#include <algorithm>
#include <vector>

#include "Entity.hpp"

class EntityGrid : public Grid {
	public: 
		EntityGrid() : Grid(GRID_SIZE, GRID_SIZE) {}
		int lastClientSideEntityId = -1;

		int add(int serverId, Entity* ent) {
			this->entities[serverId] = ent;

			return serverId;
		}

		int add(Entity* ent) {
			this->entities[this->lastClientSideEntityId--] = ent;

			return this->lastClientSideEntityId;
		}

		void remove(Entity* ent) {
			auto it = this->entities.begin();

			while (it != this->entities.end()) {
				if (it->second == ent) {
					this->entities.erase(it);
				}

				it++;
			}
		}

		Entity* get(int id) {
			if (this->entities.find(id) == this->entities.end()) {
				cout << "Creating new entity " << endl;
				this->add(id, new Entity());
			}

			return this->entities[id];
		}

		map<int, Entity*> entities;
};
