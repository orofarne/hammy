project(hammy_worker_linux)

cmake_minimum_required(VERSION 2.6)

set(CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")

OPTION(TESTS "enable_testing" ON)

find_package(GLIB2)
find_package(EV)
find_package(MsgPack)
find_package(MOZJS)

include_directories(
  ${GLIB2_INCLUDE_DIRS}
  ${EV_INCLUDE_DIR}
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

add_library(worker_lib STATIC
  msgbuf.c
  router.c
  worker.c
  eval.cc
  child.cc
  child_bridge.cc
  fwriter.cc
)

add_executable(hammy_worker main.c)
add_dependencies(hammy_worker
  worker_lib
)
target_link_libraries(hammy_worker
  worker_lib
  ${GLIB2_LIBRARIES}
  ${EV_LIBRARY}
  ${MOZJS_LIBRARY}
  ${MsgPack_LIBRARIES}
)
