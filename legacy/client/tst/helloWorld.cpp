#include <gtest/gtest.h>

#include <string>

using std::string;

const char *expectVal = "hello gtest";
const char *actualValTrue = "hello gtest";

TEST(StrCompare, CStrEqual) {
	EXPECT_STREQ(expectVal, actualValTrue);
}
