#pragma once

#include <vector>

#include "../components/GuiComponent.hpp"
#include "../layout/LayoutConstraints.hpp"

using std::vector;
using std::string;

class Screen {
	protected:
		void add(GuiComponent* comp, LayoutConstraints* lc);

	public:
		~Screen() {
			std::cout << "Screen remove" << std::endl;
		}

		vector<GuiComponent*> components{};
		string getPreviousScreenName();
		void onMouseMoved(int x, int y);
		virtual void selectNextItem() {};
		virtual void selectPrevItem() {};
		virtual void executeCurrentItem() {};
		void onClick(int x, int y);

		GuiComponent* getComponentAt(const int x, const int y) const;

		GuiComponent* focussedComponent = nullptr;

};
