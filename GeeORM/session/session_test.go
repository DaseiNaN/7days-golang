package session

import (
	"GeeORM/dialect"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smartystreets/goconvey/convey"
	"log"
	"os"
	"testing"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	TestDB, err = sql.Open("sqlite3", "../gee.db")
	if err != nil {
		log.Fatal("Failed to connect", err)
	}
	code := m.Run()
	_ = TestDB.Close()
	os.Exit(code)
}

func NewSession() *Session {
	d, _ := dialect.GetDialect("sqlite3")
	return New(TestDB, d)
}

func TestSession_Exec(t *testing.T) {
	s := NewSession()
	convey.Convey("SQL 语句执行测试", t, func() {
		tt := []struct {
			name string
			sql  string
		}{
			{name: "删除 Table", sql: "DROP TABLE IF EXISTS User;"},
			{name: "创建 Table", sql: "CREATE TABLE User(Name text);"},
			{name: "插入 Record", sql: "INSERT INTO User(`Name`) VALUES (?), (?)"},
			{name: "查询多行数据", sql: "SELECT COUNT(*) FROM User"},
		}

		for i, tc := range tt {
			convey.Convey(tc.name, func() {
				if i < 2 {
					_, err := s.Raw(tc.sql).Exec()
					convey.So(err, convey.ShouldBeNil)
				} else if i == 2 {
					result, err := s.Raw(tc.sql, "Tom", "Sam").Exec()
					convey.So(err, convey.ShouldBeNil)
					count, err := result.RowsAffected()
					convey.So(err, convey.ShouldBeNil)
					convey.So(count, convey.ShouldEqual, 2)
				} else {
					row := s.Raw(tc.sql).QueryRow()
					var count int
					err := row.Scan(&count)
					convey.So(err, convey.ShouldBeNil)
					convey.So(count, convey.ShouldEqual, 2)
				}
			})
		}
	})
}
