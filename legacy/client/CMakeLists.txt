cmake_minimum_required(VERSION 3.10)
project(Greyvar)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_EXTENSIONS OFF)

set(CMAKE_VERBOSE_MAKEFILE on)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY bin)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY bin)
set(BUILD_SHARED_LIBS off)

find_package(PkgConfig)
find_package(Protobuf REQUIRED)

include_directories(${CMAKE_SOURCE_DIR}/src/core/)

file(GLOB_RECURSE CORE_SOURCE_FILES src/core/*.cpp src/core/*.hpp)
add_library(greyvar_core ${CORE_SOURCE_FILES})

file(GLOB_RECURSE VESSEL_SOURCE_FILES src/vessel/*.cpp src/vessel/*.hpp)
add_executable(greyvar_vessel ${VESSEL_SOURCE_FILES})

message("Core source: ${CORE_SOURCE_FILES}")

if (WINDOWS)
	add_definitions(-DGREYVAR_PLATFORM_WINDOWS)
	set(PLATFORM_CXX_FLAGS "-s")

	target_include_directories(greyvar_core PUBLIC 
	)

    SET(CMAKE_FIND_LIBRARY_PREFIXES "lib" "")
    SET(CMAKE_FIND_LIBRARY_SUFFIXES ".dll" ".dll.a" ".a" ".lib")

	target_link_libraries(greyvar_core boleas m user32 gdi32 winmm)

	set(CMAKE_EXE_LINKER_FLAGS "-static-libgcc -static-libstdc++")

	set_target_properties(greyvar_core PROPERTIES LINK_SEARCH_END_STATIC 1)

	set_target_properties(greyvar_vessel PROPERTIES OUTPUT_NAME "greyvar_windows")
	target_link_libraries(greyvar_vessel greyvar_core)
elseif(UNIX)
	add_definitions(-DGREYVAR_PLATFORM_LINUX)

	if (CMAKE_BUILD_TYPE MATCHES Debug) 
		set(PLATFORM_CXX_FLAGS "-fsanitize=address  -static-libasan ")
		add_definitions(-D_GLIBCXX_DEBUG)
	else()
		set(PLATFORM_CXX_FLAGS "-s ")
	endif()

	target_link_libraries(greyvar_core boleas pthread SDL2 SDL2_mixer grpc++)

	set_target_properties(greyvar_vessel PROPERTIES OUTPUT_NAME "greyvar-client-linux")
	target_link_libraries(greyvar_vessel greyvar_core)
	
	file(GLOB_RECURSE TST_SOURCE_FILES tst/*.cpp tst/*.hpp)
	add_executable(greyvar_test ${TST_SOURCE_FILES})
	target_include_directories(greyvar_test PUBLIC
		src/core/
	)
	target_link_libraries(greyvar_test gtest greyvar_core)
	set_target_properties(greyvar_test PROPERTIES OUTPUT_NAME "test-greyvar-client-linux")
endif()

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wshadow-compatible-local -Wextra -pedantic ${PLATFORM_CXX_FLAGS}") # -Weffc++

