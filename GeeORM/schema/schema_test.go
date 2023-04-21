package schema

import (
	"GeeORM/dialect"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

var TestDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, TestDial)
	convey.Convey("解析测试", t, func() {
		convey.Convey("表名解析", func() {
			convey.So(schema.Name, convey.ShouldEqual, "User")
		})

		convey.Convey("表列解析", func() {
			convey.So(len(schema.Fields), convey.ShouldEqual, 2)
		})
		convey.Convey("表约束解析", func() {
			convey.So(schema.GetField("Name").Tag, convey.ShouldEqual, "PRIMARY KEY")
		})
	})

}
