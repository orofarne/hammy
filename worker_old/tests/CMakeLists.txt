file(GLOB WORKER_TEST_SOURCES "*.cc")
add_executable(worker_test ${WORKER_TEST_SOURCES})
message(STATUS ${GTEST_BOTH_LIBRARIES})
target_link_libraries(worker_test
  ${GTEST_BOTH_LIBRARIES}
  ${CMAKE_THREAD_LIBS_INIT}
)
add_dependencies(worker_test worker_lib)
target_link_libraries(worker_test
  worker_lib
  ${MOZJS_LIBRARY}
)

run_test(worker_test)