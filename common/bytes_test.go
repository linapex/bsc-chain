// 版权所有 2014 go-ethereum 作者
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
// 如果没有，请参阅 <http://www.gnu.org/licenses/>.

package common

import (
	"bytes"
	"testing"
)

// TestCopyBytes 测试CopyBytes函数是否能正确复制字节切片
func TestCopyBytes(t *testing.T) {
	input := []byte{1, 2, 3, 4}

	v := CopyBytes(input)
	if !bytes.Equal(v, []byte{1, 2, 3, 4}) {
		t.Fatal("复制后内容不相等")
	}
	v[0] = 99
	if bytes.Equal(v, input) {
		t.Fatal("结果不是副本")
	}
}

// TestLeftPadBytes 测试LeftPadBytes函数的左侧填充功能
func TestLeftPadBytes(t *testing.T) {
	val := []byte{1, 2, 3, 4}
	padded := []byte{0, 0, 0, 0, 1, 2, 3, 4}

	if r := LeftPadBytes(val, 8); !bytes.Equal(r, padded) {
		t.Fatalf("LeftPadBytes(%v, 8) == %v", val, r)
	}
	if r := LeftPadBytes(val, 2); !bytes.Equal(r, val) {
		t.Fatalf("LeftPadBytes(%v, 2) == %v", val, r)
	}
}

// TestRightPadBytes 测试RightPadBytes函数的右侧填充功能
func TestRightPadBytes(t *testing.T) {
	val := []byte{1, 2, 3, 4}
	padded := []byte{1, 2, 3, 4, 0, 0, 0, 0}

	if r := RightPadBytes(val, 8); !bytes.Equal(r, padded) {
		t.Fatalf("RightPadBytes(%v, 8) == %v", val, r)
	}
	if r := RightPadBytes(val, 2); !bytes.Equal(r, val) {
		t.Fatalf("RightPadBytes(%v, 2) == %v", val, r)
	}
}

// TestFromHex 测试FromHex函数的十六进制字符串转换功能
func TestFromHex(t *testing.T) {
	input := "0x01"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("期望值 %x 实际得到 %x", expected, result)
	}
}

// TestIsHex 测试isHex函数的十六进制字符串验证功能
func TestIsHex(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"", true},
		{"0", false},
		{"00", true},
		{"a9e67e", true},
		{"A9E67E", true},
		{"0xa9e67e", false},
		{"a9e67e001", false},
		{"0xHELLO_MY_NAME_IS_STEVEN_@#$^&*", false},
	}
	for _, test := range tests {
		if ok := isHex(test.input); ok != test.ok {
			t.Errorf("isHex(%q) = %v, 期望值 %v", test.input, ok, test.ok)
		}
	}
}

// TestFromHexOddLength 测试FromHex函数处理奇数长度十六进制字符串的功能
func TestFromHexOddLength(t *testing.T) {
	input := "0x1"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("期望值 %x 实际得到 %x", expected, result)
	}
}

// TestNoPrefixShortHexOddLength 测试FromHex函数处理无前缀奇数长度十六进制字符串的功能
func TestNoPrefixShortHexOddLength(t *testing.T) {
	input := "1"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("期望值 %x 实际得到 %x", expected, result)
	}
}

// TestTrimRightZeroes 测试TrimRightZeroes函数的右侧零值去除功能
func TestTrimRightZeroes(t *testing.T) {
	tests := []struct {
		arr []byte
		exp []byte
	}{
		{FromHex("0x00ffff00ff0000"), FromHex("0x00ffff00ff")},
		{FromHex("0x00000000000000"), []byte{}},
		{FromHex("0xff"), FromHex("0xff")},
		{[]byte{}, []byte{}},
		{FromHex("0x00ffffffffffff"), FromHex("0x00ffffffffffff")},
	}
	for i, test := range tests {
		got := TrimRightZeroes(test.arr)
		if !bytes.Equal(got, test.exp) {
			t.Errorf("测试 %d, 得到 %x 期望值 %x", i, got, test.exp)
		}
	}
}
