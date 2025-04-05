// 版权所有 2017 The go-ethereum 作者
// 本文件是 go-ethereum 库的一部分。

// go-ethereum 库是免费软件：您可以根据 GNU 宽通用公共许可证的条款重新分发和/或修改
// 由自由软件基金会发布的许可证版本 3，或（根据您的选择）任何更高版本。

// go-ethereum 库的分发是希望它能有用，
// 但没有任何保证；甚至没有适销性或适用性方面的隐含保证。
// 有关详细信息，请参阅 GNU 宽通用公共许可证。

// 您应该已经收到与 go-ethereum 库一起的 GNU 宽通用公共许可证的副本。
// 如果没有，请参阅 <http://www.gnu.org/licenses/>。

package accounts

import (
	"errors"
	"fmt"
)

// ErrUnknownAccount 在任何请求的操作中返回，该操作没有后端提供指定的账户。
var ErrUnknownAccount = errors.New("未知账户")

// ErrUnknownWallet 在任何请求的操作中返回，该操作没有后端提供指定的钱包。
var ErrUnknownWallet = errors.New("未知钱包")

// ErrNotSupported 当从账户后端请求不支持的操作时返回。
var ErrNotSupported = errors.New("不支持")

// ErrInvalidPassphrase 当解密操作收到错误的密码时返回。
var ErrInvalidPassphrase = errors.New("密码错误")

// ErrWalletAlreadyOpen 如果尝试第二次打开钱包时返回。
var ErrWalletAlreadyOpen = errors.New("钱包已打开")

// ErrWalletClosed 如果钱包离线时返回。
var ErrWalletClosed = errors.New("钱包已关闭")

// AuthNeededError 在签名请求时由后端返回，用户需要提供进一步的身份验证才能成功签名。
//
// 这通常意味着需要提供密码，或者可能是
// 某些硬件设备显示的一次性PIN码。
type AuthNeededError struct {
	Needed string // 用户需要提供的额外认证信息
}

// NewAuthNeededError 创建一个新的认证错误，并设置所需的额外字段详情。
func NewAuthNeededError(needed string) error {
	return &AuthNeededError{
		Needed: needed,
	}
}

// Error 实现了标准的error接口。
func (err *AuthNeededError) Error() string {
	return fmt.Sprintf("需要认证: %s", err.Needed)
}
