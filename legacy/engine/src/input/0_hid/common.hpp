#pragma once

// Thank you https://github.com/alexdantas/sdl2-platformer/blob/master/src/InputDefinitions.hpp !
// (& http://wiki.libsdl.org/SDLScancodeLookup )
#define NUM_KEYBOARD_KEYS 282

void recvKeydownInput(SDL_Event e);
void recvKeyupInput(SDL_Event e);
void recvKeyboardInput();
void recvGamepadInputs();
void recvGamepadButtonDown(SDL_Event e);
void initGamepad(Sint32 which);
void removeGamepad(Sint32 which);
void resetKeyboard();
void recvMouseButtonClicked(SDL_Event e);
