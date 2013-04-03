# - Find mozjs
# Find the SpiderMonkey library
#
# This module defines
# MOZJS_LIBRARY
# MOZJS_FOUND, if false, do not try to link to nvtt
# MOZJS_INCLUDE_DIR, where to find the headers
#

FIND_PATH(MOZJS_INCLUDE_DIR js/jsapi.h
  PATHS
  /usr/local
  /usr
  $ENV{MOZJS_DIR}
  PATH_SUFFIXES include include-unix
)

FIND_LIBRARY(MOZJS_LIBRARY
  NAMES mozjs185
  PATHS
  /usr/local
  /usr
  $ENV{MOZJS_DIR}
  PATH_SUFFIXES lib64 lib lib/shared lib/static lib64/static
)

SET(MOZJS_FOUND "NO")
IF(MOZJS_LIBRARY AND MOZJS_INCLUDE_DIR)
  MESSAGE(STATUS "MOZJS found")
  SET(MOZJS_FOUND "YES")
ENDIF(MOZJS_LIBRARY AND MOZJS_INCLUDE_DIR)
