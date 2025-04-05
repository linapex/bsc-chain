// 版权所有 2017 The go-ethereum Authors
// 本文件是go-ethereum库的一部分。
//
// go-ethereum库是自由软件：您可以自由修改和重新发布它，
// 在遵循GNU宽通用公共许可证（GNU Lesser General Public License）的前提下，
// 该许可证由自由软件基金会发布，版本3或（根据您的选择）任何后续版本。
//
// go-ethereum库的发布是希望它能够有用，
// 但不提供任何担保；甚至没有适销性或特定用途适用性的暗示担保。
// 更多详情请参见GNU宽通用公共许可证。
//
// 您应该已经收到了一份GNU宽通用公共许可证的副本。
// 如果没有，请访问<http://www.gnu.org/licenses/>。

package accounts

import (
	"fmt"
	"reflect"
	"testing"
)

// 测试 HD 派生路径能否正确解析为我们的内部二进制表示。
func TestHDPathParsing(t *testing.T) {
	// 测试HD路径解析
	t.Parallel()
	tests := []struct {
		input  string
		output DerivationPath
	}{
		// 普通绝对派生路径
		{"m/44'/60'/0'/0", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}},
		{"m/44'/60'/0'/128", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 128}},
		{"m/44'/60'/0'/0'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0x80000000 + 0}},
		{"m/44'/60'/0'/128'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0x80000000 + 128}},
		{"m/2147483692/2147483708/2147483648/0", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}},
		{"m/2147483692/2147483708/2147483648/2147483648", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0x80000000 + 0}},

		// 普通相对派生路径
		{"0", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0}},
		{"128", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 128}},
		{"0'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0x80000000 + 0}},
		{"128'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0x80000000 + 128}},
		{"2147483648", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0x80000000 + 0}},

		// 十六进制绝对派生路径
		{"m/0x2C'/0x3c'/0x00'/0x00", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}},
		{"m/0x2C'/0x3c'/0x00'/0x80", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 128}},
		{"m/0x2C'/0x3c'/0x00'/0x00'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0x80000000 + 0}},
		{"m/0x2C'/0x3c'/0x00'/0x80'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0x80000000 + 128}},
		{"m/0x8000002C/0x8000003c/0x80000000/0x00", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}},
		{"m/0x8000002C/0x8000003c/0x80000000/0x80000000", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0x80000000 + 0}},

		// 十六进制相对派生路径
		{"0x00", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0}},
		{"0x80", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 128}},
		{"0x00'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0x80000000 + 0}},
		{"0x80'", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0x80000000 + 128}},
		{"0x80000000", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0x80000000 + 0}},

		// 特殊输入以确保它们能正常工作
		{"	m  /   44			'\n/\n   60	\n\n\t'   /\n0 ' /\t\t	0", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}},

		// 无效的派生路径
		{"", nil},              // 空相对派生路径
		{"m", nil},             // 空绝对派生路径
		{"m/", nil},            // 缺少最后的派生组件
		{"/44'/60'/0'/0", nil}, // 没有m前缀的绝对路径，可能是用户错误
		{"m/2147483648'", nil}, // 溢出32位整数
		{"m/-1'", nil},         // 不能包含负数
	}
	for i, tt := range tests {
		if path, err := ParseDerivationPath(tt.input); !reflect.DeepEqual(path, tt.output) {
			t.Errorf("test %d: parse mismatch: have %v (%v), want %v", i, path, err, tt.output)
		} else if path == nil && err == nil {
			t.Errorf("test %d: nil path and error: %v", i, err)
		}
	}
}

func testDerive(t *testing.T, next func() DerivationPath, expected []string) {
	t.Helper()
	for i, want := range expected {
		if have := next(); fmt.Sprintf("%v", have) != want {
			t.Errorf("step %d, have %v, want %v", i, have, want)
		}
	}
}

func TestHdPathIteration(t *testing.T) {
	t.Parallel()
	testDerive(t, DefaultIterator(DefaultBaseDerivationPath),
		[]string{
			"m/44'/60'/0'/0/0", "m/44'/60'/0'/0/1",
			"m/44'/60'/0'/0/2", "m/44'/60'/0'/0/3",
			"m/44'/60'/0'/0/4", "m/44'/60'/0'/0/5",
			"m/44'/60'/0'/0/6", "m/44'/60'/0'/0/7",
			"m/44'/60'/0'/0/8", "m/44'/60'/0'/0/9",
		})

	testDerive(t, DefaultIterator(LegacyLedgerBaseDerivationPath),
		[]string{
			"m/44'/60'/0'/0", "m/44'/60'/0'/1",
			"m/44'/60'/0'/2", "m/44'/60'/0'/3",
			"m/44'/60'/0'/4", "m/44'/60'/0'/5",
			"m/44'/60'/0'/6", "m/44'/60'/0'/7",
			"m/44'/60'/0'/8", "m/44'/60'/0'/9",
		})

	testDerive(t, LedgerLiveIterator(DefaultBaseDerivationPath),
		[]string{
			"m/44'/60'/0'/0/0", "m/44'/60'/1'/0/0",
			"m/44'/60'/2'/0/0", "m/44'/60'/3'/0/0",
			"m/44'/60'/4'/0/0", "m/44'/60'/5'/0/0",
			"m/44'/60'/6'/0/0", "m/44'/60'/7'/0/0",
			"m/44'/60'/8'/0/0", "m/44'/60'/9'/0/0",
		})
}
