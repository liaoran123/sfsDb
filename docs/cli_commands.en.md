# CLI Command Reference

This document provides a detailed guide for using the sfsDb command-line tool, including all available commands and their functions.

## Overview

The sfsDb command-line tool provides the following core functions:

- **Database Status Management**: View current database status
- **System Information Management**: Get table, field, index and other system information
- **Index Management**: Analyze and manage database indexes
- **Configuration Management**: View and modify database configuration
- **Backup and Restore**: Create and restore database backups
- **Performance Statistics**: View query performance and hotspot data
- **Monitoring and Alerting**: Real-time database status monitoring
- **Web Interface Management**: Enable/disable and configure the web interface

## Basic Usage

### Command Format

```bash
sfsdb [command] [parameters]
```

### Global Parameters

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `--db` | Database path | `./kvdb` |
| `--help` | Show help information | - |

## Command Details

### 1. status - Get Database Status

**Function**: Get current database status, including memory usage and storage information.

**Usage Examples**:

```bash
# Get status of default path database
sfsdb status

# Get status of specified path database
sfsdb status --db ./mydb
```

**Output Example**:

```
=== Database Status ===
Memory Usage: 1.23 MB
Total Allocated: 2.45 MB
System Memory: 5.67 MB
GC Count: 10
Storage Type: LevelDB
```

### 2. system - System Information Management

**Function**: Get table, field, index and other system information.

**Usage Example**:

```bash
# Get system information
sfsdb system
```

### 3. index - Index Management

**Function**: Analyze and manage database indexes.

**Usage Example**:

```bash
# Manage indexes
sfsdb index
```

### 4. config - Configuration Management

**Function**: View and modify database configuration.

**Usage Example**:

```bash
# Manage configuration
sfsdb config
```

### 5. backup - Backup and Restore

**Function**: Create and restore database backups.

**Usage Example**:

```bash
# Manage backups
sfsdb backup
```

### 6. stats - Performance Statistics

**Function**: View query performance and hotspot data.

**Usage Example**:

```bash
# View performance statistics
sfsdb stats
```

### 7. monitor - Monitoring and Alerting

**Function**: Real-time database status monitoring.

**Usage Example**:

```bash
# Start monitoring
sfsdb monitor
```

### 8. web - Web Interface Management

**Function**: Enable/disable and configure the web interface.

**Subcommand Parameters**:

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| `--enable` `-e` | Enable web interface | false |
| `--disable` `-d` | Disable web interface | false |
| `--start` `-s` | Start web server | false |
| `--port` `-p` | Web server port | `:8083` |

**Usage Examples**:

```bash
# Enable web interface and set port to 8083
sfsdb web --enable --port :8083

# Disable web interface
sfsdb web --disable

# Start web server (using configured port)
sfsdb web --start

# Start web server and specify port (temporarily override configuration)
sfsdb web --start --port :8085

# View current web configuration
sfsdb web
```

**Output Example**:

```
Web Interface Configuration:
Enabled: true
Port: :8084
Address: http://localhost:8084
```

## Example Workflows

### 1. Start Web Interface

```bash
# Enable web interface and set port
sfsdb web --enable --port :8083

# Start web server
sfsdb web --start

# Access web interface
# http://localhost:8083
```

### 2. Backup Database

```bash
# Create database backup
sfsdb backup

# View database status
sfsdb status
```

### 3. Monitor Database

```bash
# Start monitoring
sfsdb monitor

# View performance statistics
sfsdb stats
```

## Summary

The sfsDb command-line tool provides comprehensive database management functions, allowing you to complete complex database management tasks with simple commands. Combined with the web interface, users can choose the appropriate management method according to their needs, improving database management efficiency.
