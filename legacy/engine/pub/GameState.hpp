#pragma once

#include <vector>

#include "cvars.hpp"
#include "World.hpp"
#include "LocalPlayer.hpp"
#include "RemotePlayer.hpp"

class GameState {
	public:
		World* world = nullptr;

		std::string serverName = "the construct";

		~GameState();
		GameState(GameState const&) = delete;
		void operator=(GameState const&) = delete;

		static GameState& get() {
			static GameState instance;

			return instance;
		}

		void clear();

		void onPlayerJoin(RemotePlayer* rp);

		void removePlayerById(int id) {
			RemotePlayer* rp = this->getRemotePlayerById(id);

			this->world->entityGrid->remove(rp->ent);

			this->remotePlayers.erase(id);

			delete rp;
		}

		RemotePlayer* getRemotePlayerById(int p) {
			return GameState::get().remotePlayers[p];
		}

		LocalPlayer* getFirstLocalPlayer() {
			return this->localPlayers.at(0);
		}

		std::vector<LocalPlayer*> getLocalPlayers() {
			return this->localPlayers;
		}

		int getRemotePlayerCount() {
			return this->remotePlayers.size();
		}

		bool isFirstLocalPlayerConnected() {
			return this->getFirstLocalPlayer()->remote == nullptr;
		}

		void onNewLocalPlayer(LocalPlayer* lp);
		void onRemoveLocalPlayer(LocalPlayer* lp);

		void createWorld();
		void destroyWorld();

	private:
		void addNewLocalPlayer(LocalPlayer* lp);

		std::vector<LocalPlayer*> localPlayers = {};
		std::map<int, RemotePlayer*> remotePlayers = {};

		GameState() {
			// Initializing the UI here means it is delayed until the first 
			// get() istead of at startup. This allows us more flexibility to 
			// set cvars before startup.
			
			auto keyboardPlayer = new LocalPlayer();
			keyboardPlayer->inputDevice.type = KEYBOARD_AND_POINTER;

			this->addNewLocalPlayer(keyboardPlayer);
		}
};


