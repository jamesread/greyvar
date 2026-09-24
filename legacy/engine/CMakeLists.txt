cmake_minimum_required(VERSION 3.10)
project(boleas VERSION 1.0 DESCRIPTION "boleas game library")

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_EXTENSIONS OFF)

set(CMAKE_VERBOSE_MAKEFILE on)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY bin)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY bin)
set(BUILD_SHARED_LIBS off)

find_package(PkgConfig)
#find_package(SDL2 REQUIRED)
#find_package(SDL2_ttf REQUIRED)

include_directories(${CMAKE_SOURCE_DIR}/src/ ${CMAKE_SOURCE_DIR}/pub/)

file(GLOB_RECURSE PUBLIC_HEADERS pub/*.hpp pub/net/proto/*.h)
file(GLOB_RECURSE SOURCE_FILES src/*.cpp pub/net/proto/*.cc)
add_library(boleas SHARED ${SOURCE_FILES})

set_target_properties(boleas PROPERTIES VERSION ${PROJECT_VERSION})
set_target_properties(boleas PROPERTIES SOVERSION 1)

include(GNUInstallDirs)

install(TARGETS boleas 
	LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
	ARCHIVE DESTINATION ${CMAKE_INSTALL_LIBDIR}
)

install(
	DIRECTORY pub/ DESTINATION "${CMAKE_INSTALL_INCLUDEDIR}/boleas"
)

configure_file(boleas.pc.in boleas.pc @ONLY)

install(FILES ${CMAKE_BINARY_DIR}/boleas.pc DESTINATION ${CMAKE_INSTALL_DATAROOTDIR}/pkgconfig)

add_definitions(-DPROJECT_VERSION="${PROJECT_VESION}")

if (WINDOWS)
	target_link_libraries(boleas mingw32 SDL2main SDL2 m user32 gdi32 winmm)  #  gRPC::grpc++_reflection protobuf::libprotobuf grpc++
	set_target_properties(boleas PROPERTIES LINK_SEARCH_END_STATIC 1)
elseif(UNIX)
	add_definitions(-DPLATFORM_LINUX)

	if (CMAKE_BUILD_TYPE MATCHES Debug) 
		set(PLATFORM_CXX_FLAGS "-s -fsanitize=address ")
		add_definitions(-D_GLIBCXX_DEBUG)
	else()
		set(PLATFORM_CXX_FLAGS "-s ")
	endif()

	target_include_directories(boleas PRIVATE /usr/include/freetype2 PUBLIC include)

	target_link_libraries(boleas SDL2 SDL2_image freetype SDL2_mixer grpc++ protobuf ossp-uuid++)
endif()

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wshadow-compatible-local -Wextra -pedantic ${PLATFORM_CXX_FLAGS} ")

#include_directories(src ${SDL2_INCLUDE_DIRS} ${SDL2_TTF_INCLUDE_DIRS})
#link_directories(${SDL2_INCLUDE_DIRS} ${SDL2_TTF_INCLUDE_DIRS})
