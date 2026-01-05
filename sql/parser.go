package sql

import (
	"strings"
)

// Parser SQL解析器
type Parser struct{}

// NewParser 创建一个新的SQL解析器
func NewParser() *Parser {
	return &Parser{}
}

// Parse 解析SQL语句
func (p *Parser) Parse(sql string) (Statement, error) {
	// 去除SQL语句中的注释和空格
	sql = strings.TrimSpace(sql)
	if len(sql) == 0 {
		return nil, nil
	}

	// 转换为大写，方便匹配关键字
	sqlUpper := strings.ToUpper(sql)

	// 根据SQL语句的类型创建不同的Statement
	if strings.HasPrefix(sqlUpper, "SELECT") {
		stmt := &SelectStatement{}
		stmt.BaseStatement = BaseStatement{
			rawSQL:  sql,
			sqlType: "SELECT",
		}
		return stmt, nil
	} else if strings.HasPrefix(sqlUpper, "INSERT") {
		stmt := &InsertStatement{}
		stmt.BaseStatement = BaseStatement{
			rawSQL:  sql,
			sqlType: "INSERT",
		}
		return stmt, nil
	} else if strings.HasPrefix(sqlUpper, "UPDATE") {
		stmt := &UpdateStatement{}
		stmt.BaseStatement = BaseStatement{
			rawSQL:  sql,
			sqlType: "UPDATE",
		}
		return stmt, nil
	} else if strings.HasPrefix(sqlUpper, "DELETE") {
		stmt := &DeleteStatement{}
		stmt.BaseStatement = BaseStatement{
			rawSQL:  sql,
			sqlType: "DELETE",
		}
		return stmt, nil
	} else if strings.HasPrefix(sqlUpper, "CREATE TABLE") {
		stmt := &CreateTableStatement{}
		stmt.BaseStatement = BaseStatement{
			rawSQL:  sql,
			sqlType: "CREATE TABLE",
		}
		return stmt, nil
	} else if strings.HasPrefix(sqlUpper, "CREATE UNIQUE INDEX") || strings.HasPrefix(sqlUpper, "CREATE INDEX") {
		stmt := &CreateIndexStatement{}
		stmt.BaseStatement = BaseStatement{
			rawSQL:  sql,
			sqlType: "CREATE INDEX",
		}
		return stmt, nil
	} else if strings.HasPrefix(sqlUpper, "CREATE DATABASE") {
		stmt := &CreateDatabaseStatement{}
		stmt.BaseStatement = BaseStatement{
			rawSQL:  sql,
			sqlType: "CREATE DATABASE",
		}
		return stmt, nil
	}

	// 未知类型的SQL语句
	stmt := &UnknownStatement{}
	stmt.BaseStatement = BaseStatement{
		rawSQL:  sql,
		sqlType: "UNKNOWN",
	}
	return stmt, nil
}
