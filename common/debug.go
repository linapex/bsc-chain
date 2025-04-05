// 版权所有 2015 go-ethereum 作者
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
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

// Report 发出警告，要求用户向github跟踪器提交问题。
func Report(extra ...interface{}) {
	fmt.Fprintln(os.Stderr, "您遇到了一个难以重现的错误。请向开发者报告此问题 <3 https://github.com/ethereum/go-ethereum/issues")
	fmt.Fprintln(os.Stderr, extra...)

	_, file, line, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "%v:%v\n", file, line)

	debug.PrintStack()

	fmt.Fprintln(os.Stderr, "#### 错误！请报告此问题 ####")
}

// PrintDeprecationWarning 使用fmt.Println将给定字符串打印在框中。
func PrintDeprecationWarning(str string) {
	line := strings.Repeat("#", len(str)+4)
	emptyLine := strings.Repeat(" ", len(str))
	fmt.Printf(`
%s
# %s #
# %s #
# %s #
%s

`, line, emptyLine, str, emptyLine, line)
}
