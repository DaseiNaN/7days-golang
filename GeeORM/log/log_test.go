package log

import (
	"github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestSetLevel(t *testing.T) {
	convey.Convey("设置日志级别测试", t, func() {
		tt := []struct {
			name  string
			level int
		}{
			{name: "ErrorLevel", level: ErrorLevel},
			{name: "Disabled", level: Disabled},
		}
		for _, tc := range tt {
			convey.Convey(tc.name, func() {
				SetLevel(tc.level)
				if tc.level == ErrorLevel {
					convey.So(infoLog.Writer(), convey.ShouldNotEqual, os.Stdout)
					convey.So(errorLog.Writer(), convey.ShouldEqual, os.Stdout)
				} else if tc.level == Disabled {
					convey.So(infoLog.Writer(), convey.ShouldNotEqual, os.Stdout)
					convey.So(errorLog.Writer(), convey.ShouldNotEqual, os.Stdout)
				}

			})
		}
	})
}
