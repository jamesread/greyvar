#include <SDL2/SDL_mixer.h>
#include <iostream>
#include <map>

#include "Sound.hpp"
#include "cvars.hpp"

using std::string, std::endl, std::cout;

map<string, Mix_Chunk*> soundCache;

#define SND_FREQUENCY 44100
#define SND_CHANNELS 2
#define SND_SAMPLE_SIZE 2048

Mix_Chunk* loadSound(const string& filename) {
	if (soundCache.count(filename) == 0) {
		Mix_Chunk* snd = Mix_LoadWAV(string("res/snd/").append(filename).c_str());

		if (snd == nullptr) {
			cout << "failed to load sound: " << filename << endl;
		} else {
			cout << "loaded sound: " << filename << endl;
		}

		soundCache[filename] = snd;
	}

	Mix_Volume(SoundChannel::UI, cvarGeti("snd_channel_ui_volume"));

	return soundCache[filename];
}

void initSound() {
	Mix_Init(MIX_INIT_MOD);

	if (Mix_OpenAudio(SND_FREQUENCY, MIX_DEFAULT_FORMAT, SND_CHANNELS, SND_SAMPLE_SIZE) < 0) {
		cout << "Audio setup failed." << endl;
	} else {
		cout << "Audio setup complete." << endl;
	}

	loadSound("interface/interface1.wav");
}

void quitSound() {
	cout << "quit sound" << endl;

	for (auto it : soundCache) {
		Mix_FreeChunk(it.second);
	}

	soundCache.clear();

	Mix_Quit();
}

void playSound(const string& filename, SoundChannel ch) {
	cout << "play sound: " << filename << endl;

	if (Mix_PlayChannel(ch, loadSound(filename), 0) == -1) {
		cout << "failed to play: " << Mix_GetError() << endl;
	}
}
