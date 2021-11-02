package compare

import (
	"github.com/itnotebooks/fireye/config"
	"time"
)

// 定位最近异常时间
func CompareTime(t time.Time) bool {

	// 判断最近的异常信息发生时间是否在检测范围内
	someTime_ago := int64(config.GLOBAL_CONFIG.Minutes * 60)

	return time.Now().Unix()-t.Unix() <= someTime_ago
}
