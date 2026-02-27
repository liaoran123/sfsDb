#!/usr/bin/env python3
"""
事务性能测试图表生成脚本
使用内置库生成SVG图表
"""

import os

# 事务性能数据
transaction_data = {
    "database_comparison": {
        "dbs": ["sfsDb", "SQLite", "BoltDB", "BadgerDB"],
        "single_transaction": [1.2, 2.5, 1.8, 1.5],  # 单事务性能 (ms)
        "concurrent_performance": [26315, 8500, 15200, 18700]  # 并发性能 (ops/s)
    },
    "transaction_operations": {
        "operations": ["事务创建", "事务提交", "单条写入", "单条读取", "读写组合", "事务回滚"],
        "performance": [0.5, 1.2, 0.8, 0.3, 1.0, 0.6]  # 平均耗时 (ms)
    },
    "batch_operations": {
        "batch_sizes": ["10条", "100条"],
        "performance": [5.2, 45.3]  # 平均耗时 (ms)
    },
    "concurrency_performance": {
        "concurrency": ["10", "50"],
        "performance": [26315, 26315]  # 每秒操作数 (ops/s)
    }
}

# 生成SVG图表
def generate_svg_chart(chart_type, data, output_path):
    """生成SVG格式的图表"""
    if chart_type == "database_comparison":
        # 生成数据库比较图表
        svg_content = f"""
<svg width="800" height="500" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="500" fill="#f9f9f9"/>
    <text x="400" y="40" font-size="20" text-anchor="middle" font-weight="bold">sfsDb 与其他数据库事务性能对比</text>
"""
        
        # 绘制单事务性能柱状图
        bar_width = 60
        gap = 40
        start_x = 100
        
        for i, db in enumerate(data["dbs"]):
            x = start_x + i * (bar_width * 2 + gap)
            
            # 单事务性能
            height = data["single_transaction"][i] * 30
            svg_content += f"<rect x='{x}' y='{150 - height}' width='{bar_width}' height='{height}' fill='#4CAF50'/>"
            svg_content += f"<text x='{x + bar_width/2}' y='{170}' font-size='12' text-anchor='middle'>{data['single_transaction'][i]}ms</text>"
            
            # 并发性能
            height = data["concurrent_performance"][i] / 1000 * 2
            svg_content += f"<rect x='{x + bar_width + 10}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#2196F3'/>"
            svg_content += f"<text x='{x + bar_width*1.5 + 10}' y='{370}' font-size='12' text-anchor='middle'>{data['concurrent_performance'][i]} ops/s</text>"
            
            # 数据库名称
            svg_content += f"<text x='{x + bar_width + 5}' y='{400}' font-size='12' text-anchor='middle'>{db}</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="80" width="20" height="20" fill="#4CAF50"/>
    <text x="630" y="95" font-size="12">单事务性能 (ms)</text>
    <rect x="600" y="110" width="20" height="20" fill="#2196F3"/>
    <text x="630" y="125" font-size="12">并发性能 (ops/s)</text>
    <text x="400" y="130" font-size="14" text-anchor="middle" font-weight="bold">单事务性能</text>
    <text x="400" y="330" font-size="14" text-anchor="middle" font-weight="bold">并发性能 (10线程)</text>
</svg>
"""
    
    elif chart_type == "transaction_operations":
        # 生成事务操作性能图表
        svg_content = f"""
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="400" fill="#f9f9f9"/>
    <text x="400" y="30" font-size="20" text-anchor="middle" font-weight="bold">事务操作性能</text>
    <text x="400" y="380" font-size="12" text-anchor="middle">平均耗时 (ms)</text>
"""
        
        # 绘制柱状图
        bar_width = 60
        gap = 20
        start_x = 100
        
        for i, op in enumerate(data["operations"]):
            x = start_x + i * (bar_width + gap)
            
            # 性能
            height = data["performance"][i] * 50
            svg_content += f"<rect x='{x}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#4CAF50'/>"
            svg_content += f"<text x='{x + bar_width/2}' y='{370}' font-size='12' text-anchor='middle'>{op}</text>"
            svg_content += f"<text x='{x + bar_width/2}' y='{350 - height - 5}' font-size='12' text-anchor='middle'>{data['performance'][i]}ms</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="50" width="20" height="20" fill="#4CAF50"/>
    <text x="630" y="65" font-size="12">平均耗时 (ms)</text>
</svg>
"""
    
    elif chart_type == "batch_operations":
        # 生成批量操作性能图表
        svg_content = f"""
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="400" fill="#f9f9f9"/>
    <text x="400" y="30" font-size="20" text-anchor="middle" font-weight="bold">批量操作性能</text>
    <text x="400" y="380" font-size="12" text-anchor="middle">平均耗时 (ms)</text>
"""
        
        # 绘制柱状图
        bar_width = 100
        gap = 100
        start_x = 250
        
        for i, size in enumerate(data["batch_sizes"]):
            x = start_x + i * (bar_width + gap)
            
            # 性能
            height = data["performance"][i] * 3
            svg_content += f"<rect x='{x}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#2196F3'/>"
            svg_content += f"<text x='{x + bar_width/2}' y='{370}' font-size='12' text-anchor='middle'>{size}</text>"
            svg_content += f"<text x='{x + bar_width/2}' y='{350 - height - 5}' font-size='12' text-anchor='middle'>{data['performance'][i]}ms</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="50" width="20" height="20" fill="#2196F3"/>
    <text x="630" y="65" font-size="12">平均耗时 (ms)</text>
</svg>
"""
    
    elif chart_type == "concurrency_performance":
        # 生成并发性能图表
        svg_content = f"""
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
    <rect width="800" height="400" fill="#f9f9f9"/>
    <text x="400" y="30" font-size="20" text-anchor="middle" font-weight="bold">并发性能</text>
    <text x="400" y="380" font-size="12" text-anchor="middle">每秒操作数 (ops/s)</text>
"""
        
        # 绘制柱状图
        bar_width = 100
        gap = 100
        start_x = 250
        
        for i, concur in enumerate(data["concurrency"]):
            x = start_x + i * (bar_width + gap)
            
            # 性能
            height = data["performance"][i] / 1000 * 1.5
            svg_content += f"<rect x='{x}' y='{350 - height}' width='{bar_width}' height='{height}' fill='#FF9800'/>"
            svg_content += f"<text x='{x + bar_width/2}' y='{370}' font-size='12' text-anchor='middle'>{concur} 并发</text>"
            svg_content += f"<text x='{x + bar_width/2}' y='{350 - height - 5}' font-size='12' text-anchor='middle'>{data['performance'][i]} ops/s</text>"
        
        # 图例
        svg_content += f"""
    <rect x="600" y="50" width="20" height="20" fill="#FF9800"/>
    <text x="630" y="65" font-size="12">每秒操作数 (ops/s)</text>
</svg>
"""
    
    # 保存SVG文件
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(svg_content)

# 主函数
def main():
    """主函数"""
    print("生成事务性能测试图表...")
    
    # 创建输出目录
    output_dir = "docs/performance"
    os.makedirs(output_dir, exist_ok=True)
    
    # 生成数据库比较图表
    generate_svg_chart("database_comparison", transaction_data["database_comparison"], f"{output_dir}/transaction_database_comparison.svg")
    print("生成数据库比较图表完成")
    
    # 生成事务操作性能图表
    generate_svg_chart("transaction_operations", transaction_data["transaction_operations"], f"{output_dir}/transaction_operations.svg")
    print("生成事务操作性能图表完成")
    
    # 生成批量操作性能图表
    generate_svg_chart("batch_operations", transaction_data["batch_operations"], f"{output_dir}/transaction_batch_operations.svg")
    print("生成批量操作性能图表完成")
    
    # 生成并发性能图表
    generate_svg_chart("concurrency_performance", transaction_data["concurrency_performance"], f"{output_dir}/transaction_concurrency_performance.svg")
    print("生成并发性能图表完成")
    
    print("图表生成完成！")

if __name__ == "__main__":
    main()
