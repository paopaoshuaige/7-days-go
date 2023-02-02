package schema

import (
	"day2-reflect-schema/dialect"
	"go/ast"
	"reflect"
)

// field表示数据库的一列
type Field struct {
	Name string // 字段名
	Type string // 类型
	Tag  string // 约束
}

// schema表示数据库的表
type Schema struct {
	Model      interface{}       // 被映射的对象
	Name       string            // 表名
	Fields     []*Field          // 字段
	FieldNames []string          // 所有的字段名
	fieldMap   map[string]*Field // 字段名和Field的映射关系，直接调用不用遍历Fields
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// values返回dest的成员变量的值
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	// 遍历当前实例对象的所有字段
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}

type ITableName interface {
	TableName() string
}

// 将任意的对象解析为 Schema 实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// indirect返回指针指向的实例，valueof返回值，type是返回类型
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	var tableName string
	t, ok := dest.(ITableName)
	if !ok {
		tableName = modelType.Name()
	} else {
		tableName = t.TableName()
	}
	schema := &Schema{
		Model:    dest,
		Name:     tableName,
		fieldMap: make(map[string]*Field),
	}

	// 获取实例的字段的个数，通过下标取到特定字段
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		// 字段名没问题
		if !p.Anonymous && ast.IsExported(p.Name) {
			// 转换数据类型赋值
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			// 查找与标签字符串中的键相关联的值，没有返回nil
			if v, ok := p.Tag.Lookup("peeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
