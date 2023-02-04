package session

import (
	"peeorm/clause"
	"reflect"
)

func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		table := s.Moedl(value).RefTable()                        // 新建一个表并返回（如果是同一个表不新建）
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames) // 生成Insert子句
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.VALUES, recordValues...) // 生成value语句
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (s *Session) Find(values interface{}) error {
	// 获取values（切片）的指针（反射）
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	// 获取元素类型（切片的类型的结构体实例，比如传进来destSlice.Type是[]User，那destSlice.Type().Elem()就是User）
	destType := destSlice.Type().Elem()
	table := s.Moedl(reflect.New(destType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	// 用上面构造出来的sql语句修改之后查询所有字段，值赋给rows
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	for rows.Next() {
		// 获取destType的实例，相当于传进来的切片类型的结构体实例
		dest := reflect.New(destType).Elem()
		var values []interface{}
		// 遍历所有字段
		for _, name := range table.FieldNames {
			// 获取dest的struct结构并且获取它的地址，转换为接口，相当于获取传进来的切片的结构体
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		// 将rows该行记录每一列的值依次赋值给 values 中的每一个字段，因为是接口地址，所以会直接赋值给dest
		if err := rows.Scan(values...); err != nil {
			return err
		}
		// 因为destSlice是外面的切片，dest是存储在切片里的结构体，用反射获取到地址直接加进去就OK
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}
