#include "Renderer.hpp"
#include "GameState.hpp"
#include "gui/Gui.hpp"

GameState::~GameState() {
	this->clear();
}

void GameState::clear() {
	cout << "Clearing the GameState" << endl;

	for (auto rp : this->remotePlayers) {
		delete(rp.second);
	}

	this->remotePlayers.clear();

	for (auto lp : this->localPlayers) {
		delete(lp);
	}

	this->localPlayers.clear();

	this->destroyWorld();
}

void GameState::createWorld() {
	this->world = new World();
}

void GameState::destroyWorld() {
	if (this->world != nullptr) {
		delete(this->world);
		this->world = nullptr;
	}
}

void GameState::onPlayerJoin(RemotePlayer* rp) {
	this->remotePlayers[rp->id] = rp;

	this->world->entityGrid->add(rp->ent);
}

void GameState::addNewLocalPlayer(LocalPlayer* lp) {
	lp->username = "username"; //cvarGetW(L"lp_" + to_wstring(this->localPlayers.size()) + L"_username");

	this->localPlayers.push_back(lp);
}

void GameState::onNewLocalPlayer(LocalPlayer* lp) {
	this->addNewLocalPlayer(lp);

	Gui::get().refreshPlayers();
}

void GameState::onRemoveLocalPlayer(LocalPlayer* lp) {
	this->localPlayers.erase(std::remove(this->localPlayers.begin(), this->localPlayers.end(), lp), this->localPlayers.end());

	Gui::get().refreshPlayers();
}
