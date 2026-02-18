# sfsDb C语言示例 CMake配置
cmake_minimum_required(VERSION 3.10)

project(sfsdb_c_example C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)

# 设置编译选项
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -Wall -O2")

# 查找sfsDb静态库
set(SFSDB_LIB_DIR "${CMAKE_SOURCE_DIR}/../output/lib")

# 根据目标架构选择库文件
if(CMAKE_SYSTEM_PROCESSOR STREQUAL "aarch64")
    set(SFSDB_LIBRARY "sfsdb_linux_arm64")
elif(CMAKE_SYSTEM_PROCESSOR STREQUAL "mips64")
    set(SFSDB_LIBRARY "sfsdb_linux_mips64")
else()
    set(SFSDB_LIBRARY "sfsdb_linux_amd64")
endif()

# 添加可执行文件
add_executable(sfsdb_c_example main.c)

# 链接sfsDb库
target_link_libraries(sfsdb_c_example
    PRIVATE
        -L${SFSDB_LIB_DIR}
        -l${SFSDB_LIBRARY}
)

# 设置输出目录
set_target_properties(sfsdb_c_example PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin"
)

# 安装规则
install(TARGETS sfsdb_c_example
    RUNTIME DESTINATION bin
)

# 清理规则
add_custom_target(clean-example
    COMMAND ${CMAKE_COMMAND} -E remove_directory "${CMAKE_BINARY_DIR}/bin"
    COMMAND ${CMAKE_COMMAND} -E remove_directory "${CMAKE_SOURCE_DIR}/c_example_db"
    COMMENT "Cleaning example build files"
)

# 运行规则
add_custom_target(run-example
    COMMAND ${CMAKE_BINARY_DIR}/bin/sfsdb_c_example
    DEPENDS sfsdb_c_example
    COMMENT "Running sfsDb C example"
)

# 交叉编译配置示例
# 1. ARM64交叉编译:
#    cmake -DCMAKE_TOOLCHAIN_FILE=arm64.toolchain.cmake ..
#
# 2. MIPS64交叉编译:
#    cmake -DCMAKE_TOOLCHAIN_FILE=mips64.toolchain.cmake ..
