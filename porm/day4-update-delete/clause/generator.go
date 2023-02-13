package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _UPDATE
	generators[DELETE] = _DELETE
	generators[COUNT] = _COUNT
}

// 根据表名删除
func _DELETE(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

func _UPDATE(values ...interface{}) (string, []interface{}) {
	// 获取表名
	tableName := values[0]
	// 获取所有待更新的参数，转换成map类型
	m := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	// 获取所有的key和value
	for k, v := range m {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(keys, ",")), vars
}

// 找到的项目数，*代表所有
func _COUNT(values ...interface{}) (string, []interface{}) {
	return _select(values[0], []string{"count(*)"})
}

// 转换成问号
func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

// 根据传入的值返回insert语句
func _insert(values ...interface{}) (string, []interface{}) {
	// INSERT INTO $tableName ($fields)
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}

// 返回VALUES语句
func _values(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1), ($v2), ...
	var bindStr string
	var sql strings.Builder
	var vars []interface{}
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{}) // 将value转成了切片,切片长度只有1
		// 转换成对应v切片数量的问号
		bindStr = genBindVars(len(v))
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i+1 != len(values) { // 不结束就分割一下
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars
}

// 返回select语句
func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

func _limit(values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	return "LIMIT ?", values
}

func _where(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}
