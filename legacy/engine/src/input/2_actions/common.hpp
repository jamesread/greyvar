#pragma once

#include <queue>

#include "input/PlayerInput.hpp"
#include "LocalPlayer.hpp"
#include "input/PlayerInput.hpp"

extern queue<PlayerInput*> boundPlayerInputQueue;

void executeActionInputs();

