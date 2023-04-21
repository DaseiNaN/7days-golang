package session

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	s := NewSession().Model(&User{})
	convey.Convey("表操作测试", t, func() {
		convey.Convey("DROP", func() {
			err := s.DropTable()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("CREATE", func() {
			err := s.CreateTable()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("HasTable", func() {
			ok := s.HasTable()
			convey.So(ok, convey.ShouldBeTrue)
		})
	})

}
