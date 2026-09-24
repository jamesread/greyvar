#include "Screen.hpp"

#include "../layout/LayoutConstraints.hpp"
#include "../components/Label.hpp"

class DefaultScreen : public Screen {
	public:
		DefaultScreen() {
			auto cons = new LayoutConstraints();

			this->add(new Label("Boleas Engine, Default Screen"), cons);
		}
};
