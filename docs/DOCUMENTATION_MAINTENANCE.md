# sfsDb 文档维护指南

## 文档结构

sfsDb 文档采用多语言、模块化的结构组织：

```
docs/
├── README.md              # 主文档，包含语言选择和目录
├── zh/                    # 中文版文档
│   ├── basic/             # 基础功能
│   ├── advanced/          # 高级功能
│   └── optimization/      # 优化与最佳实践
├── en/                    # 英文版文档
│   ├── basic/             # 基础功能
│   ├── advanced/          # 高级功能
│   └── optimization/      # 优化与最佳实践
└── DOCUMENTATION_MAINTENANCE.md  # 本文档
```

## 文档维护最佳实践

### 1. 同步更新原则

- **同时更新**：修改功能时，应同时更新中英文文档
- **保持一致**：确保中英文文档的内容和结构保持一致
- **版本标记**：在文档顶部添加版本号，确保用户知道文档的更新状态

### 2. 翻译规范

- **术语统一**：使用统一的技术术语翻译
- **风格一致**：保持文档的语言风格一致
- **避免直译**：对于技术概念，应使用行业标准的翻译
- **保持简洁**：翻译应简洁明了，避免冗长的表达

### 3. 目录结构管理

- **镜像结构**：中英文文档应保持相同的目录结构
- **文件命名**：使用相同的文件名，确保用户可以轻松找到对应语言的文档
- **链接管理**：内部链接应使用相对路径，确保在不同语言版本中都能正常工作

### 4. 文档更新流程

1. **功能修改**：修改代码或添加新功能
2. **文档更新**：
   - 首先更新主语言文档（建议为中文）
   - 然后翻译并更新对应英文文档
3. **链接检查**：确保所有内部链接正常工作
4. **版本更新**：更新文档版本号
5. **测试验证**：确保文档示例代码可以正常运行

### 5. 工具推荐

- **Markdown 编辑器**：使用支持多语言的 Markdown 编辑器
- **版本控制**：利用 Git 进行文档版本管理
- **翻译辅助**：可以使用翻译工具辅助，但需要人工审核确保准确性
- **链接检查**：使用工具检查文档中的死链接

## 文档内容映射

| 功能模块 | 中文版路径 | 英文版路径 |
|---------|-----------|-----------|
| 数据库初始化 | zh/basic/initialization.md | en/basic/initialization.md |
| 创建表与设置字段 | zh/basic/table_creation.md | en/basic/table_creation.md |
| 插入数据 | zh/basic/data_insertion.md | en/basic/data_insertion.md |
| 查询数据 | zh/basic/data_query.md | en/basic/data_query.md |
| 删除记录 | zh/basic/data_deletion.md | en/basic/data_deletion.md |
| 主键管理 | zh/advanced/primary_key.md | en/advanced/primary_key.md |
| 索引管理 | zh/advanced/index_management.md | en/advanced/index_management.md |
| 全文搜索 | zh/advanced/full_text_search.md | en/advanced/full_text_search.md |
| 字段修改 | zh/advanced/field_modification.md | en/advanced/field_modification.md |
| 事务管理 | zh/advanced/transaction.md | en/advanced/transaction.md |
| 其他功能 | zh/advanced/other_features.md | en/advanced/other_features.md |
| 对象池与内存管理 | zh/optimization/object_pool.md | en/optimization/object_pool.md |
| 半结构化数据支持 | zh/optimization/semi_structured.md | en/optimization/semi_structured.md |
| 最佳实践 | zh/optimization/best_practices.md | en/optimization/best_practices.md |
| 常见问题 | zh/optimization/common_issues.md | en/optimization/common_issues.md |
| 总结 | zh/optimization/summary.md | en/optimization/summary.md |

## 文档版本管理

建议使用以下版本号格式：

```
{主版本号}.{次版本号}.{修订号}
```

- **主版本号**：重大功能变更或架构调整
- **次版本号**：新功能添加
- **修订号**：错误修复或文档更新

每次更新文档时，应更新版本号并在文档顶部添加更新日志。

## 结语

良好的文档维护对于开源项目的成功至关重要。通过遵循本指南，您可以确保 sfsDb 的文档始终保持最新、准确和一致，为用户提供更好的使用体验。