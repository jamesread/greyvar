#include <gtest/gtest.h>

#include "Grid.hpp"

#include <string>

auto grid = new Grid(3, 4);

TEST(Grid, ConstructorArgumentSizes) {
	ASSERT_EQ(grid->w, 3);
	ASSERT_EQ(grid->h, 4);
}

