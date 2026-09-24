#include <sstream>

using std::stringstream;

struct GridLineProperties {
	int weight = 0;
	int size = 0;
	int offset = 0;
	int index = -1;
	int largestComponentSize = 0;

	GridLineProperties() {}

	string str() const {
		stringstream ss;
		ss << "GridLineProperties. index:" << this->index << " size: " << this->size << " offset:" << this->offset << " size:" << size << " (edge " << (this->offset + this->size) << ")";

		return ss.str();
	}
};


