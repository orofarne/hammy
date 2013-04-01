project(hammy_worker)

cmake_minimum_required(VERSION 2.6)

set(CMAKE_MODULE_PATH "${CMAKE_SOURCE_DIR}/cmake")

find_package(MsgPack)
find_package(MOZJS)

set(WORKER_SOURCES
  main.cc
  worker.hh
  worker.cc
  fwriter.hh
  fwriter.cc
)

add_executable(worker ${WORKER_SOURCES})
target_link_libraries(worker ${MsgPack_LIBRARIES})