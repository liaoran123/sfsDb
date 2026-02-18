# sfsDb C语言示例编译配置

# 编译器
CC = gcc
CXX = g++

# 编译选项
CFLAGS = -Wall -O2
CXXFLAGS = -Wall -O2

# 链接选项
LDFLAGS = -L../output/lib

# 依赖库
LIBS = -lsfsdb_linux_amd64

# 目标文件
TARGET = sfsdb_c_example

# 源文件
SRCS = main.c

# 头文件
HDRS = 

# 构建规则
all: $(TARGET)

$(TARGET): $(SRCS) $(HDRS)
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^ $(LIBS)

# 交叉编译规则
arm64: CC = aarch64-linux-gnu-gcc
arm64: LIBS = -lsfsdb_linux_arm64
arm64: $(TARGET)

mips64: CC = mips64-linux-gnuabi64-gcc
mips64: LIBS = -lsfsdb_linux_mips64
mips64: $(TARGET)

# 清理规则
clean:
	rm -f $(TARGET)
	rm -rf ./c_example_db

# 运行规则
run: $(TARGET)
	./$(TARGET)

# 调试规则
debug: CFLAGS += -g
debug: CXXFLAGS += -g
debug: $(TARGET)

# 安装规则
install: $(TARGET)
	@echo "Installing $(TARGET)..."
	@mkdir -p $(DESTDIR)/usr/bin
	@cp $(TARGET) $(DESTDIR)/usr/bin/

# 卸载规则
uninstall:
	@echo "Uninstalling $(TARGET)..."
	@rm -f $(DESTDIR)/usr/bin/$(TARGET)

.PHONY: all clean run debug install uninstall arm64 mips64

# 使用说明
# 1. 编译x86版本:
#    make
#
# 2. 编译ARM64版本:
#    make arm64
#
# 3. 编译MIPS64版本:
#    make mips64
#
# 4. 运行示例:
#    make run
#
# 5. 清理构建文件:
#    make clean
