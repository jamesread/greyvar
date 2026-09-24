#include "GameState.hpp" 
#include "ActionInput.hpp" 

#include <chrono>

#include <iostream> 

using std::cout; 
using namespace std;
using namespace std::chrono;

/**
This namespace implements;

- input rate limiting
- input semaphores

*/
namespace ActionInputFirewall {

ActionInputType getActionInputType(ActionInput ai) {
	if (ai > _WALK_START && ai < _WALK_END) {
		return WALK;
	} else if (ai > _MENU_START && ai < _MENU_END) {
		return USE_MENU;
	} else if (ai == POINT) {
		return OTHER;
	}

	cout << "AI Firewall, unclassified type for input action: " << ai << endl;

	return OTHER;
}

bool canWalk() {
	if (!GameState::get().isFirstLocalPlayerConnected()) {
		cout << "no local player" << endl;
		return false;
	}

//	if (NetworkManager::get().waitingForMove) {
//		return false;
//	} else {
		return true;
//	}
}

system_clock::time_point nextMenuAction = system_clock::now();

bool canUseMenu() {
	if (system_clock::now() > nextMenuAction) {
		nextMenuAction = system_clock::now() + milliseconds(350);
		return true;
	} else {
		return false;
	}
}

bool canDo(ActionInput ai) {
	ActionInputType ait = getActionInputType(ai);

	switch (ait) {
		case USE_MENU: return canUseMenu();
		case WALK: return canWalk();
		case OTHER: return true;
	}

	cout << "canDo() - ait not checked: " << ait << endl;

	return false;
}



}
