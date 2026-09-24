#include <iostream>
#include <SDL2/SDL.h>
#include <map>

#include "../PlayerInput.hpp"
#include "LocalPlayer.hpp"
#include "GameState.hpp"

using namespace std;

// Gamepad input events can be assigned to arbitary local players (ie: multiple connected gamepads).

map<SDL_GameController*, LocalPlayer*> gamepadsForLocalPlayers; 

const int DEAD_ZONE = 8000;

void handleGamepadAxis(SDL_GameController* controller, LocalPlayer* lp) {
	auto axisX = SDL_GameControllerGetAxis(controller, SDL_CONTROLLER_AXIS_LEFTX);
	auto axisY = SDL_GameControllerGetAxis(controller, SDL_CONTROLLER_AXIS_LEFTY);

	if (abs(axisX) > DEAD_ZONE) {
		if (axisX > 0) {
			new PlayerInput(lp, GAMEPAD_JOYSTICK_RIGHT);
		} else {
			new PlayerInput(lp, GAMEPAD_JOYSTICK_LEFT);
		}
	} 

	if (abs(axisY) > DEAD_ZONE) {
		if (axisY > 0) {
			new PlayerInput(lp, GAMEPAD_JOYSTICK_DOWN);
		} else {
			new PlayerInput(lp, GAMEPAD_JOYSTICK_UP);
		}
	} 
}

HidInputGesture sdlButtonToHidInputGesture(uint8_t b) {
	switch (b) {
		case SDL_CONTROLLER_BUTTON_A: return GAMEPAD_A;
		case SDL_CONTROLLER_BUTTON_B: return GAMEPAD_B;
		case SDL_CONTROLLER_BUTTON_X: return GAMEPAD_X;
		case SDL_CONTROLLER_BUTTON_Y: return GAMEPAD_Y;
		case SDL_CONTROLLER_BUTTON_START: return GAMEPAD_START;
		case SDL_CONTROLLER_BUTTON_BACK: return GAMEPAD_BACK;
		case SDL_CONTROLLER_BUTTON_LEFTSTICK: return GAMEPAD_LEFTSTICK;
		case SDL_CONTROLLER_BUTTON_RIGHTSTICK: return GAMEPAD_RIGHTSTICK;
		case SDL_CONTROLLER_BUTTON_DPAD_UP: return GAMEPAD_DPAD_UP;
		case SDL_CONTROLLER_BUTTON_DPAD_DOWN: return GAMEPAD_DPAD_DOWN;
		case SDL_CONTROLLER_BUTTON_DPAD_LEFT: return GAMEPAD_DPAD_LEFT;
		case SDL_CONTROLLER_BUTTON_DPAD_RIGHT: return GAMEPAD_DPAD_RIGHT;
		case SDL_CONTROLLER_BUTTON_GUIDE: return GAMEPAD_GUIDE;
		default: return GAMEPAD_INVALID;
	}
}

void recvGamepadButtonDown(SDL_Event e) {
	LocalPlayer* lp = gamepadsForLocalPlayers[SDL_GameControllerFromInstanceID(e.cbutton.which)];

	new PlayerInput(lp, sdlButtonToHidInputGesture(e.cbutton.button));
}

void removeGamepad(Sint32 i) {
	SDL_GameController* controller = SDL_GameControllerFromInstanceID(i);

	LocalPlayer* lp = gamepadsForLocalPlayers[controller];

	cout << "Removing local player: " << lp << endl;

	GameState::get().onRemoveLocalPlayer(lp);
}

void initGamepad(Sint32 i) {
	//for (int i = 0; i < SDL_NumJoysticks(); i++) {
	
	if (SDL_IsGameController(i)) {
		SDL_GameController* controller = SDL_GameControllerOpen(i);

		if (!controller) {
			cout << "Could not open gamepad: " << i << endl;
		} else {
			cout << "Found gamepad " << i << ": " << SDL_GameControllerName(controller) << endl;

			auto lp = new LocalPlayer();
			lp->inputDevice.type = GAMEPAD;
			lp->inputDevice.device.gamepad = controller;

			gamepadsForLocalPlayers[controller] = lp;

			GameState::get().onNewLocalPlayer(lp);
		}
	}
}

void recvGamepadInputs() {
	for (auto pair : gamepadsForLocalPlayers) {
		handleGamepadAxis(pair.first, pair.second);
	}
}


