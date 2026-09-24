#pragma once

#include "RemotePlayer.hpp"
#include "input/0_hid/HidInputDevice.hpp"

class LocalPlayer {
	public: 
		string username;

		uint32_t playerId;

		RemotePlayer* remote;
		HidInputDevice inputDevice;
};
