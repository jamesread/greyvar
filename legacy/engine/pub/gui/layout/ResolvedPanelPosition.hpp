#pragma once

class ResolvedPanelPosition {
	public: 
		int x = 0;
		int y = 0;
		int w = 0;
		int h = 0;

		int xw() const {
			return x + w;
		}

		int yh() const {
			return y + h;
		}
};
