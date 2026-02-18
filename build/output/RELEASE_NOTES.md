# sfsDb v1.0.0 Prebuild Package

## Included Files

### Static Library Files
- lib/sfsdb_linux_amd64.a (x86 Industrial PC)
- lib/sfsdb_linux_arm64.a (ARM64 Gateway)
- lib/sfsdb_linux_mips64.a (MIPS64 Router)

### Demo Programs
- demo/demo_linux_amd64 (x86 Example)
- demo/demo_linux_arm64 (ARM64 Example)
- demo/demo_linux_mips64 (MIPS64 Example)

### Example Source Code
- examples/main.go (Complete example code)
- examples/performance_test.go (Performance test code)
- examples/c_example/ (C language example)
- examples/README.md (Usage instructions)

## Usage

### Run Demo Program
`ash
# ARM64 Platform
./demo/demo_linux_arm64
`

### Integrate with C/C++ Project
`ash
# ARM64 Platform
gcc main.c -L./lib -l:lib/sfsdb_linux_arm64.a -o app
`

## Supported Architectures
- x86_64 (amd64)
- ARM64 (arm64)
- MIPS64 (mips64)

## Performance Metrics
- Memory Usage: < 5MB
- Startup Time: < 100ms
- Concurrency Support: 1000+ concurrent connections

## Security Features
- Data Encryption: AES-256
- Transaction Support: ACID compatible
- Permission Control: Fine-grained permission management
