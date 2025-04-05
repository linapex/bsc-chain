// 版权所有 2018 go-ethereum 作者
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

package accounts

import (
	"testing"
)

func TestURLParsing(t *testing.T) {
	t.Parallel()
	url, err := parseURL("https://ethereum.org")
	if err != nil {
		t.Errorf("意外错误: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("预期: %v, 实际: %v", "https", url.Scheme)
	}
	if url.Path != "ethereum.org" {
		t.Errorf("预期: %v, 实际: %v", "ethereum.org", url.Path)
	}

	for _, u := range []string{"ethereum.org", ""} {
		if _, err = parseURL(u); err == nil {
			t.Errorf("输入 %v, 预期错误, 实际: nil", u)
		}
	}
}

func TestURLString(t *testing.T) {
	t.Parallel()
	url := URL{Scheme: "https", Path: "ethereum.org"}
	if url.String() != "https://ethereum.org" {
		t.Errorf("预期: %v, 实际: %v", "https://ethereum.org", url.String())
	}

	url = URL{Scheme: "", Path: "ethereum.org"}
	if url.String() != "ethereum.org" {
		t.Errorf("预期: %v, 实际: %v", "ethereum.org", url.String())
	}
}

func TestURLMarshalJSON(t *testing.T) {
	t.Parallel()
	url := URL{Scheme: "https", Path: "ethereum.org"}
	json, err := url.MarshalJSON()
	if err != nil {
		t.Errorf("意外错误: %v", err)
	}
	if string(json) != "\"https://ethereum.org\"" {
		t.Errorf("预期: %v, 实际: %v", "\"https://ethereum.org\"", string(json))
	}
}

func TestURLUnmarshalJSON(t *testing.T) {
	t.Parallel()
	url := &URL{}
	err := url.UnmarshalJSON([]byte("\"https://ethereum.org\""))
	if err != nil {
		t.Errorf("意外错误: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("预期: %v, 实际: %v", "https", url.Scheme)
	}
	if url.Path != "ethereum.org" {
		t.Errorf("预期: %v, 实际: %v", "https", url.Path)
	}
}

func TestURLComparison(t *testing.T) {
	t.Parallel()
	tests := []struct {
		urlA   URL
		urlB   URL
		expect int
	}{
		{URL{"https", "ethereum.org"}, URL{"https", "ethereum.org"}, 0},
		{URL{"http", "ethereum.org"}, URL{"https", "ethereum.org"}, -1},
		{URL{"https", "ethereum.org/a"}, URL{"https", "ethereum.org"}, 1},
		{URL{"https", "abc.org"}, URL{"https", "ethereum.org"}, -1},
	}

	for i, tt := range tests {
		result := tt.urlA.Cmp(tt.urlB)
		if result != tt.expect {
			t.Errorf("测试 %d: 比较不匹配: 预期: %d, 实际: %d", i, tt.expect, result)
		}
	}
}
