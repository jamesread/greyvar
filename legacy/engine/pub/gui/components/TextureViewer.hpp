#include "GuiComponent.hpp"

class TextureViewer : public GuiComponent {
	public:
		TextureViewer(const string& texture, const uint32_t size) {
			this->textureName = texture;
			this->rendererFunc = "texture";

			this->size = size;

			this->minimumHeight = this->size + 32;
			this->minimumWidth = this->size + 32;
		}

		TextureViewer(string texture) : TextureViewer(texture, 64) {
		}

		string textureName{};

		string toString() const override {
			return "TextureViewer {" + this->textureName + "}";
		}

		uint32_t size;
	private:
};
