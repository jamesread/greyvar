#pragma once

#include "Scene.hpp"
#include "input/0_hid/HidInputGesture.hpp"
#include "input/2_actions/ActionInput.hpp"
#include "input/PlayerInput.hpp"

#include <queue>

extern queue<PlayerInput*> unboundPlayerInputQueue;

void inputBind(Scene scene, HidInputGesture hidInput, ActionInput ai);
void defaultInputBindings();

void lookupActionBindingForPlayerInput();
