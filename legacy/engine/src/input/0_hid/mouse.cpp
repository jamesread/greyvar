#include "GameState.hpp"
#include "../PlayerInput.hpp"

// Mice input events will always use the first local player.

void recvMouseButtonClicked(SDL_Event e) {
	HidInputGesture gesture = MOUSE_CLICK_LEFT;

	switch (e.button.button) {
		case SDL_BUTTON_LEFT:
			gesture = MOUSE_CLICK_LEFT;
			break;
		case SDL_BUTTON_RIGHT:
			gesture = MOUSE_CLICK_RIGHT;
			break;
	}

	auto pi = new PlayerInput(GameState::get().getFirstLocalPlayer(), gesture, false);
	pi->pointerX = e.button.x;
	pi->pointerY = e.button.y;
	pi->queueForBinding();
}
