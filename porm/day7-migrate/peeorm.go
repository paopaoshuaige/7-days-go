package main

import (
	"database/sql"
	"peeorm/dialect"
	"peeorm/log"
	"peeorm/session"
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
