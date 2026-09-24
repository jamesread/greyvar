#pragma once

#include <iostream>
#include <vector>
#include <chrono>
#include <map>

#include "components/MenuItem.hpp"

#include "layout/LayoutManager.hpp"

#include "../Scene.hpp"
#include "../util.hpp"

#include "../Sound.hpp"

#include "screens/DefaultScreen.hpp"

using std::string;

void startGame();

class Gui {
	public:
		Gui(Gui const&) = delete;
		void operator=(Gui const&) = delete;

		static Gui& get() {
			static Gui instance;

			return instance;
		}

		Screen* defaultScreen = new DefaultScreen(); // FIXME const
		Screen* currentScreen;

		map<string, Screen*> screens;

		Scene scene = MENU;

		void quit();

		void toggleConsole();

		void refreshPlayers();

		void toggleMenu() {
			playSound("interface/interface3.wav", UI);

			if (this->scene == MENU) {
				this->scene = PLAY;
			} else {
				this->scene = MENU;
			}
		}

		void addMessage(string message) {
			this->messages[std::chrono::system_clock::now()] = message;
		}

		void onWindowResized() {
			if (this->scene == MENU) {
				this->layoutManager->doLayout(this->currentScreen);
			}
		}

		void onMouseMoved(const int x, const int y) const;

		void goBack() {
			this->setScreen(this->currentScreen->getPreviousScreenName());
		}

		void addScreen(const string& name, Screen* s) {
			this->screens[name] = s;
		}

		void setScreen(const string& name) {
			cout << "Changing screen to " << name << endl;
			Screen* changedScreen = nullptr;

			auto it = this->screens.find(name);

			if (it == this->screens.end()) {
				changedScreen = this->defaultScreen;
			} else {
				changedScreen = it->second;
			}

			this->layoutManager->onChanged(changedScreen);
			this->layoutManager->doLayout(changedScreen);
			this->currentScreen = changedScreen;
		}

		std::map<std::chrono::_V2::system_clock::time_point, string> messages;
		
		LayoutManager* layoutManager = new LayoutManager();
	private:
		Gui();

};

