// Package session
// 描述: 与数据库进行交互, 比如 SQL 语句的执行等;
package session

import (
	"GeeORM/clause"
	"GeeORM/dialect"
	"GeeORM/log"
	"GeeORM/schema"
	"database/sql"
	"strings"
)

type Session struct {
	db       *sql.DB         // sql.Open() 方法连接数据库成功后返回的指针;
	sql      strings.Builder // 用来拼接 SQL 语句;
	sqlVars  []interface{}   // SQL 语句中的占位符的对应值;
	dialect  dialect.Dialect // SQL 数据库适配
	refTable *schema.Schema  // 表与对象的映射
	clause   clause.Clause   // 用于构造 SQL 语句
	tx       *sql.Tx         // 用于支持 SQL 事务
}

type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

func (s *Session) Clear() {
	s.sql.Reset()   // 清空 SQL 语句
	s.sqlVars = nil // 清空 SQL 变量
	s.clause = clause.Clause{}
}

func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

func (s *Session) Exec() (sql.Result, error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	result, err := s.DB().Exec(s.sql.String(), s.sqlVars...)
	if err != nil {
		log.Error(err)
	}
	return result, err
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (*sql.Rows, error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	rows, err := s.DB().Query(s.sql.String(), s.sqlVars...)
	if err != nil {
		log.Error(err)
	}
	return rows, err
}
