cmake_minimum_required(VERSION 2.8.12)
project(Greyvar-qml-launcher)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_EXTENSIONS OFF)

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY bin)

set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTOUIC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_INCLUDE_CURRENT_DIR ON)

add_executable(greyvar-qml-launcher src/BackEnd.hpp src/main.cpp  src/Updater.hpp)

include_directories(${Qt5Widgets_INCLUDE_DIRS} ${QtQml_INCLUDE_DIRS})
add_definitions(${Qt5Widgets_DEFINITIONS} ${QtQml_DEFINITIONS} ${${Qt5Quick_DEFINITIONS}})

find_package(Qt5 COMPONENTS Widgets Qml Quick REQUIRED)

target_link_libraries(greyvar-qml-launcher Qt5::Widgets Qt5::Qml Qt5::Quick curl pthread)

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wshadow-compatible-local -Wextra -pedantic")


