package session

import "peeorm/log"

// 启动事务
func (s *Session) Begin() (err error) {
	log.Info("transaction begin")
	// 调用s.db.Begin得到*sql.tx赋值给s.tx
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

// 提交
func (s *Session) Commit() (err error) {
	log.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
		return
	}
	return
}

// 回滚
func (s *Session) Rollback() (err error) {
	log.Info("tarnsaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
		return
	}
	return
}
