package main

import (
	"database/sql"
	"fmt"
	"peeorm/dialect"
	"peeorm/log"
	"peeorm/session"
	"strings"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

type TxFunc func(*session.Session) (interface{}, error)

// 给用户提供的接口，传进来一个函数，里面有数据库的操作
func (engine *Engine) Transaction(f TxFunc) (reslute interface{}, err error) {
	// 获取数据库指针
	s := engine.NewSession()
	// 启动事务
	if err := s.Begin(); err != nil {
		return nil, err
	}
	// 通过defer和recover的组合达成事务操作
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p)
		} else if err != nil {
			_ = s.Rollback()
		} else {
			err = s.Commit()
		}
	}()

	return f(s)
}

// 用来计算前后两个字段切片的差集。新表 - 旧表 = 新增字段，旧表 - 新表 = 删除字段。
func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

func (engine *Engine) Migrate(value interface{}) error {
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		if !s.Moedl(value).HasTable() { // 如果当前表不存在
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}

		table := s.RefTable() // 获取表
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		// 获取字段名
		columns, _ := rows.Columns()
		// 新增字段和删除字段
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		// 新增了哪几列，删除了哪几列
		log.Infof("added cols %v, deleted cols %v", addCols, delCols)

		// 遍历所有新增的列
		for _, col := range addCols {
			// 获取字段
			f := table.GetField(col)
			// 加入新的列
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		if len(delCols) == 0 { // 如果没有删除就出去
			return
		}

		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		// 删除旧表，给新的重命名
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}

// 创建引擎实例并且连接数据库
func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// 发送ping以确保数据库连接处于活动状态。
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	// 确保特定的dialect存在
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return
	}
	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success")
	return
}

// 关闭数据库链接
func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
}

// 连接数据库，返回db指针，调用ping，是否正常连接。
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}
