package session

import (
	"database/sql"
	"day1_database/log"
	"strings"
)

type Session struct {
	db      *sql.DB         // 连接数据库之后操作数据库的指针
	sql     strings.Builder // sql语句
	sqlVars []interface{}   // 占位符对应值
}

// 新建一个数据库链接
func New(db *sql.DB) *Session {
	return &Session{db: db}
}

// 清空sql语句
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

// 获取db链接
func (s *Session) DB() *sql.DB {
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
