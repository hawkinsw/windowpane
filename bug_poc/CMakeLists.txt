cmake_minimum_required(VERSION 3.20)

project(BugPoc DESCRIPTION "Demonstration of https://github.com/golang/go/issues/58870")


if(NOT WIN32)
    message(FATAL_ERROR "Compatible only with Windows.")
endif()

add_executable(BugPoc bug_poc.cpp)
target_link_libraries(BugPoc wsock32 ws2_32)
set_target_properties(BugPoc PROPERTIES CXX_STANDARD 20)
