// 版权 2014 The go-ethereum Authors
// 本文件是go-ethereum库的一部分。
//
// go-ethereum库是自由软件：您可以自由地重新分发和/或修改
// 本软件，遵循由自由软件基金会发布的GNU Lesser General Public License条款，
// 可以是该许可证的第3版，或（根据您的选择）任何后续版本。
//
// go-ethereum库的发布是希望它能有用，
// 但没有任何担保；甚至没有适销性或特定用途适用性的隐含担保。
// 详情请参阅GNU Lesser General Public License。
//
// 您应该已经收到一份GNU Lesser General Public License的副本
// 如果没有，请参阅<http://www.gnu.org/licenses/>。

package common

import (
	"testing"
)

func TestStorageSizeString(t *testing.T) {
	tests := []struct {
		size StorageSize
		str  string
	}{
		{2839274474874, "2.58 TiB"},
		{2458492810, "2.29 GiB"},
		{2381273, "2.27 MiB"},
		{2192, "2.14 KiB"},
		{12, "12.00 B"},
	}

	for _, test := range tests {
		if test.size.String() != test.str {
			t.Errorf("%f: 得到 %q, 期望 %q", float64(test.size), test.size.String(), test.str)
		}
	}
}

func TestStorageSizeTerminalString(t *testing.T) {
	tests := []struct {
		size StorageSize
		str  string
	}{
		{2839274474874, "2.58TiB"},
		{2458492810, "2.29GiB"},
		{2381273, "2.27MiB"},
		{2192, "2.14KiB"},
		{12, "12.00B"},
	}

	for _, test := range tests {
		if test.size.TerminalString() != test.str {
			t.Errorf("%f: 得到 %q, 期望 %q", float64(test.size), test.size.TerminalString(), test.str)
		}
	}
}
