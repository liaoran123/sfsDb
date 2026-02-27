#!/usr/bin/env python3
"""
性能测试数据提取和图表生成脚本
"""

import os
import json
import matplotlib.pyplot as plt
import pandas as pd

# 性能测试数据
performance_data = {
    "database_comparison": {
        "dbs": ["sfsDb", "SQLite", "BoltDB", "RocksDB", "BadgerDB"],
        "primary_key_search": [18.6, 100, 27.5, 20, 30],  # 微秒/次
        "insert_performance": [29.9, 75, 30, 22.5, 35],    # 微秒/次
        "batch_insert": [12.6, 35, 11.5, 7.5, 12.5]       # 微秒/条
    },
    "acid_vs_non_acid": {
        "operations": ["单次插入", "批量插入", "主键搜索", "数据更新"],
        "non_acid": [29.923, 126.404, 18.557, 43.879],    # 微秒/次
        "acid": [32.421, 189.980, 22.248, 42.801]         # 微秒/次
    },
    "data_volume_impact": {
        "data_volumes": ["1,000", "10,000", "100,000", "1,000,000"],
        "primary_key_search": [18.6, 20.3, 23.3, 31.5],   # 微秒/次
        "insert_performance": [29.9, 32.5, 40.2, 50.5]     # 微秒/次
    },
    "concurrency_impact": {
        "concurrency": ["1-10", "10-100", "100-1000", "1000+"],
        "primary_key_search": [18.6, 21.8, 30.2, 42.5],   # 微秒/次
        "insert_performance": [29.9, 32.5, 40.2, 60.5]     # 微秒/次
    }
}

# 生成数据库性能比较图
def generate_database_comparison_chart():
    """生成数据库性能比较图表"""
    data = performance_data["database_comparison"]
    
    fig, ax = plt.subplots(figsize=(12, 6))
    
    x = range(len(data["dbs"]))
    width = 0.25
    
    ax.bar([i - width for i in x], data["primary_key_search"], width, label="主键搜索 (微秒/次)")
    ax.bar(x, data["insert_performance"], width, label="单次插入 (微秒/次)")
    ax.bar([i + width for i in x], data["batch_insert"], width, label="批量插入 (微秒/条)")
    
    ax.set_xlabel('数据库')
    ax.set_ylabel('性能 (微秒)')
    ax.set_title('sfsDb 与其他嵌入式数据库性能比较')
    ax.set_xticks(x)
    ax.set_xticklabels(data["dbs"])
    ax.legend()
    
    # 保存图表
    output_dir = "docs/performance"
    os.makedirs(output_dir, exist_ok=True)
    plt.savefig(f"{output_dir}/database_comparison.png", dpi=150, bbox_inches='tight')
    plt.close()

# 生成ACID vs 非ACID性能比较图
def generate_acid_comparison_chart():
    """生成ACID vs 非ACID性能比较图表"""
    data = performance_data["acid_vs_non_acid"]
    
    fig, ax = plt.subplots(figsize=(12, 6))
    
    x = range(len(data["operations"]))
    width = 0.35
    
    ax.bar([i - width/2 for i in x], data["non_acid"], width, label="非ACID模式")
    ax.bar([i + width/2 for i in x], data["acid"], width, label="ACID事务模式")
    
    ax.set_xlabel('操作类型')
    ax.set_ylabel('性能 (微秒/次)')
    ax.set_title('ACID vs 非ACID模式性能比较')
    ax.set_xticks(x)
    ax.set_xticklabels(data["operations"])
    ax.legend()
    
    # 保存图表
    output_dir = "docs/performance"
    os.makedirs(output_dir, exist_ok=True)
    plt.savefig(f"{output_dir}/acid_comparison.png", dpi=150, bbox_inches='tight')
    plt.close()

# 生成数据量影响图
def generate_data_volume_chart():
    """生成数据量增长对性能的影响图表"""
    data = performance_data["data_volume_impact"]
    
    fig, ax = plt.subplots(figsize=(12, 6))
    
    ax.plot(data["data_volumes"], data["primary_key_search"], marker='o', label="主键搜索 (微秒/次)")
    ax.plot(data["data_volumes"], data["insert_performance"], marker='s', label="插入性能 (微秒/次)")
    
    ax.set_xlabel('数据量 (条)')
    ax.set_ylabel('性能 (微秒)')
    ax.set_title('数据量增长对性能的影响')
    ax.legend()
    
    # 保存图表
    output_dir = "docs/performance"
    os.makedirs(output_dir, exist_ok=True)
    plt.savefig(f"{output_dir}/data_volume_impact.png", dpi=150, bbox_inches='tight')
    plt.close()

# 生成并发影响图
def generate_concurrency_chart():
    """生成并发增长对性能的影响图表"""
    data = performance_data["concurrency_impact"]
    
    fig, ax = plt.subplots(figsize=(12, 6))
    
    ax.plot(data["concurrency"], data["primary_key_search"], marker='o', label="主键搜索 (微秒/次)")
    ax.plot(data["concurrency"], data["insert_performance"], marker='s', label="插入性能 (微秒/次)")
    
    ax.set_xlabel('并发数')
    ax.set_ylabel('性能 (微秒)')
    ax.set_title('并发增长对性能的影响')
    ax.legend()
    
    # 保存图表
    output_dir = "docs/performance"
    os.makedirs(output_dir, exist_ok=True)
    plt.savefig(f"{output_dir}/concurrency_impact.png", dpi=150, bbox_inches='tight')
    plt.close()

# 主函数
def main():
    """主函数"""
    print("生成性能测试图表...")
    generate_database_comparison_chart()
    generate_acid_comparison_chart()
    generate_data_volume_chart()
    generate_concurrency_chart()
    print("图表生成完成！")

if __name__ == "__main__":
    main()
