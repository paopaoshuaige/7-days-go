package session

import (
	"errors"
	"peeorm/clause"
	"reflect"
)

func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		s.CallMethod(BeforeInsert, value)
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
	s.CallMethod(AfterInsert, nil)

	return result.RowsAffected()
}

func (s *Session) Find(values interface{}) error {
	s.CallMethod(BeforeQuery, nil)
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
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		// 因为destSlice是外面的切片，dest是存储在切片里的结构体，用反射获取到地址直接加进去就OK
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

// 传入所有待更新的k-v
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	// 通过断言判断是否为map[s]interface类型，不是就转换赋值给m一个空的，否则就赋值给已存在的
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		// 遍历所有kv，映射
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	// 构造更新当前待更新的k-v的子句
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	// 构造sql语句
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	// 执行输出sql语句
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterUpdate, nil)
	// 返回受影响的行数
	return result.RowsAffected()
}

func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	// 构造deletesql语句
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	// 执行 输出
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	// 构造子句，SELECT count(*) FROM User
	s.clause.Set(clause.COUNT, s.refTable.Name)
	// 根据Count和where构造sql语句
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	// 根据这个sql语句查询对应的列，返回结果集有几行
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}

// 将限制条件加入到子句
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	// 对应Sql语句：LIMIT ?
	return s
}

// 添加where限制条件
func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	// 往vars里面加入传进来的参数，生成where子句
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

// 添加orderby条件
func (s *Session) Orderby(desc string) *Session {
	// 生成Orderby子句
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	// 通过反射新建一个dest类型的Slice实例
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	// 查找所有记录里面的第一行(使用反射获取地址直接赋值给destSlice)
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	// 如果没有找到
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	// 把destSlice第一条数据保存（也就是下标为0的），dest是通过反射获取到的value地址，会保存在value里
	dest.Set(destSlice.Index(0))
	return nil
}
