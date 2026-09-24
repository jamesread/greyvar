clang-tidy --checks '*,-cppcoreguidelines-pro-type-union-access,-llvm-include-order,-fuchsia-default-arguments,-cppcoreguidelines*' src/core/*.cpp > rpt/clang-tidy.txt
grep -oP '^\/home(.*?)\:' rpt/clang-tidy.txt  | uniq -c | sort -n >> rpt/clang-tidy.txt
