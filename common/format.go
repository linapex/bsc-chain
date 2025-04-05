// 版权所有 2016 go-ethereum 作者
// 本文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是自由软件：您可以根据 GNU 较宽松通用公共许可证的条款重新分发和/或修改它，
// 该许可证由自由软件基金会发布，版本 3 或（根据您的选择）任何更高版本。
//
// go-ethereum 库的发布是希望它能有用，
// 但没有任何保证；甚至没有适销性或特定用途适用性的暗示保证。
// 有关更多详情，请参阅 GNU 较宽松通用公共许可证。
//
// 您应该已经收到一份 GNU 较宽松通用公共许可证的副本。
// 如果没有，请参阅 <http://www.gnu.org/licenses/>。

package common

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// PrettyDuration 是time.Duration值的美化打印版本，它会从格式化文本表示中去除不必要的精度。
type PrettyDuration time.Duration

var prettyDurationRe = regexp.MustCompile(`\.[0-9]{4,}`)

// String 实现了Stringer接口，允许对持续时间值进行美化打印，四舍五入到三位小数。
func (d PrettyDuration) String() string {
	label := time.Duration(d).String()
	if match := prettyDurationRe.FindString(label); len(match) > 4 {
		label = strings.Replace(label, match, match[:4], 1)
	}
	return label
}

// PrettyAge 是time.Duration值的美化打印版本，它会将值四舍五入到最重要的单个单位，包括天/周/年。
type PrettyAge time.Time

// ageUnits 是年龄美化打印使用的单位列表。
var ageUnits = []struct {
	Size   time.Duration
	Symbol string
}{
	{12 * 30 * 24 * time.Hour, "y"},
	{30 * 24 * time.Hour, "mo"},
	{7 * 24 * time.Hour, "w"},
	{24 * time.Hour, "d"},
	{time.Hour, "h"},
	{time.Minute, "m"},
	{time.Second, "s"},
}

// String 实现了Stringer接口，允许对持续时间值进行美化打印，四舍五入到最重要的时间单位。
func (t PrettyAge) String() string {
	// 计算时间差并处理0的边界情况
	diff := time.Since(time.Time(t))
	if diff < time.Second {
		return "0"
	}
	// 在返回前累积最多3个时间单位的精度
	result, prec := "", 0

	for _, unit := range ageUnits {
		if diff > unit.Size {
			result = fmt.Sprintf("%s%d%s", result, diff/unit.Size, unit.Symbol)
			diff %= unit.Size

			if prec += 1; prec >= 3 {
				break
			}
		}
	}
	return result
}

// FormatMilliTime 将毫秒时间戳格式化为可读字符串
func FormatMilliTime(n int64) string {
	if n < 0 {
		return "无效时间"
	}
	if n == 0 {
		return ""
	}
	return time.UnixMilli(n).Format("2006-01-02 15:04:05.000")
}
