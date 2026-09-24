#pragma once

using std::string;

enum SoundChannel {
	UI = 1,
	PLAYER = 2,
};

void playSound(const string& filename, SoundChannel ch);
