file(GLOB WORKER_TEST_SOURCES "*.cc")
add_executable(worker_test ${WORKER_TEST_SOURCES})
add_dependencies(worker_test worker_lib)
target_link_libraries(worker_test
  worker_lib
  ${GTEST_BOTH_LIBRARIES}
  ${CMAKE_THREAD_LIBS_INIT}
  ${GLIB2_LIBRARIES}
)

run_test(worker_test)
