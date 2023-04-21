package session

import (
	"GeeORM/log"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

type Account struct {
	ID       int `geeorm:"PRIMARY KEY"`
	Password string
}

func (account *Account) BeforeInsert(s *Session) error {
	log.Info("Before Insert", account)
	account.ID += 1000
	return nil
}

func (account *Account) AfterQuery(s *Session) error {
	log.Info("After Query", account)
	account.Password = "******"
	return nil
}

var (
	acc1 = &Account{
		ID:       1,
		Password: "123456",
	}
	acc2 = &Account{
		ID:       2,
		Password: "qwerty",
	}
)

func testHooksInit(t *testing.T) *Session {
	t.Helper()
	s := NewSession().Model(&Account{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(acc1, acc2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("failed init test records")
	}
	return s
}

func TestSession_CallMethod(t *testing.T) {
	s := testHooksInit(t)

	convey.Convey("Call 测试", t, func() {
		a := &Account{}
		err := s.First(a)
		convey.So(err, convey.ShouldBeNil)
		convey.So(a.ID, convey.ShouldEqual, 1001)
		convey.So(a.Password, convey.ShouldEqual, "******")
	})
}
