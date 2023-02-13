package clause

import (
	"strings"
)

// sql语句和值
type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

// 根据对应的子句生成对应的sql语句
func (c *Clause) Set(name Type, vars ...interface{}) {
	// 如果当前调用者的sql语句是空的
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	// 就调用generators定义的函数获取sql和values
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

// 根据传入的type顺序构造sql语句。
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	// 返回拼接好的sql语句和值
	return strings.Join(sqls, " "), vars
}
