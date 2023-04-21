package clause

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestClause_Build(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name = ?", "Tom")
	clause.Set(ORDERBY, "Age ASC")
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)

	convey.Convey("Clause 测试", t, func() {

		convey.Convey("SQL 构造", func() {
			expected := "SELECT * FROM User WHERE Name = ? ORDER BY Age ASC LIMIT ?"
			convey.So(sql, convey.ShouldEqual, expected)
		})

		convey.Convey("Vars 构造", func() {
			expected := []interface{}{"Tom", 3}
			convey.So(vars, convey.ShouldResemble, expected)
		})

	})

}
