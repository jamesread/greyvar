#pragma once

#include <iostream>
#include <vector>

#include "../../Sound.hpp"
#include "GuiComponent.hpp"
#include "MenuItem.hpp"

using namespace std;

class Menu: public GuiComponent {
	public:
		Menu() {
			this->rendererFunc = "menu";
		}

		void add(string title, MenuItemAction action) {
			this->items.push_back(new MenuItem(title, action));

			this->minimumHeight = 30 * (this->items.size() + 1);
		}

		void add(string title) {
			this->add(title, nullptr);
		}

		void selectPrevMenuItem() {
			if (this->currentlySelectedMenuItem > 0) {
				this->currentlySelectedMenuItem--;

				playSound("interface/interface1.wav", UI);
			} else {
				playSound("interface/interface2.wav", UI);
			}
		}

		void selectNextMenuItem() {
			if (this->currentlySelectedMenuItem < this->items.size() - 1) {
				this->currentlySelectedMenuItem++;

				playSound("interface/interface1.wav", UI);
			} else {
				playSound("interface/interface2.wav", UI);
			}
		}

		void executeCurrentMenuItem() {
			playSound("interface/interface3.wav", UI);

			MenuItem* item = this->items.at(this->currentlySelectedMenuItem);

			item->action();
		}

		int getItemCount() {
			return this->items.size();
		}

		int getSelectedIndex() {
			return this->currentlySelectedMenuItem;
		}

		string getItemName(uint32_t item) {
			return this->items.at(item)->text;
		}

	private:

		std::vector<MenuItem*> items;
		
		unsigned int currentlySelectedMenuItem = 0;

};
