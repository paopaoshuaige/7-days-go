package session

import (
	"database/sql"
	"peeorm/clause"
	"peeorm/dialect"
	"peeorm/log"
	"peeorm/schema"
	"strings"
)

type Session struct {
	tx       *sql.Tx // 操作事务的指针
	db       *sql.DB // 连接数据库之后操作数据库的指针
	dialect  dialect.Dialect
	refTable *schema.Schema
	clause   clause.Clause
	sql      strings.Builder // sql语句
	sqlVars  []interface{}   // 占位符对应值
}

// CommonDB是db的最小函数集
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error) // 单个查询
	QueryRow(query string, args ...interface{}) *sql.Row        // 查询一行
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// 判断CommonDB是DB还是Tx
var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// 新建一个数据库链接
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

// 清空sql语句
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

// 获取db链接
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// 修改sql语句和占位符的值
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// 打印sql语句
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// 查询单个字段
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// 查询多个字段
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
