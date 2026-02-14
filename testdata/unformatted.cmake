CMAKE_MINIMUM_REQUIRED( VERSION  3.20 )

PROJECT(  myproject   VERSION 1.0.0
    LANGUAGES CXX )


SET(SOURCES
  src/main.cpp
  src/utils.cpp
  src/config.cpp
        )

ADD_EXECUTABLE( myproject ${SOURCES} )

TARGET_LINK_LIBRARIES(myproject
    public  fmt::fmt
    private  spdlog::spdlog
)

IF(CMAKE_BUILD_TYPE STREQUAL "Debug")
    TARGET_COMPILE_OPTIONS(myproject private -Wall -Wextra -Wpedantic)
ELSE()
  MESSAGE(STATUS "Release build")
ENDIF()

# Find dependencies
FIND_PACKAGE(fmt required)
FIND_PACKAGE(spdlog required CONFIG)

INSTALL(targets myproject
    destination ${CMAKE_INSTALL_BINDIR}
)
