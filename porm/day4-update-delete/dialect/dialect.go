package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

type Dialect interface {
	// 将 Go 语言的类型转换为该数据库的数据类型。
	DataTypeOf(typ reflect.Value) string
	// 从sqlite_master里面查询信息（sql语句啥的）
	TableExistSQL(tableName string) (string, []interface{})
}

// 注册Dialect
func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

// 获取Dialect
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
