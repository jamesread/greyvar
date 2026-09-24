#pragma once

class LayoutConstraints {
	public:
		int row = 0;
		int col = 0;
		int rowWeight = 0;
		int colWeight = 0;

		enum {
			HORIZONTAL,
			VERTICAL,
			BOTH
		} fill = BOTH;

		enum {
			N, E, S, W, NE, NW, SE, SW
		} anchor = NW; 
		
		enum {
			NORTH,
			EAST,
			SOUTH,
			WEST
		} alignment; 
};
