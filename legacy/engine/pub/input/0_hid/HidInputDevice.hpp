#pragma once

enum HidInputDeviceType {
	KEYBOARD_AND_POINTER,
	GAMEPAD
};

union HidInputDevicePointer {
	SDL_GameController* gamepad;
	size_t keyboard;
};

class HidInputDevice {
	public:
	HidInputDeviceType type;
	HidInputDevicePointer device;
};
