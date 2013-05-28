add_custom_command(
  OUTPUT cases_mpac.o
  COMMAND ld -r -b binary -o ${CMAKE_CURRENT_BINARY_DIR}/cases_mpac.o cases.mpac
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  DEPENDS cases.mpac
)
add_custom_command(
  OUTPUT libtests_res.a
  COMMAND ar rc libtests_res.a ${CMAKE_CURRENT_BINARY_DIR}/cases_mpac.o
  DEPENDS cases_mpac.o
)
add_custom_target(tests_res DEPENDS libtests_res.a)

file(GLOB WORKER_TEST_SOURCES "*.cc")
add_executable(worker_test ${WORKER_TEST_SOURCES})
set_target_properties(worker_test PROPERTIES
  COMPILE_FLAGS -Wno-conversion-null)
add_dependencies(worker_test
  worker_lib
  tests_res
)
target_link_libraries(worker_test
  worker_lib
  ${CMAKE_CURRENT_BINARY_DIR}/libtests_res.a
  ${GTEST_BOTH_LIBRARIES}
  ${CMAKE_THREAD_LIBS_INIT}
  ${GLIB2_LIBRARIES}
  ${EV_LIBRARY}
  ${MOZJS_LIBRARY}
  ${MsgPack_LIBRARIES}
)

run_test(worker_test)
