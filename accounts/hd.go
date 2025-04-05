// 版权所有 2017 The go-ethereum Authors
// 本文件是go-ethereum库的一部分。
//
// go-ethereum库是自由软件：您可以自由修改和重新发布它
// 根据GNU Lesser General Public License的条款发布，该许可证由
// Free Software Foundation发布，可以是许可证的第3版，或者
// （由您选择）任何更高版本。
//
// go-ethereum库的发布是希望它有用，
// 但没有任何担保；甚至没有适销性或特定用途适用性的暗示担保。
// 详情请参阅GNU Lesser General Public License。
//
// 您应该已经收到了一份GNU Lesser General Public License的副本
// 如果没有，请参见<http://www.gnu.org/licenses/>。

package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// DefaultRootDerivationPath 是自定义派生路径的根路径
// 第一个账户将位于 m/44'/60'/0'/0，第二个账户位于 m/44'/60'/0'/1，以此类推
var DefaultRootDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}

// DefaultBaseDerivationPath 是自定义派生路径的基础路径
// 第一个账户将位于 m/44'/60'/0'/0/0，第二个账户位于 m/44'/60'/0'/0/1，以此类推
var DefaultBaseDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0, 0}

// LegacyLedgerBaseDerivationPath 是旧版Ledger设备的自定义派生路径
// 第一个账户将位于 m/44'/60'/0'/0，第二个账户位于 m/44'/60'/0'/1，以此类推
var LegacyLedgerBaseDerivationPath = DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}

// DerivationPath represents the computer friendly version of a hierarchical
// deterministic wallet account derivation path.
//
// The BIP-32 spec https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
// defines derivation paths to be of the form:
//
//	m / purpose' / coin_type' / account' / change / address_index
//
// The BIP-44 spec https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// defines that the `purpose` be 44' (or 0x8000002C) for crypto currencies, and
// SLIP-44 https://github.com/satoshilabs/slips/blob/master/slip-0044.md assigns
// the `coin_type` 60' (or 0x8000003C) to Ethereum.
//
// The root path for Ethereum is m/44'/60'/0'/0 according to the specification
// from https://github.com/ethereum/EIPs/issues/84, albeit it's not set in stone
// yet whether accounts should increment the last component or the children of
// that. We will go with the simpler approach of incrementing the last component.
type DerivationPath []uint32

// ParseDerivationPath 将用户指定的派生路径字符串转换为内部二进制表示形式。
//
// 完整派生路径需要以`m/`前缀开头，相对派生路径(将被附加到默认根路径)
// 不能在第一个元素前有前缀。空格会被忽略。
func ParseDerivationPath(path string) (DerivationPath, error) {
	var result DerivationPath

	// Handle absolute or relative paths
	components := strings.Split(path, "/")
	switch {
	case len(components) == 0:
		return nil, errors.New("empty derivation path")

	case strings.TrimSpace(components[0]) == "":
		return nil, errors.New("ambiguous path: use 'm/' prefix for absolute paths, or no leading '/' for relative ones")

	case strings.TrimSpace(components[0]) == "m":
		components = components[1:]

	default:
		result = append(result, DefaultRootDerivationPath...)
	}
	// All remaining components are relative, append one by one
	if len(components) == 0 {
		return nil, errors.New("empty derivation path") // Empty relative paths
	}
	for _, component := range components {
		// Ignore any user added whitespace
		component = strings.TrimSpace(component)
		var value uint32

		// Handle hardened paths
		if strings.HasSuffix(component, "'") {
			value = 0x80000000
			component = strings.TrimSpace(strings.TrimSuffix(component, "'"))
		}
		// Handle the non hardened component
		bigval, ok := new(big.Int).SetString(component, 0)
		if !ok {
			return nil, fmt.Errorf("invalid component: %s", component)
		}
		max := math.MaxUint32 - value
		if bigval.Sign() < 0 || bigval.Cmp(big.NewInt(int64(max))) > 0 {
			if value == 0 {
				return nil, fmt.Errorf("component %v out of allowed range [0, %d]", bigval, max)
			}
			return nil, fmt.Errorf("component %v out of allowed hardened range [0, %d]", bigval, max)
		}
		value += uint32(bigval.Uint64())

		// Append and repeat
		result = append(result, value)
	}
	return result, nil
}

// String 实现stringer接口，将二进制派生路径转换为其规范表示形式。
func (path DerivationPath) String() string {
	result := "m"
	for _, component := range path {
		var hardened bool
		if component >= 0x80000000 {
			component -= 0x80000000
			hardened = true
		}
		result = fmt.Sprintf("%s/%d", result, component)
		if hardened {
			result += "'"
		}
	}
	return result
}

// MarshalJSON 将派生路径转换为json序列化字符串
func (path DerivationPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(path.String())
}

// UnmarshalJSON 将json序列化字符串反序列化回派生路径
func (path *DerivationPath) UnmarshalJSON(b []byte) error {
	var dp string
	var err error
	if err = json.Unmarshal(b, &dp); err != nil {
		return err
	}
	*path, err = ParseDerivationPath(dp)
	return err
}

// DefaultIterator 创建一个BIP-32路径迭代器，通过增加最后一个组件来推进:
// 例如 m/44'/60'/0'/0/0, m/44'/60'/0'/0/1, m/44'/60'/0'/0/2, ... m/44'/60'/0'/0/N。
func DefaultIterator(base DerivationPath) func() DerivationPath {
	path := make(DerivationPath, len(base))
	copy(path[:], base[:])
	// Set it back by one, so the first call gives the first result
	path[len(path)-1]--
	return func() DerivationPath {
		path[len(path)-1]++
		return path
	}
}

// LedgerLiveIterator 为Ledger Live创建一个bip44路径迭代器。
// Ledger Live递增第三个组件而不是第五个组件
// 例如 m/44'/60'/0'/0/0, m/44'/60'/1'/0/0, m/44'/60'/2'/0/0, ... m/44'/60'/N'/0/0。
func LedgerLiveIterator(base DerivationPath) func() DerivationPath {
	path := make(DerivationPath, len(base))
	copy(path[:], base[:])
	// Set it back by one, so the first call gives the first result
	path[2]--
	return func() DerivationPath {
		// ledgerLivePathIterator iterates on the third component
		path[2]++
		return path
	}
}
