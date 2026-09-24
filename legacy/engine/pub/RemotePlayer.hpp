#pragma once

#include "Entity.hpp"

#include <string>

using namespace std;

class RemotePlayer {
	public:
		RemotePlayer();

		int id = -9999;
		string username = "unnamed player";

		Entity* ent = new Entity();
};
