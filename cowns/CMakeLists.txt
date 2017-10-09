cmake_minimum_required(VERSION 2.8.4)

project(owns CXX C)

set(VERBOSE 1)

if(NOT CMAKE_BUILD_TYPE)
set(CMAKE_BUILD_TYPE Debug)
endif()

set(CMAKE_C_FLAGS_DEBUG "-std=gnu11 -O0 -g")
set(CMAKE_CXX_FLAGS_DEBUG "-std=gnu++11 -O0 -g")

set(CMAKE_C_FLAGS_RELEASE "-std=gnu11 -Os -flto")
set(CMAKE_CXX_FLAGS_RELEASE "-std=gnu++11 -Os -flto")

include_directories(
    .
    src/
    ${CMAKE_CURRENT_BINARY_DIR}
)

set(LIB_SOURCES
    src/owns.c
    src/protocol.c
    ${CMAKE_CURRENT_BINARY_DIR}/ons.pb-c.c
)

set(CLIENT_SOURCES
    src/main.c    
)

add_custom_command(OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/ons.pb-c.c ${CMAKE_CURRENT_BINARY_DIR}/ons.pb-c.h
                   COMMAND protoc-c --c_out=${CMAKE_CURRENT_BINARY_DIR} 
                                --proto_path=${CMAKE_CURRENT_LIST_DIR}/../protocol/
                                ${CMAKE_CURRENT_LIST_DIR}/../protocol/*.proto)

add_library(owns ${LIB_SOURCES})
target_link_libraries(owns zmq czmq protobuf protobuf-c pthread)

add_executable(owns-client ${CLIENT_SOURCES})
target_link_libraries(owns-client ${OPTIONAL_LIBS} owns)

set(CMAKE_INSTALL_PATH /usr/local)
install(TARGETS owns DESTINATION lib/)
file(GLOB HEADERS owns/**.h)
install(FILES ${HEADERS} DESTINATION include/${PROJECT_NAME})


