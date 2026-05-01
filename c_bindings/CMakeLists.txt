cmake_minimum_required(VERSION 3.10)
project(sfsDb_C_Example)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# 编译 Go 代码为静态库
add_custom_command(
    OUTPUT libsfsdb.a libsfsdb.h
    COMMAND go build -buildmode=c-archive -o libsfsdb.a c_bindings.go
    DEPENDS c_bindings.go
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
)

add_custom_target(go_lib DEPENDS libsfsdb.a libsfsdb.h)

# 创建静态库目标
add_library(sfsdb STATIC IMPORTED GLOBAL)
set_target_properties(sfsdb PROPERTIES
    IMPORTED_LOCATION ${CMAKE_CURRENT_SOURCE_DIR}/libsfsdb.a
)
target_link_libraries(sfsdb INTERFACE pthread)
add_dependencies(sfsdb go_lib)

# 编译 C 示例
add_executable(main main.c)
target_link_libraries(main PRIVATE sfsdb)

# 运行示例的目标
add_custom_target(run
    COMMAND main
    DEPENDS main
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
)

# 清理目标
add_custom_target(clean_all
    COMMAND ${CMAKE_COMMAND} -E remove libsfsdb.a libsfsdb.h main
    COMMAND ${CMAKE_COMMAND} -E remove_directory c_example_db
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
)
