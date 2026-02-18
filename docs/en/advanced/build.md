# sfsDb Build Guide

## 1. Overview

This guide provides detailed instructions for building sfsDb, including pre-compilation solutions, cross-compilation configurations, Docker build environments, and C/C++ integration steps. By following this guide, you can quickly set up a build environment for sfsDb and generate pre-compiled libraries for industrial embedded systems.

## 2. Pre-compilation Solution

### 2.1 Download Pre-compiled Packages

sfsDb provides complete pre-compiled packages supporting multiple architectures:

- **x86_64 (amd64)**: For industrial PCs
- **ARM64 (arm64)**: For ARM development boards and industrial gateways
- **MIPS64 (mips64)**: For MIPS routers and embedded devices

### 2.2 Pre-compiled Package Contents

The pre-compiled package includes the following files:

```
sfsdb-prebuild-v1.0.0/
├── lib/                  # Static library files
│   ├── sfsdb_linux_amd64.a     # x86 industrial PC
│   ├── sfsdb_linux_arm64.a     # ARM64 gateway
│   └── sfsdb_linux_mips64.a    # MIPS64 router
├── demo/                 # Demo programs
│   ├── demo_linux_amd64        # x86 demo
│   ├── demo_linux_arm64        # ARM64 demo
│   └── demo_linux_mips64       # MIPS64 demo
├── examples/             # Example source code
│   ├── main.go                 # Go language example
│   ├── c_example/              # C language example
│   │   ├── main.c              # C example code
│   │   ├── Makefile            # Make configuration
│   │   └── CMakeLists.txt      # CMake configuration
│   └── performance_test.go     # Performance test code
├── RELEASE_NOTES.md      # Release notes
└── INTEGRATION_GUIDE.md  # Integration guide
```

## 3. Local Build

### 3.1 Environment Requirements

- **Go version**: 1.20 or higher
- **Operating system**: Linux, Windows
- **Cross-compilation toolchain** (optional): For ARM64 and MIPS64 architectures

### 3.2 Linux Build

#### Step 1: Clone the repository

```bash
git clone https://github.com/liaoran123/sfsDb.git
cd sfsDb
```

#### Step 2: Run the build script

```bash
# Run the build script
./build/build.sh

# Build results are in the build/output directory
ls -la build/output/
```

### 3.3 Windows Build

#### Step 1: Clone the repository

```powershell
git clone https://github.com/liaoran123/sfsDb.git
cd sfsDb
```

#### Step 2: Run the build script

```powershell
# Run the build script (using PowerShell)
powershell -ExecutionPolicy Bypass -File .\build\build.ps1

# Build results are in the build/output directory
dir .\build\output\
```

### 3.4 Build Dependencies

- **Static library build**: Requires `ar` command (provided by Git for Windows or MinGW-w64)
- **Cross-compilation**: Requires corresponding architecture cross-compilation toolchain

## 4. Docker Build Environment

### 4.1 Build Docker Image

```bash
# Build Docker image
docker build -t sfsdb-build-env .
```

### 4.2 Run Docker Container

```bash
# Run container for build
docker run -it --rm -v $(pwd):/app sfsdb-build-env bash -c "./build/build.sh"
```

### 4.3 Docker Environment Advantages

- **Consistency**: Ensures identical build results in any environment
- **Dependency management**: Built-in all necessary dependencies and toolchains
- **Isolation**: Does not affect local system environment

## 5. Cross-compilation Configuration

### 5.1 ARM64 Cross-compilation

```bash
# Set environment variables
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=0

# Build static library
go build -buildmode=c-archive -o lib/sfsdb_linux_arm64.a build/main.go

# Build demo program
go build -o demo/demo_linux_arm64 build/main.go
```

### 5.2 MIPS64 Cross-compilation

```bash
# Set environment variables
export GOOS=linux
export GOARCH=mips64
export CGO_ENABLED=0

# Build demo program (static library not supported)
go build -o demo/demo_linux_mips64 build/main.go
```

## 6. C/C++ Integration

### 6.1 Using Makefile Integration

#### Step 1: Copy library files

```bash
# Copy static library to project directory
cp sfsdb-prebuild-v1.0.0/lib/sfsdb_linux_arm64.a /path/to/your/project/lib/

# Copy header files (if needed)
cp sfsdb-prebuild-v1.0.0/examples/c_example/*.h /path/to/your/project/include/
```

#### Step 2: Configure Makefile

```makefile
# sfsDb integration configuration
SFSDB_LIB_DIR = ./lib
SFSDB_LIB = -lsfsdb_linux_arm64

# Compilation options
CFLAGS += -I./include
LDFLAGS += -L$(SFSDB_LIB_DIR) $(SFSDB_LIB)

# Target file
target: main.o
	$(CC) $(LDFLAGS) -o $@ $^
```

#### Step 3: Compile project

```bash
# Compile project
make

# Run program
./target
```

### 6.2 Using CMake Integration

#### Step 1: Copy library files

```bash
# Copy static library to project directory
cp sfsdb-prebuild-v1.0.0/lib/sfsdb_linux_arm64.a /path/to/your/project/lib/
```

#### Step 2: Configure CMakeLists.txt

```cmake
# sfsDb integration configuration
set(SFSDB_LIB_DIR "${CMAKE_SOURCE_DIR}/lib")
set(SFSDB_LIBRARY "sfsdb_linux_arm64")

# Link sfsDb library
target_link_libraries(your_target
    PRIVATE
        -L${SFSDB_LIB_DIR}
        -l${SFSDB_LIBRARY}
)
```

#### Step 3: Compile project

```bash
# Create build directory
mkdir -p build && cd build

# Configure CMake
cmake ..

# Compile project
cmake --build .

# Run program
./your_target
```

## 7. Performance Optimization

### 7.1 Memory Optimization

- **Use batch operations**: For large data inserts, use batch operations to reduce memory usage
- **Set cache size appropriately**: Adjust cache parameters based on device memory
- **Release resources promptly**: Close the database when done to avoid memory leaks

### 7.2 Speed Optimization

- **Create appropriate indexes**: Create indexes for frequently queried fields
- **Use primary key queries**: Primary key queries are fastest, use them whenever possible
- **Reduce concurrent operations**: On resource-constrained devices, control concurrency appropriately

### 7.3 Stability Optimization

- **Use transactions**: For critical operations, use transactions to ensure data consistency
- **Regular backups**: Backup the database regularly to prevent data loss
- **Error handling**: Implement comprehensive error handling to improve system stability

## 8. Common Problem Solutions

### 8.1 Compilation Errors

#### Problem: undefined reference to `xxx`

**Solution**:
- Check if library files are linked correctly
- Ensure you're using the correct architecture library
- Verify compilation command includes `-L` and `-l` options

#### Problem: cannot find -lsfsdb_linux_arm64

**Solution**:
- Confirm library file exists in the specified directory
- Check library file permissions
- Ensure you're using the correct library file name

### 8.2 Runtime Errors

#### Problem: Database initialization failed

**Solution**:
- Check if database path is writable
- Ensure device has sufficient storage space
- Verify file system permissions

#### Problem: Out of memory

**Solution**:
- Reduce batch operation size
- Increase device memory if possible
- Optimize data structures to reduce memory usage

### 8.3 Performance Issues

#### Problem: Slow query speed

**Solution**:
- Create indexes for query fields
- Use primary key queries
- Reduce amount of data returned by queries

#### Problem: Slow insertion speed

**Solution**:
- Use batch insertion
- Reduce number of indexes
- Optimize data structures

## 9. CI/CD Integration

### 9.1 GitHub Actions Configuration

sfsDb provides complete GitHub Actions configuration for automatic building and releasing pre-compiled packages:

- **Automatic cross-compilation**: Executes multi-architecture cross-compilation automatically on each code commit
- **Generate pre-compiled packages**: Automatically generates pre-compiled packages for all architectures
- **Publish to GitHub Release**: Automatically uploads pre-compiled packages to the Release page

### 9.2 Local CI Configuration

If you need to configure a build process in a local or private CI system, you can follow these steps:

1. **Install dependencies**: Go 1.20+, cross-compilation toolchain
2. **Configure environment variables**: Set `GOOS`, `GOARCH`, etc.
3. **Execute build script**: Run `./build/build.sh` or `./build/build.ps1`
4. **Upload build artifacts**: Upload pre-compiled packages to the specified location

## 10. Version Management

### 10.1 Version Number Format

sfsDb uses semantic versioning format:

```
v{major}.{minor}.{patch}
```

- **Major version**: Incompatible API changes
- **Minor version**: Backward-compatible feature additions
- **Patch version**: Backward-compatible bug fixes

### 10.2 Version Upgrade

- **Patch version upgrade**: Simply replace library files
- **Minor version upgrade**: May require updating header files, API remains compatible
- **Major version upgrade**: Requires updating code to adapt to new API

## 11. Technical Support

### 11.1 Problem Feedback

If you encounter issues, you can provide feedback through:

- **GitHub Issues**: https://github.com/liaoran123/sfsDb/issues
- **Email**: support@sfsdb.io
- **Community**: https://github.com/liaoran123/sfsDb/discussions

### 11.2 Common Questions

**Q**: Which operating systems does sfsDb support?
**A**: sfsDb primarily supports Linux systems, including x86, ARM64, and MIPS64 architectures.

**Q**: What scenarios is sfsDb suitable for?
**A**: sfsDb is suitable for industrial embedded systems, edge computing devices, IoT gateways, and other resource-constrained scenarios.

**Q**: Does sfsDb support transactions?
**A**: Yes, sfsDb supports complete ACID transactions, including four isolation levels.

**Q**: How to get the latest version of sfsDb?
**A**: You can download the latest pre-compiled packages from the GitHub Release page or build from source.

## 12. Summary

sfsDb provides a complete build solution supporting multi-architecture cross-compilation, C/C++ integration, and Docker build environments. By following the steps in this guide, you can quickly build and integrate sfsDb for industrial embedded systems, enjoying its high performance and reliability.

If you have any questions or suggestions, please feel free to provide feedback. We will continue to improve sfsDb to provide better database solutions for the industrial embedded field.