package session

import (
	"day2-reflect-schema/log"
	"day2-reflect-schema/schema"
	"fmt"
	"reflect"
	"strings"
)

// 用于给 refTable 赋值
func (s *Session) Moedl(value interface{}) *Session {
	// 如果没有值，或者是对象的类型和表的对象类型不相同
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		// 就为这个新的对象解析然后赋值
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

// 获取数据库表
func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Moedl is not set")
	}
	return s.refTable
}

// 创建table
func (s *Session) CreateTable() error {
	// 获取表信息
	table := s.RefTable()
	var colums []string
	for _, field := range table.Fields {
		// 遍历所有字段，往字符串切片里面存储字段名，类型，tag
		colums = append(colums, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	desc := strings.Join(colums, ",")
	// 创建表，输出sql语句
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

func (s *Session) DropTable() error {
	// 删除表
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	// 执行查询sql语句
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	// 返回是否存在当前表
	return tmp == s.RefTable().Name
}
