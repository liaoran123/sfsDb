package sql

import "strings"

// Visitor 用于遍历SQL AST并提取信息
type Visitor struct{}

// NewVisitor 创建一个新的AST访问器
func NewVisitor() *Visitor {
	return &Visitor{}
}

// ExtractTableNames 从Statement中提取表名
func (v *Visitor) ExtractTableNames(stmt Statement) ([]string, error) {
	var tables []string

	// 简化实现：基于SQL字符串解析提取表名
	sql := stmt.Raw()
	sql = strings.TrimSpace(sql)
	if len(sql) == 0 {
		return tables, nil
	}

	// 提取表名的简化实现
	sqlUpper := strings.ToUpper(sql)

	switch stmt.Type() {
	case "SELECT":
		// 处理SELECT语句的表名提取
		fromIndex := strings.Index(sqlUpper, " FROM ")
		if fromIndex == -1 {
			return tables, nil
		}

		// 提取FROM子句到WHERE/ORDER BY等关键字之间的部分
		whereIndex := strings.Index(sqlUpper, " WHERE ")
		orderIndex := strings.Index(sqlUpper, " ORDER BY ")
		groupIndex := strings.Index(sqlUpper, " GROUP BY ")
		havingIndex := strings.Index(sqlUpper, " HAVING ")
		limitIndex := strings.Index(sqlUpper, " LIMIT ")

		// 确定表名部分的结束位置
		endIndex := len(sql)
		if whereIndex != -1 && whereIndex > fromIndex {
			endIndex = whereIndex
		}
		if orderIndex != -1 && orderIndex > fromIndex {
			if orderIndex < endIndex {
				endIndex = orderIndex
			}
		}
		if groupIndex != -1 && groupIndex > fromIndex {
			if groupIndex < endIndex {
				endIndex = groupIndex
			}
		}
		if havingIndex != -1 && havingIndex > fromIndex {
			if havingIndex < endIndex {
				endIndex = havingIndex
			}
		}
		if limitIndex != -1 && limitIndex > fromIndex {
			if limitIndex < endIndex {
				endIndex = limitIndex
			}
		}

		// 提取表名部分
		tablePart := sql[fromIndex+6 : endIndex]
		tablePart = strings.TrimSpace(tablePart)

		// 处理JOIN语句，提取所有表名
		// 替换所有JOIN关键字为统一的分隔符
		tablePartUpper := strings.ToUpper(tablePart)
		tablePartUpper = strings.ReplaceAll(tablePartUpper, " INNER JOIN ", " JOIN ")
		tablePartUpper = strings.ReplaceAll(tablePartUpper, " LEFT JOIN ", " JOIN ")
		tablePartUpper = strings.ReplaceAll(tablePartUpper, " RIGHT JOIN ", " JOIN ")
		tablePartUpper = strings.ReplaceAll(tablePartUpper, " FULL JOIN ", " JOIN ")
		tablePartUpper = strings.ReplaceAll(tablePartUpper, " CROSS JOIN ", " JOIN ")

		// 按JOIN分割表名
		joinParts := strings.Split(tablePartUpper, " JOIN ")
		for _, joinPart := range joinParts {
			// 找到对应的原始表名部分
			// 查找ON关键字，确定表名结束位置
			onIndex := strings.Index(joinPart, " ON ")
			var tableName string
			if onIndex != -1 {
				tableName = strings.TrimSpace(joinPart[:onIndex])
			} else {
				tableName = strings.TrimSpace(joinPart)
			}

			// 提取表名，忽略别名
			spaceIndex := strings.Index(tableName, " ")
			if spaceIndex != -1 {
				tableName = tableName[:spaceIndex]
			}
			tableName = strings.TrimSpace(tableName)

			if tableName != "" {
				tables = append(tables, tableName)
			}
		}

	case "DELETE":
		// 查找FROM关键字后的表名
		fromIndex := strings.Index(sqlUpper, " FROM ")
		if fromIndex == -1 {
			return tables, nil
		}

		// 提取原始SQL中的表名部分
		fromPart := sql[fromIndex+6:]
		fromPartUpper := sqlUpper[fromIndex+6:]

		// 查找WHERE或ORDER BY等后续关键字
		whereIndex := strings.Index(fromPartUpper, " WHERE ")

		// 确定表名部分的结束位置
		endIndex := len(fromPart)
		if whereIndex != -1 && whereIndex < endIndex {
			endIndex = whereIndex
		}

		// 提取表名（保持原始大小写）
		tableName := strings.TrimSpace(fromPart[:endIndex])
		if tableName != "" {
			tables = append(tables, tableName)
		}

	case "INSERT":
		// 处理INSERT语句的表名提取
		// 查找INSERT INTO关键字，前面可以有0个或多个空格
		intoIndex := strings.Index(sqlUpper, "INSERT INTO ")
		if intoIndex == -1 {
			return tables, nil
		}

		// 提取INSERT INTO后的表名
		intoPart := sql[intoIndex+12:]
		intoPartUpper := sqlUpper[intoIndex+12:]

		// 查找左括号或空格，确定表名结束位置
		leftParenIndex := strings.Index(intoPart, " (")
		spaceIndex := strings.Index(intoPartUpper, " ")

		// 确定表名部分的结束位置
		endIndex := len(intoPart)
		if leftParenIndex != -1 && leftParenIndex < endIndex {
			endIndex = leftParenIndex
		}
		if spaceIndex != -1 && spaceIndex < endIndex {
			endIndex = spaceIndex
		}

		// 提取表名（保持原始大小写）
		tableName := strings.TrimSpace(intoPart[:endIndex])
		if tableName != "" {
			tables = append(tables, tableName)
		}

	case "UPDATE":
		// 处理UPDATE语句的表名提取
		updateIndex := strings.Index(sqlUpper, "UPDATE ")
		if updateIndex == -1 {
			return tables, nil
		}

		// 提取UPDATE后的表名
		updatePart := sql[updateIndex+7:]
		updatePartUpper := sqlUpper[updateIndex+7:]

		// 查找SET关键字，确定表名结束位置
		setIndex := strings.Index(updatePartUpper, " SET ")

		// 确定表名部分的结束位置
		endIndex := len(updatePart)
		if setIndex != -1 && setIndex < endIndex {
			endIndex = setIndex
		}

		// 提取表名（保持原始大小写）
		tableName := strings.TrimSpace(updatePart[:endIndex])
		if tableName != "" {
			tables = append(tables, tableName)
		}

	case "CREATE TABLE":
		// 处理CREATE TABLE语句的表名提取
		createIndex := strings.Index(sqlUpper, "CREATE TABLE ")
		if createIndex == -1 {
			return tables, nil
		}

		// 提取CREATE TABLE后的表名
		tablePart := sql[createIndex+13:]

		// 查找左括号，确定表名结束位置
		leftParenIndex := strings.Index(tablePart, " (")
		if leftParenIndex == -1 {
			return tables, nil
		}

		// 提取表名（保持原始大小写）
		tableName := strings.TrimSpace(tablePart[:leftParenIndex])
		if tableName != "" {
			tables = append(tables, tableName)
		}

	case "CREATE INDEX":
		// 处理CREATE INDEX语句的表名提取
		createIndex := strings.Index(sqlUpper, "CREATE INDEX ")
		if createIndex == -1 {
			return tables, nil
		}

		// 查找ON关键字，确定表名位置
		onIndex := strings.Index(sqlUpper[createIndex:], " ON ")
		if onIndex == -1 {
			return tables, nil
		}

		// 提取ON后的表名
		tablePart := sql[createIndex+onIndex+4:]
		tablePartUpper := sqlUpper[createIndex+onIndex+4:]

		// 查找左括号或空格，确定表名结束位置
		leftParenIndex := strings.Index(tablePart, " (")
		spaceIndex := strings.Index(tablePartUpper, " ")

		// 确定表名部分的结束位置
		endIndex := len(tablePart)
		if leftParenIndex != -1 && leftParenIndex < endIndex {
			endIndex = leftParenIndex
		}
		if spaceIndex != -1 && spaceIndex < endIndex {
			endIndex = spaceIndex
		}

		// 提取表名（保持原始大小写）
		tableName := strings.TrimSpace(tablePart[:endIndex])
		if tableName != "" {
			tables = append(tables, tableName)
		}

	default:
		return tables, nil
	}

	return tables, nil
}

// ExtractColumns 从Statement中提取列名
func (v *Visitor) ExtractColumns(stmt Statement) ([]string, error) {
	var columns []string

	// 简化实现：基于SQL字符串解析提取列名
	sql := stmt.Raw()
	sql = strings.TrimSpace(sql)
	if len(sql) == 0 {
		return columns, nil
	}

	// 根据SQL语句类型提取列名
	switch stmt.Type() {
	case "SELECT":
		// 查找SELECT和FROM之间的列名部分
		sqlUpper := strings.ToUpper(sql)
		selectIndex := strings.Index(sqlUpper, "SELECT ")
		fromIndex := strings.Index(sqlUpper, " FROM ")
		if selectIndex == -1 || fromIndex == -1 || fromIndex < selectIndex {
			return columns, nil
		}

		// 提取列名部分，处理换行符
		colsPart := sql[selectIndex+7 : fromIndex]
		colsPart = strings.ReplaceAll(colsPart, "\n", " ")
		colsPart = strings.ReplaceAll(colsPart, "\r", " ")
		colsPart = strings.TrimSpace(colsPart)

		// 简单的列名解析：按逗号分割
		colNames := strings.Split(colsPart, ",")
		for _, col := range colNames {
			col = strings.TrimSpace(col)
			if col == "*" {
				continue // 跳过通配符
			}
			// 提取列名，忽略可能的别名
			aliasIndex := strings.Index(strings.ToUpper(col), " AS ")
			if aliasIndex == -1 {
				aliasIndex = strings.Index(col, " ")
			}
			if aliasIndex != -1 {
				col = col[:aliasIndex]
			}
			col = strings.TrimSpace(col)
			if col != "" {
				columns = append(columns, col)
			}
		}

	case "INSERT":
		// 查找INSERT INTO和VALUES之间的列名部分
		sqlUpper := strings.ToUpper(sql)
		intoIndex := strings.Index(sqlUpper, "INSERT INTO ")
		valuesIndex := strings.Index(sqlUpper, " VALUES ")
		if intoIndex == -1 || valuesIndex == -1 || valuesIndex < intoIndex {
			return columns, nil
		}

		// 提取列名部分
		colsPart := sql[intoIndex+12 : valuesIndex]
		colsPart = strings.TrimSpace(colsPart)

		// 检查是否包含括号
		if len(colsPart) > 2 && colsPart[0] == '(' && colsPart[len(colsPart)-1] == ')' {
			// 移除括号
			colsPart = colsPart[1 : len(colsPart)-1]
			colsPart = strings.TrimSpace(colsPart)

			// 简单的列名解析：按逗号分割
			colNames := strings.Split(colsPart, ",")
			for _, col := range colNames {
				col = strings.TrimSpace(col)
				if col != "" {
					columns = append(columns, col)
				}
			}
		}

	case "UPDATE":
		// 查找SET和WHERE之间的列名部分
		sqlUpper := strings.ToUpper(sql)
		setIndex := strings.Index(sqlUpper, " SET ")
		whereIndex := strings.Index(sqlUpper, " WHERE ")
		if setIndex == -1 {
			return columns, nil
		}

		// 确定SET子句的结束位置
		endIndex := len(sql)
		if whereIndex != -1 && whereIndex > setIndex {
			endIndex = whereIndex
		}

		// 提取SET子句部分
		setPart := sql[setIndex+4 : endIndex]
		setPart = strings.TrimSpace(setPart)

		// 简单的列名解析：按逗号分割SET子句
		setClauses := strings.Split(setPart, ",")
		for _, clause := range setClauses {
			clause = strings.TrimSpace(clause)
			// 查找等号，提取列名
			equalIndex := strings.Index(clause, "=")
			if equalIndex != -1 {
				col := clause[:equalIndex]
				col = strings.TrimSpace(col)
				if col != "" {
					columns = append(columns, col)
				}
			}
		}

	case "CREATE TABLE":
		// 查找CREATE TABLE和右括号之间的列定义部分
		sqlUpper := strings.ToUpper(sql)
		createIndex := strings.Index(sqlUpper, "CREATE TABLE ")
		if createIndex == -1 {
			return columns, nil
		}

		// 提取表名和列定义部分
		tablePart := sql[createIndex+13:]
		leftParenIndex := strings.Index(tablePart, " (")
		rightParenIndex := strings.LastIndex(tablePart, ")")
		if leftParenIndex == -1 || rightParenIndex == -1 || rightParenIndex < leftParenIndex {
			return columns, nil
		}

		// 提取列定义部分
		colsPart := tablePart[leftParenIndex+2 : rightParenIndex]
		colsPart = strings.TrimSpace(colsPart)

		// 简单的列名解析：按逗号分割
		colDefs := strings.Split(colsPart, ",")
		for _, colDef := range colDefs {
			colDef = strings.TrimSpace(colDef)
			// 跳过索引定义
			if strings.HasPrefix(strings.ToUpper(colDef), "PRIMARY KEY") ||
			   strings.HasPrefix(strings.ToUpper(colDef), "UNIQUE") ||
			   strings.HasPrefix(strings.ToUpper(colDef), "INDEX") {
				continue
			}

			// 提取列名（第一个空格前的部分）
			spaceIndex := strings.Index(colDef, " ")
			if spaceIndex != -1 {
				col := colDef[:spaceIndex]
				col = strings.TrimSpace(col)
				if col != "" {
					columns = append(columns, col)
				}
			}
		}
	}

	return columns, nil
}

// ColumnTableMapping 表示列名到表名的映射关系
type ColumnTableMapping struct {
	ColumnName string // 列名，如 "id", "u.id", "name"
	TableName  string // 表名或表别名，如 "users", "u"
	IsQualified bool   // 是否为限定列名（包含表名前缀）
}

// ExtractColumnTableMapping 从Statement中提取列名与表名的映射关系
func (v *Visitor) ExtractColumnTableMapping(stmt Statement) ([]ColumnTableMapping, error) {
	var mappings []ColumnTableMapping

	// 简化实现：基于SQL字符串解析提取列名和表名映射
	sql := stmt.Raw()
	// 移除换行符，简化解析
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\r", " ")
	sql = strings.TrimSpace(sql)
	if len(sql) == 0 {
		return mappings, nil
	}

	// 只处理SELECT语句
	if stmt.Type() != "SELECT" {
		return mappings, nil
	}

	// 查找SELECT和FROM之间的列名部分
	sqlUpper := strings.ToUpper(sql)
	selectIndex := strings.Index(sqlUpper, "SELECT ")
	fromIndex := strings.Index(sqlUpper, " FROM ")
	if selectIndex == -1 || fromIndex == -1 || fromIndex < selectIndex {
		return mappings, nil
	}

	// 提取列名部分
	colsPart := sql[selectIndex+7 : fromIndex]
	colsPart = strings.TrimSpace(colsPart)

	// 解析列名
	colNames := strings.Split(colsPart, ",")
	for _, col := range colNames {
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}
		if col == "*" {
			continue // 跳过通配符
		}

		// 提取列名，忽略可能的别名
		aliasIndex := strings.Index(strings.ToUpper(col), " AS ")
		if aliasIndex == -1 {
			aliasIndex = strings.Index(col, " ")
		}
		if aliasIndex != -1 {
			col = col[:aliasIndex]
		}
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}

		// 解析列名中的表名前缀
		var tableName string
		isQualified := false

		// 检查是否为限定列名（包含表名前缀）
		if dotIndex := strings.Index(col, "."); dotIndex != -1 {
			// 提取表名前缀和列名
			tableName = col[:dotIndex]
			isQualified = true
		}

		// 添加映射关系
		mappings = append(mappings, ColumnTableMapping{
			ColumnName: col,
			TableName:  tableName,
			IsQualified: isQualified,
		})
	}

	return mappings, nil
}

// ExtractTablesWithAliases 从Statement中提取表名及其别名
func (v *Visitor) ExtractTablesWithAliases(stmt Statement) ([]map[string]string, error) {
	var tables []map[string]string

	// 简化实现：基于SQL字符串解析提取表名及其别名
	sql := stmt.Raw()
	sql = strings.TrimSpace(sql)
	if len(sql) == 0 {
		return tables, nil
	}

	// 只处理SELECT语句
	if stmt.Type() != "SELECT" {
		return tables, nil
	}

	// 提取FROM子句及JOIN子句
	sqlUpper := strings.ToUpper(sql)

	// 查找FROM关键字
	fromIndex := strings.Index(sqlUpper, " FROM ")
	if fromIndex == -1 {
		return tables, nil
	}

	// 查找WHERE、ORDER BY、GROUP BY等关键字，确定表名部分的结束位置
	var endIndex int = len(sql)
	if whereIndex := strings.Index(sqlUpper, " WHERE "); whereIndex != -1 && whereIndex > fromIndex {
		endIndex = whereIndex
	}
	if orderIndex := strings.Index(sqlUpper, " ORDER BY "); orderIndex != -1 && orderIndex > fromIndex {
		if orderIndex < endIndex {
			endIndex = orderIndex
		}
	}
	if groupIndex := strings.Index(sqlUpper, " GROUP BY "); groupIndex != -1 && groupIndex > fromIndex {
		if groupIndex < endIndex {
			endIndex = groupIndex
		}
	}
	if havingIndex := strings.Index(sqlUpper, " HAVING "); havingIndex != -1 && havingIndex > fromIndex {
		if havingIndex < endIndex {
			endIndex = havingIndex
		}
	}
	if limitIndex := strings.Index(sqlUpper, " LIMIT "); limitIndex != -1 && limitIndex > fromIndex {
		if limitIndex < endIndex {
			endIndex = limitIndex
		}
	}

	// 提取表名部分（FROM和JOIN子句）
	tablePart := sql[fromIndex+6 : endIndex]
	tablePart = strings.TrimSpace(tablePart)

	// 替换所有换行符为空格，简化解析
	tablePart = strings.ReplaceAll(tablePart, "\n", " ")
	tablePart = strings.ReplaceAll(tablePart, "\r", " ")

	// 分割表名部分，处理JOIN子句
	// 替换所有JOIN关键字为统一的分隔符
	tablePartUpper := strings.ToUpper(tablePart)
	// 替换各种JOIN类型
	tablePartUpper = strings.ReplaceAll(tablePartUpper, " INNER JOIN ", " JOIN ")
	tablePartUpper = strings.ReplaceAll(tablePartUpper, " LEFT JOIN ", " JOIN ")
	tablePartUpper = strings.ReplaceAll(tablePartUpper, " RIGHT JOIN ", " JOIN ")
	tablePartUpper = strings.ReplaceAll(tablePartUpper, " FULL JOIN ", " JOIN ")
	tablePartUpper = strings.ReplaceAll(tablePartUpper, " CROSS JOIN ", " JOIN ")

	// 按JOIN分割表名
	joinParts := strings.Split(tablePartUpper, " JOIN ")
	for _, joinPart := range joinParts {
		joinPart = strings.TrimSpace(joinPart)
		if joinPart == "" {
			continue
		}

		// 查找ON关键字，确定表名结束位置
		onIndex := strings.Index(joinPart, " ON ")
		var tableClause string
		if onIndex != -1 {
			tableClause = strings.TrimSpace(joinPart[:onIndex])
		} else {
			tableClause = joinPart
		}

		// 提取表名和别名
		spaceIndex := strings.Index(tableClause, " ")
		if spaceIndex != -1 {
			// 提取表名和别名
			tableName := tableClause[:spaceIndex]
			alias := tableClause[spaceIndex+1:]
			alias = strings.TrimSpace(alias)
			tables = append(tables, map[string]string{
				"table": tableName,
				"alias": alias,
			})
		} else {
			// 没有别名
			tables = append(tables, map[string]string{
				"table": tableClause,
				"alias": "",
			})
		}
	}

	return tables, nil
}

// ExtractIndexInfo 从CREATE INDEX语句中提取索引信息
func (v *Visitor) ExtractIndexInfo(stmt Statement) (map[string]interface{}, error) {
	indexInfo := make(map[string]interface{})

	// 只处理CREATE INDEX语句
	if stmt.Type() != "CREATE INDEX" {
		return indexInfo, nil
	}

	// 简化实现：基于SQL字符串解析提取索引信息
	sql := stmt.Raw()
	sql = strings.TrimSpace(sql)
	if len(sql) == 0 {
		return indexInfo, nil
	}

	sqlUpper := strings.ToUpper(sql)

	// 提取索引名，处理CREATE UNIQUE INDEX和CREATE INDEX两种情况
	var indexNameStart int
	if strings.HasPrefix(sqlUpper, "CREATE UNIQUE INDEX ") {
		indexNameStart = 20 // "CREATE UNIQUE INDEX "的长度
	} else if strings.HasPrefix(sqlUpper, "CREATE INDEX ") {
		indexNameStart = 13 // "CREATE INDEX "的长度
	} else {
		return indexInfo, nil
	}

	onIndex := strings.Index(sqlUpper, " ON ")
	if onIndex == -1 || onIndex <= indexNameStart {
		return indexInfo, nil
	}

	indexName := strings.TrimSpace(sql[indexNameStart:onIndex])
	indexInfo["name"] = indexName

	// 提取表名
	tablePart := sql[onIndex+4:]
	tablePartUpper := sqlUpper[onIndex+4:]
	leftParenIndex := strings.Index(tablePart, " (")
	spaceIndex := strings.Index(tablePartUpper, " ")

	endIndex := len(tablePart)
	if leftParenIndex != -1 && leftParenIndex < endIndex {
		endIndex = leftParenIndex
	}
	if spaceIndex != -1 && spaceIndex < endIndex {
		endIndex = spaceIndex
	}

	tableName := strings.TrimSpace(tablePart[:endIndex])
	indexInfo["table"] = tableName

	// 提取索引字段
	if leftParenIndex != -1 {
		rightParenIndex := strings.Index(tablePart, ")")
		if rightParenIndex != -1 && rightParenIndex > leftParenIndex {
			fieldsPart := tablePart[leftParenIndex+2 : rightParenIndex]
			fieldsPart = strings.TrimSpace(fieldsPart)
			fields := strings.Split(fieldsPart, ",")
			for i, field := range fields {
				fields[i] = strings.TrimSpace(field)
			}
			indexInfo["fields"] = fields
		}
	}

	// 检查是否为唯一索引
	if strings.Contains(sqlUpper, "UNIQUE") {
		indexInfo["unique"] = true
	} else {
		indexInfo["unique"] = false
	}

	return indexInfo, nil
}

// FormatSQL 格式化SQL语句
func FormatSQL(sql string) (string, error) {
	// 简化实现，直接返回原始SQL
	return sql, nil
}
