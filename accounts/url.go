// 版权 2017 go-ethereum 作者
// 本文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是免费软件：您可以根据自由软件基金会发布的 GNU 较低版本通用公共许可证的条款，重新分发和/或修改它，版本 3 或
// （或任何更高版本）。
//
// go-ethereum 库分发的目的是希望它有用，
// 但没有任何保证；甚至没有隐含的
// 适销性或适合特定用途的保证。有关更多详细信息，请参阅
// GNU 较低版本通用公共许可证。
//
// 您应该已经收到与 go-ethereum 库一起的 GNU 较低版本通用公共许可证的副本。如果没有，请参阅 <http://www.gnu.org/licenses/>。

package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// URL 表示钱包或账户的规范标识URL。
//
// 它是 url.URL 的简化版本，具有重要的限制（在这里被视为特性）：
// 它仅包含可值拷贝的组件，并且不对特殊字符进行任何URL编码/解码。
//
// 前者对于允许账户被复制而不留下对原始版本的实时引用很重要，
// 而后者对于确保与RFC 3986规范允许的多种形式相对立的单一规范形式很重要。
//
// 因此，这些URL不应在以太坊钱包或账户范围之外使用。
type URL struct {
	Scheme string // 协议方案，用于标识有能力的账户后端
	Path   string // 路径，用于后端标识唯一实体
}

// parseURL 将用户提供的URL转换为账户特定的结构。
func parseURL(url string) (URL, error) {
	parts := strings.Split(url, "://")
	if len(parts) != 2 || parts[0] == "" {
		return URL{}, errors.New("protocol scheme missing")
	}
	return URL{
		Scheme: parts[0],
		Path:   parts[1],
	}, nil
}

// String 实现了 stringer 接口。
func (u URL) String() string {
	if u.Scheme != "" {
		return fmt.Sprintf("%s://%s", u.Scheme, u.Path)
	}
	return u.Path
}

// TerminalString 实现了 log.TerminalStringer 接口。
func (u URL) TerminalString() string {
	url := u.String()
	if len(url) > 32 {
		return url[:31] + ".."
	}
	return url
}

// MarshalJSON 实现了 json.Marshaller 接口。
func (u URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON 解析URL。
func (u *URL) UnmarshalJSON(input []byte) error {
	var textURL string
	err := json.Unmarshal(input, &textURL)
	if err != nil {
		return err
	}
	url, err := parseURL(textURL)
	if err != nil {
		return err
	}
	u.Scheme = url.Scheme
	u.Path = url.Path
	return nil
}

// Cmp 比较x和y并返回:
//
//	-1 如果 x <  y
//	 0 如果 x == y
//	+1 如果 x >  y
func (u URL) Cmp(url URL) int {
	if u.Scheme == url.Scheme {
		return strings.Compare(u.Path, url.Path)
	}
	return strings.Compare(u.Scheme, url.Scheme)
}
