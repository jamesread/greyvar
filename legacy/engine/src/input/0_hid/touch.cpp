#include "GameState.hpp"
#include "../PlayerInput.hpp"

// Touch input events will always use the first local player.

void toucheytouchey() {
	new PlayerInput(GameState::get().getFirstLocalPlayer(), TOUCH_TAP);
}
