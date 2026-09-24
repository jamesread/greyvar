#include "gui/components/ComboSpinner.hpp"

void ComboSpinner::add(const string& item) {
	this->values.push_back(item);

	if (this->values.size() == 1) {
		this->current = item;
	}
}
