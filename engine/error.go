package engine

import "errors"

var (
	//没有主键
	ErrNoPrimaryKey = errors.New("no primary key")
	//没有记录
	ErrNoUpdateFields = errors.New("no update fields")
)
