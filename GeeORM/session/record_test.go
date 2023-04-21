package session

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	user1 = &User{
		Name: "Tom",
		Age:  18,
	}
	user2 = &User{
		Name: "Sam",
		Age:  25,
	}
	user3 = &User{
		Name: "Jack",
		Age:  25,
	}
)

func testRecordInit(t *testing.T) *Session {
	t.Helper()
	s := NewSession().Model(&User{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("failed init test records")
	}
	return s
}

func TestSession_Insert(t *testing.T) {
	s := testRecordInit(t)
	convey.Convey("INSERT 测试", t, func() {
		affected, err := s.Insert(user3)
		convey.So(affected, convey.ShouldEqual, 1)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSession_Find(t *testing.T) {
	s := testRecordInit(t)
	convey.Convey("FIND 测试", t, func() {
		var users []User
		err := s.Find(&users)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(users), convey.ShouldEqual, 2)
	})
}

func TestSession_Limit(t *testing.T) {
	s := testRecordInit(t)
	convey.Convey("LIMIT 测试", t, func() {
		var users []User
		err := s.Limit(1).Find(&users)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(users), convey.ShouldEqual, 1)
	})
}

func TestSession_Update(t *testing.T) {
	s := testRecordInit(t)
	convey.Convey("UPDATE 测试", t, func() {
		affected, err := s.Where("Name = ?", "Tom").Update("Age", 30)
		convey.So(err, convey.ShouldBeNil)
		u := &User{}
		err = s.OrderBy("Age DESC").First(u)
		convey.So(err, convey.ShouldBeNil)
		convey.So(affected, convey.ShouldEqual, 1)
		convey.So(u.Age, convey.ShouldEqual, 30)
	})
}

func TestSession_DeleteAndCount(t *testing.T) {
	s := testRecordInit(t)
	convey.Convey("DELETE&COUNT 测试", t, func() {
		affected, err := s.Where("Name = ?", "Tom").Delete()
		convey.So(err, convey.ShouldBeNil)
		convey.So(affected, convey.ShouldEqual, 1)
		count, err := s.Count()
		convey.So(err, convey.ShouldBeNil)
		convey.So(count, convey.ShouldEqual, 1)
	})
}
