#include "Tile.hpp"

Tile::Tile(const string& textureName, bool traversable) {
	this->textureName = textureName;
	this->traversable = traversable;
}
