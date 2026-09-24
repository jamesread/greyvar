#include "GameState.hpp"
#include "Renderer.hpp"
#include "gui/components/Label.hpp"
#include "gui/components/TextureViewer.hpp"
#include "gui/components/Button.hpp"
#include "gui/components/Menu.hpp"
#include "gui/components/ComboSpinner.hpp"
#include "gui/Gui.hpp"
#include "gui/utils/TextAlignment.hpp"
#include "util.hpp"

#include <ctime>

using namespace std;

wstring s2ws(string str) {
	using convert_typeX = std::codecvt_utf8<wchar_t>;
    std::wstring_convert<convert_typeX, wchar_t> converterX;

    return converterX.from_bytes(str);
}

void renderComponent(GuiComponent* component) {
	renderTextShadowWithBackground(L"^1<unsupported component!>", component->pos.x, component->pos.y, 24, LEFT_MIDDLE, colorText, colorTextSubtle, 0);
}

void renderLabel(Label* lbl) {
	if (false) {
		renderTextShadowWithBackground(s2ws(lbl->getText()), lbl->pos.x, lbl->pos.y, lbl->fontSize, LEFT_MIDDLE, colorText, colorTextSubtle, 0);
	} else {
		renderTextShadow(s2ws(lbl->getText()), lbl->pos.x, lbl->pos.y, lbl->fontSize);
	}
}

void renderButton(Button* btn) {
	const short borderWidth = 2;

	renderRect(btn->color, btn->pos.x, btn->pos.y, btn->pos.w, btn->pos.h);

	SDL_Color shadow  = btn->color;
	SDL_Color txtColor;

	if (btn->hasFocus) {
		shadow = colorHighlight;
		txtColor = colorTextHover;
	} else {
		shadow.r = shadow.g = shadow.b -= 100;
		txtColor = colorText;
	}

	renderRect(shadow, btn->pos.x + borderWidth, btn->pos.y + borderWidth, btn->pos.w - (borderWidth * 2), btn->pos.h - (borderWidth * 2));

	renderText(s2ws(btn->getText()), btn->pos, txtColor, true, 24, LEFT_MIDDLE);
}

void renderMenu(Menu* menu) {
	const int itemHeight = 30;
	int yOffset = 0;

	const int selectedIndex = menu->getSelectedIndex();

	for (int i = 0; i < menu->getItemCount(); i++) {
		SDL_Color color{255, 255, 255, 255};

		if (i == selectedIndex) {
			color = {0, 255, 255, 255};
			renderTextShadow(L">", menu->pos.x - 20, menu->pos.y + yOffset + (i * itemHeight), 100, 100, 24, LEFT_MIDDLE, color);
		}

		renderTextShadow(s2ws(menu->getItemName(i)), menu->pos.x, menu->pos.y + yOffset + (i * itemHeight), 24);
	}
}

void renderTexture(TextureViewer* view) {
	SDL_Rect pos{};
	pos.x = view->pos.x;
	pos.y = view->pos.y;
	pos.w = view->size;
	pos.h = view->size;

	auto tex = Renderer::get().resCache->loadTexture(view->textureName);

	SDL_RenderCopy(Renderer::get().sdlRen, tex, nullptr, &pos);
}

void renderComboSpinner(ComboSpinner* cbo) {
	renderTextShadow(s2ws(cbo->getTitle()), cbo->pos.x, cbo->pos.y, 24);

	renderTextShadow(L"<", cbo->pos.x + 280, cbo->pos.y, 24);

	if (cbo->hasFocus) {
		renderTextShadowWithBackground(s2ws(cbo->getValue()), cbo->pos.x + 300, cbo->pos.y, 24, LEFT_MIDDLE, colorInactive, colorHighlight, 0);
	} else {
		renderTextShadowWithBackground(s2ws(cbo->getValue()), cbo->pos.x + 300, cbo->pos.y, 24, LEFT_MIDDLE, colorInactive, colorText, 0);
	}

	renderTextShadow(L">", cbo->pos.x + 600, cbo->pos.y, 24);
}

void renderSplash() {
	using namespace std::chrono;

	static const auto SPLASH_TIME = 3000.0;
	static auto splashStarted = high_resolution_clock::now();
	auto diff = high_resolution_clock::now() - splashStarted;

	const int w = Renderer::get().window_w;
	const int h = Renderer::get().window_h;

	int c = duration_cast<milliseconds>(diff).count();

	SDL_Color textColor{255, 255, 255, 0};

	if (c < SPLASH_TIME) {
		textColor.r = textColor.g = textColor.b = (c / SPLASH_TIME) * 255;
	} else {
		textColor.r = textColor.g = textColor.b = 255 - ((c / SPLASH_TIME) * 255);

		if (textColor.r < 2) {
			Gui::get().setScreen("main");
		}
	}

	renderRect({0, 0, 0, 255}, 0, 0, Renderer::get().window_w, Renderer::get().window_h);

	renderTextShadow(L"Greyvar", w / 2, h * .8, 100, 100, 72, CENTER, textColor);
}

void renderPanel() {
	renderBackgroundSolidColor({130, 130, 130, 255});

	auto currentScreen = Gui::get().currentScreen;

	if (cvarGetb("debug_layout_manager")) {
		for (auto pair : Gui::get().layoutManager->rowProperties) {
			auto row = pair.second;

			renderRect({175, 175, 225, 255}, 0, row.offset + 32, Renderer::get().window_w, row.size);
		}

		for (auto pair : Gui::get().layoutManager->colProperties) {
			auto col = pair.second;

			renderRect({225, 175, 175, 255}, col.offset + 32, 0, col.size, Renderer::get().window_h);

		}
	}

	if (currentScreen != nullptr) {
		for (auto component : currentScreen->components) {
			if (component->rendererFunc == "label") {
				renderLabel((Label*)component);
			} else if (component->rendererFunc == "combospinner") {
				renderComboSpinner((ComboSpinner*)component);
			} else if (component->rendererFunc == "button") {
				renderButton((Button*)component);
			} else if (component->rendererFunc == "menu") {
				renderMenu((Menu*)component);
			} else if (component->rendererFunc == "texture") {
				renderTexture((TextureViewer*)component);
			} else if (component->rendererFunc == "splash") {
				renderSplash();
			} else {
				renderComponent(component);
			}
		}
	}
}

