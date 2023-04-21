package GeeORM

import (
	"GeeORM/session"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatal("Fail to connect", err)
	}
	return engine
}

func TestNewEngine(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
}

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestEngine_Transaction(t *testing.T) {
	convey.Convey("ROLLBACK 测试", t, func() {
		engine := OpenDB(t)
		defer engine.Close()

		s := engine.NewSession()
		_ = s.Model(&User{}).DropTable()
		_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
			_ = s.Model(&User{}).CreateTable()
			_, err = s.Insert(&User{
				Name: "Tom",
				Age:  18,
			})
			return nil, errors.New("Error")
		})

		convey.So(err, convey.ShouldNotBeNil)
		ok := s.HasTable()
		convey.So(ok, convey.ShouldBeFalse)
	})

	convey.Convey("COMMIT 测试", t, func() {
		engine := OpenDB(t)
		defer engine.Close()

		s := engine.NewSession()
		_ = s.Model(&User{}).DropTable()
		_, err := engine.Transaction(func(session *session.Session) (result interface{}, err error) {
			_ = s.Model(&User{}).CreateTable()
			_, err = s.Insert(&User{
				Name: "Tom",
				Age:  18,
			})
			return
		})
		convey.So(err, convey.ShouldBeNil)
		u := &User{}
		err = s.First(u)
		convey.So(err, convey.ShouldBeNil)
		convey.So(u.Name, convey.ShouldEqual, "Tom")
	})
}

func TestEngine_Migrate(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()

	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text PRIMARY KEY, XXX integer);").Exec()
	_, _ = s.Raw("INSERT INTO User(`Name`) VALUES  (?), (?)", "Tom", "Sam").Exec()

	convey.Convey("MIGRATE 测试", t, func() {
		engine.Migrate(&User{})
		rows, _ := s.Raw("SELECT * FROM User").QueryRows()
		columns, _ := rows.Columns()
		convey.So(columns, convey.ShouldResemble, []string{"Name", "Age"})
	})

}
