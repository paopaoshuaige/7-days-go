package peeorm

import (
	"database/sql"
	"day2-reflect-schema/dialect"
	"day2-reflect-schema/log"
	"day2-reflect-schema/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
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
