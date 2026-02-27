#!/usr/bin/env python3
"""
简单的性能测试图表生成脚本
使用内置库生成SVG图表
"""

import os
import json

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

# 生成SVG图表
def generate_svg_chart(chart_type, data, output_path):
    """生成SVG格式的图表"""
    if chart_type == "database_comparison":
        # 生成数据库比较图表
        svg_content = f"""
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="400" fill="#f9f9f9"/>
    <text x="400" y="30" font-size="20" text-anchor="middle" font-weight="bold">sfsDb 与其他嵌入式数据库性能比较</text>
    <text x="400" y="380" font-size="12" text-anchor="middle">性能 (微秒)</text>
"""
        
        # 绘制柱状图
        bar_width = 50
        gap = 20
        start_x = 100
        
        for i, db in enumerate(data["dbs"]):
            x = start_x + i * (bar_width * 3 + gap)
            
            # 主键搜索
            height = data["primary_key_search"][i] * 2
            svg_content += f"<rect x='{x}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#4CAF50'/>"
            
            # 单次插入
            height = data["insert_performance"][i] * 2
            svg_content += f"<rect x='{x + bar_width + 5}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#2196F3'/>"
            
            # 批量插入
            height = data["batch_insert"][i] * 2
            svg_content += f"<rect x='{x + bar_width * 2 + 10}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#FF9800'/>"
            
            # 数据库名称
            svg_content += f"<text x='{x + bar_width}' y='370' font-size='12' text-anchor='middle'>{db}</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="50" width="20" height="20" fill="#4CAF50"/>
    <text x="630" y="65" font-size="12">主键搜索</text>
    <rect x="600" y="80" width="20" height="20" fill="#2196F3"/>
    <text x="630" y="95" font-size="12">单次插入</text>
    <rect x="600" y="110" width="20" height="20" fill="#FF9800"/>
    <text x="630" y="125" font-size="12">批量插入</text>
</svg>
"""
    
    elif chart_type == "acid_comparison":
        # 生成ACID比较图表
        svg_content = f"""
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="400" fill="#f9f9f9"/>
    <text x="400" y="30" font-size="20" text-anchor="middle" font-weight="bold">ACID vs 非ACID模式性能比较</text>
    <text x="400" y="380" font-size="12" text-anchor="middle">性能 (微秒/次)</text>
"""
        
        # 绘制柱状图
        bar_width = 60
        gap = 40
        start_x = 150
        
        for i, op in enumerate(data["operations"]):
            x = start_x + i * (bar_width * 2 + gap)
            
            # 非ACID模式
            height = data["non_acid"][i] * 1.5
            svg_content += f"<rect x='{x}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#4CAF50'/>"
            
            # ACID模式
            height = data["acid"][i] * 1.5
            svg_content += f"<rect x='{x + bar_width + 10}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#2196F3'/>"
            
            # 操作名称
            svg_content += f"<text x='{x + bar_width + 5}' y='370' font-size='12' text-anchor='middle'>{op}</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="50" width="20" height="20" fill="#4CAF50"/>
    <text x="630" y="65" font-size="12">非ACID模式</text>
    <rect x="600" y="80" width="20" height="20" fill="#2196F3"/>
    <text x="630" y="95" font-size="12">ACID事务模式</text>
</svg>
"""
    
    elif chart_type == "data_volume_impact":
        # 生成数据量影响图表
        svg_content = f"""
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="400" fill="#f9f9f9"/>
    <text x="400" y="30" font-size="20" text-anchor="middle" font-weight="bold">数据量增长对性能的影响</text>
    <text x="400" y="380" font-size="12" text-anchor="middle">性能 (微秒)</text>
"""
        
        # 绘制折线图
        points = []
        for i, val in enumerate(data["primary_key_search"]):
            x = 100 + i * 200
            y = 350 - val * 5
            points.append(f"{x},{y}")
        
        # 主键搜索折线
        svg_content += f"<polyline points='{' '.join(points)}' fill='none' stroke='#4CAF50' stroke-width='2'/>"
        
        # 绘制数据点
        for i, val in enumerate(data["primary_key_search"]):
            x = 100 + i * 200
            y = 350 - val * 5
            svg_content += f"<circle cx='{x}' cy='{y}' r='4' fill='#4CAF50'/>"
        
        # 插入性能折线
        points = []
        for i, val in enumerate(data["insert_performance"]):
            x = 100 + i * 200
            y = 350 - val * 3
            points.append(f"{x},{y}")
        
        svg_content += f"<polyline points='{' '.join(points)}' fill='none' stroke='#2196F3' stroke-width='2'/>"
        
        # 绘制数据点
        for i, val in enumerate(data["insert_performance"]):
            x = 100 + i * 200
            y = 350 - val * 3
            svg_content += f"<circle cx='{x}' cy='{y}' r='4' fill='#2196F3'/>"
        
        # X轴标签
        for i, volume in enumerate(data["data_volumes"]):
            x = 100 + i * 200
            svg_content += f"<text x='{x}' y='370' font-size='12' text-anchor='middle'>{volume}</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="50" width="20" height="20" fill="#4CAF50"/>
    <text x="630" y="65" font-size="12">主键搜索</text>
    <rect x="600" y="80" width="20" height="20" fill="#2196F3"/>
    <text x="630" y="95" font-size="12">插入性能</text>
</svg>
"""
    
    elif chart_type == "concurrency_impact":
        # 生成并发影响图表
        svg_content = f"""
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="400" fill="#f9f9f9"/>
    <text x="400" y="30" font-size="20" text-anchor="middle" font-weight="bold">并发增长对性能的影响</text>
    <text x="400" y="380" font-size="12" text-anchor="middle">性能 (微秒)</text>
"""
        
        # 绘制折线图
        points = []
        for i, val in enumerate(data["primary_key_search"]):
            x = 100 + i * 200
            y = 350 - val * 4
            points.append(f"{x},{y}")
        
        # 主键搜索折线
        svg_content += f"<polyline points='{' '.join(points)}' fill='none' stroke='#4CAF50' stroke-width='2'/>"
        
        # 绘制数据点
        for i, val in enumerate(data["primary_key_search"]):
            x = 100 + i * 200
            y = 350 - val * 4
            svg_content += f"<circle cx='{x}' cy='{y}' r='4' fill='#4CAF50'/>"
        
        # 插入性能折线
        points = []
        for i, val in enumerate(data["insert_performance"]):
            x = 100 + i * 200
            y = 350 - val * 2.5
            points.append(f"{x},{y}")
        
        svg_content += f"<polyline points='{' '.join(points)}' fill='none' stroke='#2196F3' stroke-width='2'/>"
        
        # 绘制数据点
        for i, val in enumerate(data["insert_performance"]):
            x = 100 + i * 200
            y = 350 - val * 2.5
            svg_content += f"<circle cx='{x}' cy='{y}' r='4' fill='#2196F3'/>"
        
        # X轴标签
        for i, concur in enumerate(data["concurrency"]):
            x = 100 + i * 200
            svg_content += f"<text x='{x}' y='370' font-size='12' text-anchor='middle'>{concur}</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="50" width="20" height="20" fill="#4CAF50"/>
    <text x="630" y="65" font-size="12">主键搜索</text>
    <rect x="600" y="80" width="20" height="20" fill="#2196F3"/>
    <text x="630" y="95" font-size="12">插入性能</text>
</svg>
"""
    
    # 保存SVG文件
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(svg_content)

# 主函数
def main():
    """主函数"""
    print("生成性能测试图表...")
    
    # 创建输出目录
    output_dir = "docs/performance"
    os.makedirs(output_dir, exist_ok=True)
    
    # 生成数据库比较图表
    generate_svg_chart("database_comparison", performance_data["database_comparison"], f"{output_dir}/database_comparison.svg")
    print("生成数据库比较图表完成")
    
    # 生成ACID比较图表
    generate_svg_chart("acid_comparison", performance_data["acid_vs_non_acid"], f"{output_dir}/acid_comparison.svg")
    print("生成ACID比较图表完成")
    
    # 生成数据量影响图表
    generate_svg_chart("data_volume_impact", performance_data["data_volume_impact"], f"{output_dir}/data_volume_impact.svg")
    print("生成数据量影响图表完成")
    
    # 生成并发影响图表
    generate_svg_chart("concurrency_impact", performance_data["concurrency_impact"], f"{output_dir}/concurrency_impact.svg")
    print("生成并发影响图表完成")
    
    print("图表生成完成！")

if __name__ == "__main__":
    main()
