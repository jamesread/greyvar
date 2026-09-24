#pragma once

#include <iostream>

#include "0_hid/HidInputGesture.hpp"
#include "2_actions/ActionInput.hpp"
#include "LocalPlayer.hpp"

/**
The whole purpose of the input system is to build this object;

- Layer 0 (hid) creates it
- Layer 1 (bindings) finds actions from above layer
- Layer 2 (actions) execute actions, passing them through the ActionFirewall

sticks together from layer 0 and layer 2
*/
class PlayerInput {
	public:
		LocalPlayer* localPlayer;

		// Layer 0 - Hid
		HidInputGesture hidInputGesture;
		HidInputDevice hidDevice;

		// Layer 2 - Action
		ActionInput actionInput;

		PlayerInput(LocalPlayer* localPlayer, HidInputGesture hidInputGesture, bool queue);
		PlayerInput(LocalPlayer* localPlayer, HidInputGesture hidInputGesture) : PlayerInput(localPlayer, hidInputGesture, true) {}

		void queueForBinding();

		friend ostream& operator<<(ostream& out, const PlayerInput& pi) {
			return out << "PI {hid: " << pi.hidInputGesture << ", action: " << pi.actionInput << " }" << endl;
		}

		int pointerX = 0;
		int pointerY = 0;
	private:
		bool queuedForBinding = false;
};

