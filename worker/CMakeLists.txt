project(hammy_worker)

cmake_minimum_required(VERSION 2.6)

set(CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")

OPTION(TESTS "enable_testing" ON)

find_package(MsgPack)
find_package(MOZJS)

include_directories(
  ${MsgPack_INCLUDE_DIRS}
  ${MOZJS_INCLUDE_DIR}
)

if(TESTS)
  find_package(Threads)
  find_package(GTest)
  include_directories(${GTEST_INCLUDE_DIR})

  add_custom_target(test)
  macro(run_test test_target)
    add_custom_target(${test_target}_runtest
        COMMAND ${test_target}
        DEPENDS ${test_target}
        WORKING_DIRECTORY "${CMAKE_BINARY_DIR}")
    add_dependencies(test ${test_target}_runtest)
  endmacro()

  add_subdirectory(tests)
endif(TESTS)

set(WORKER_LIB_SOURCES
  eval.cc
  fwriter.cc
  worker.cc
)

add_library(worker_lib STATIC ${WORKER_LIB_SOURCES})

add_executable(worker main.cc)
add_dependencies(worker worker_lib)
target_link_libraries(worker
  worker_lib
  ${MsgPack_LIBRARIES}
  ${MOZJS_LIBRARY}
)