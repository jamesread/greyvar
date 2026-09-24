#include <queue>
#include <iostream>

#include "PlayerInput.hpp"
#include "input/1_bindings/common.hpp"

using namespace std;

PlayerInput::PlayerInput(LocalPlayer* localPlayer, HidInputGesture hidInputGesture, bool queue) {
	this->localPlayer = localPlayer;
	this->hidInputGesture = hidInputGesture;

	if (queue) {
		this->queueForBinding();
	}
}

void PlayerInput::queueForBinding() {
	if (this->queuedForBinding) {
		throw "Already queued.";
	}

	this->queuedForBinding = true;

	unboundPlayerInputQueue.push(this);
}

