#include "cvars.hpp"
#include "Renderer.hpp"
#include "gui/layout/LayoutManager.hpp"

#include <map>

void LayoutManager::onChanged(Screen* screen) {
	this->lastScreen = screen;
	this->rowProperties.clear();
	this->colProperties.clear();

	int numCols = 1;
	int numRows = 1;

	for (const auto& comp : screen->components) {
		if (comp->layoutConstraints.col > numCols) {
			numCols = comp->layoutConstraints.col; 
		}

		if (comp->layoutConstraints.row > numRows) {
			numRows = comp->layoutConstraints.row;
		}
	}

	for (int row = 0; row <= numRows; row++) {
		rowProperties[row].index = row;
	}

	for (int col = 0; col <= numCols; col++) {
		colProperties[col].index = col;
	}

	if (cvarGetb("debug_layout_manager")) {
		cout << "LayoutManager::onChanged, created layout with " << this->colProperties.size() << " cols and " << this->rowProperties.size() << " rows" << endl;
	}

	for (const auto& comp : screen->components) {
		if (comp->layoutConstraints.rowWeight > 0) {
			this->rowProperties.at(comp->layoutConstraints.row).weight = 1;
		}

		if (comp->layoutConstraints.colWeight > 0) {
			this->rowProperties.at(comp->layoutConstraints.col).weight = 1;
		}
	}
}

int getBiggestComponentInLine(Screen* screen, int index, bool isRow) {
	int biggestComponent = 0;

	for (const auto& comp : screen->components) {
		int cmp;

		if (isRow) {
			cmp = comp->layoutConstraints.row;
		} else {
			cmp = comp->layoutConstraints.col;
		}

		if (cmp == index) {
			int compSize;
			
			if (isRow) {
				compSize = comp->getMinimumHeight();
			} else {
				compSize = comp->getMinimumWidth();
			}

			if (compSize > biggestComponent) {
				biggestComponent = compSize;
			}
		}
	}

	return biggestComponent;
}

int getSparePixels(Screen* screen, map<int, GridLineProperties> gridLineProperties, const int usableSize, int& weightedComponentsSize) {
	int minimumSize = 0;

	for (auto& gridLine : gridLineProperties) {
		const int componentSize = getBiggestComponentInLine(screen, gridLine.first, true);

		minimumSize += componentSize;

		if (gridLine.second.weight == 1) {
			weightedComponentsSize += componentSize;
		}
	}

	return usableSize - minimumSize;
}

void positionOffsetLines(map<int, GridLineProperties>& gridLine) {
	int offset = 0;
	int pad = 6;

	for (auto& gridLine : gridLine) {
		gridLine.second.offset = offset;

		offset += gridLine.second.size + pad;
	}
}

void LayoutManager::doLayout(Screen* screen) {
	{
		int weightedComponentsHeight = 0;
		int sparePixelHeight = getSparePixels(screen, this->rowProperties, Renderer::get().window_h - (this->windowPadding * 2), weightedComponentsHeight);

		for (auto& row : this->rowProperties) {
			const int componentHeight = getBiggestComponentInLine(screen, row.first, true);

			int rowHeight = componentHeight;

			if (row.second.weight == 1) {
				float ratio = (float)componentHeight / (float)weightedComponentsHeight;

				rowHeight += ratio * sparePixelHeight;
			}

			row.second.size = rowHeight;
		}

		positionOffsetLines(this->rowProperties);
	}

	{
		int weightedComponentsWidth = 0;
		int sparePixelWidth = getSparePixels(screen, this->colProperties, Renderer::get().window_w - (this->windowPadding * 2), weightedComponentsWidth);

		if (cvarGetb("debug_layout_manager")) {
			cout << "layoutManager: spare pixel width" << sparePixelWidth << endl;
		}

		for (auto& col : this->colProperties) {
			const int componentWidth = getBiggestComponentInLine(screen, col.first, false);

			int colWidth = componentWidth;
			
			if (col.second.weight == 1) {
				float ratio = (float)componentWidth / (float)weightedComponentsWidth;
				
				colWidth += ratio * sparePixelWidth;
			}

			col.second.size = colWidth;
		}
		
		positionOffsetLines(this->colProperties);
	}

	this->applyLayoutToComponents(screen);
}

void LayoutManager::applyLayoutToComponents(Screen* screen) {
	if (cvarGetb("debug_layout_manager")) {
		this->debugLayout();
	}
		
	for (const auto& comp : screen->components) {
		if (cvarGetb("debug_layout_manager")) {
			cout << "Applying layout to " << comp->toString() << " which is in col " << comp->layoutConstraints.col << " and row " << comp->layoutConstraints.row << " " << endl;
		}

		auto col = this->colProperties.at(comp->layoutConstraints.col);
		auto row = this->rowProperties.at(comp->layoutConstraints.row);

		comp->pos.x = this->windowPadding + col.offset;
		comp->pos.y = this->windowPadding + row.offset;
		comp->pos.w = col.size;
		comp->pos.h = row.size;
	}
}

void LayoutManager::debugLayout() const {
	cout << "Completed layout, rows:" << this->rowProperties.size() << "\tcols:" << this->colProperties.size() << endl;

	for (const auto& row : this->rowProperties) {
		cout << "row " << row.first << " " << row.second.str() << endl;
	}

	for (const auto& col : this->colProperties) {
		cout << "col " << col.first << " " << col.second.str() << endl;
	}
}
