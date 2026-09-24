#include "GameState.hpp"
#include "Renderer.hpp"
#include "LocalPlayer.hpp"

#define JOIN_TXT_KEYBOARD 1
#define JOIN_TXT_GAMEPAD_KEYBOARD 2
#define JOIN_TXT_GAMEPAD 3

using std::wstring;

void renderHudPlayer(LocalPlayer* lp, int x, int y, int fontSize, bool firstPlayer) {
	string statusText = "";
	string usernameText = "";

	SDL_Color playerColor{};
	uint8_t backgroundIntensity;

	if (lp->remote == nullptr) {
		playerColor = {255, 255, 255, 200};
		backgroundIntensity = 100;
		usernameText = "nobody";
	} else {
		playerColor = rgbaToSdlColor(lp->remote->ent->primaryColor);
		backgroundIntensity = 200;
		usernameText = "username";// lp->username;
	}
	
	renderRect(playerColor, x - (fontSize * .5), y - (fontSize * 1), fontSize / 2, fontSize + (5 * 2) + (fontSize / 2));
	renderRect({100, 100, 100, backgroundIntensity}, x, y - (fontSize * 1), 172, 48);

	SDL_Rect inputDeviceIconPos{};
	inputDeviceIconPos.x = x + 5;
	inputDeviceIconPos.y = (y - fontSize) + (5 * 1.5);
	inputDeviceIconPos.w = 32;
	inputDeviceIconPos.h = 32;

	SDL_Texture* texInputDeviceIcon;


	int statusTextType = 0;
	
	switch (lp->inputDevice.type) {
		case GAMEPAD:
			if (firstPlayer) {
				texInputDeviceIcon = Renderer::get().resCache->loadHud("gamepadKeyboard.png");

				statusTextType = JOIN_TXT_GAMEPAD_KEYBOARD;
			} else {
				texInputDeviceIcon = Renderer::get().resCache->loadHud("gamepad.png");
				statusTextType = JOIN_TXT_GAMEPAD;
			}

			break;
		case KEYBOARD_AND_POINTER:
			texInputDeviceIcon = Renderer::get().resCache->loadHud("keyboard.png");
			statusTextType = JOIN_TXT_KEYBOARD;
			break;
		default:
			texInputDeviceIcon = Renderer::get().resCache->loadHud("question.png");
			break;
	}
	
	switch (statusTextType) {
		case JOIN_TXT_KEYBOARD:
			statusText = "Press Enter";
			break;
		case JOIN_TXT_GAMEPAD_KEYBOARD:
			statusText = "Press Start/Enter";
			break;
		case JOIN_TXT_GAMEPAD:
			statusText = "Press Start";
			break;
	}
	
	int smallTextOffset = 25;

	renderTextShadow(L"username", x + 40, y - fontSize, fontSize);
	renderTextShadow(L"status", x + 40 , (y - fontSize) + smallTextOffset, fontSize / 2);
	
	SDL_RenderCopy(Renderer::get().sdlRen, texInputDeviceIcon, nullptr, &inputDeviceIconPos);
}

void renderHud() {
	int fontSize = 24;
	int y = Renderer::get().window_h - 60;
	int x = -150;

	for (auto lp : GameState::get().getLocalPlayers()) {
		bool firstPlayer = lp == GameState::get().getFirstLocalPlayer();

		x += 200;

		renderHudPlayer(lp, x, y, fontSize, firstPlayer);
	}

	renderTextShadowWithBackground(L"Players: ^5" + to_wstring(GameState::get().getRemotePlayerCount()), Renderer::get().window_w - 200, y, fontSize, LEFT_MIDDLE, colorTextSubtle, colorInactive, 0);
}


