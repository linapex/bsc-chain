// 版权 2017 go-ethereum 作者
// 此文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是免费的软件：您可以根据自由软件基金会发布的 GNU 较低版本通用公共许可证的条款，重新分发和/或修改它，版本 3 或
// （或任何更高版本）。
//
// go-ethereum 库分发的目的是希望它有用，
// 但没有任何保证；甚至没有隐含的
// 适销性或适合特定用途的保证。有关更多详细信息，请参阅
// GNU 较低版本通用公共许可证。
//
// 您应该已经收到与 go-ethereum 库一起的 GNU 较低版本通用公共许可证的副本。如果没有，请参阅 <http://www.gnu.org/licenses/>。

// 包 accounts 实现了高级的以太坊账户管理。
package accounts

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"golang.org/x/crypto/sha3"
)

// Account 表示位于特定位置的以太坊账户，由可选的 URL 字段定义。
type Account struct {
	Address common.Address `json:"address"` // 从密钥派生的以太坊账户地址
	URL     URL            `json:"url"`     // 后端中的可选资源定位器
}

const (
	MimetypeDataWithValidator = "data/validator"
	MimetypeTypedData         = "data/typed"
	MimetypeClique            = "application/x-clique-header"
	MimetypeParlia            = "application/x-parlia-header"
	MimetypeTextPlain         = "text/plain"
)

// Wallet 表示可能包含一个或多个账户的软件或硬件钱包（从同一种子派生）。
type Wallet interface {
	// URL 检索此钱包可访问的规范路径。它
	// 被上层用来定义多个后端中所有钱包的排序顺序。
	URL() URL

	// Status 返回文本状态以帮助用户了解钱包的当前状态。
	// 它还返回一个错误，指示钱包可能遇到的任何故障。
	Status() (string, error)

	// Open 初始化对钱包实例的访问。它不是为了解锁或
	// 解密账户密钥，而只是为了建立与硬件钱包的连接
	// 和/或访问派生种子。
	//
	// 实现特定钱包实例的实现可能会使用或不使用密码参数。
	// 没有无密码打开方法的原因是为了努力实现统一的钱包处理，
	// 不受不同后端提供商的影响。
	//
	// 请注意，如果您打开钱包，您必须关闭它以释放任何分配的
	// 资源（在处理硬件钱包时尤其重要）。
	Open(passphrase string) error

	// Close 释放任何由打开的钱包实例持有的资源。
	Close() error

	// Accounts 检索钱包当前知道的签名账户列表。
	// 对于分层确定性钱包，列表将不是详尽的，
	// 而只是包含在账户派生期间显式固定的账户。
	Accounts() []Account

	// Contains 返回账户是否是此特定钱包的一部分。
	Contains(account Account) bool

	// Derive 尝试在指定的派生路径上显式派生分层确定性账户。
	// 如果请求，派生的账户将被添加到钱包的跟踪账户列表中。
	Derive(path DerivationPath, pin bool) (Account, error)

	// SelfDerive 设置一个基本账户派生路径，钱包尝试从中发现非零账户并自动将它们添加到跟踪账户列表中。
	//
	// 请注意，自我派生将增加指定路径的最后一个组件，而不是下降到子路径中，以允许从非零组件开始发现账户。
	//
	// 一些硬件钱包在其演变过程中切换了派生路径，因此
	// 此方法支持提供多个基础以发现旧用户账户。
	// 只有最后一个基础将用于派生下一个空账户。
	//
	// 您可以通过调用 SelfDerive 并传递 nil 链状态读取器来禁用自动账户发现。
	SelfDerive(bases []DerivationPath, chain ethereum.ChainStateReader)

	// SignData 请求钱包签署给定数据的哈希
	// 它通过其地址包含在内的账户来查找指定的账户，
	// 或者可选地通过嵌入的 URL 字段中的任何位置元数据来查找。
	//
	// 如果钱包需要额外的身份验证来签署请求（例如
	// 解密账户的密码，或验证交易的 PIN 码），
	// 将返回一个 AuthNeededError 实例，其中包含有关用户需要哪些字段或操作的信息。
	// 用户可以通过 SignDataWithPassphrase 提供所需的详细信息来重试，或通过其他方式（例如在密钥库中解锁账户）。
	SignData(account Account, mimeType string, data []byte) ([]byte, error)

	// SignDataWithPassphrase 与 SignData 相同，但也需要密码
	// 注意：有可能错误调用可能会混淆两个字符串，并
	// 在 mimetype 字段中提供密码，反之亦然。因此，实施
	// 不应回显 mimetype 或在错误响应中返回 mimetype
	SignDataWithPassphrase(account Account, passphrase, mimeType string, data []byte) ([]byte, error)

	// SignText 请求钱包签署给定数据的哈希，前缀
	// 以太坊前缀方案
	// 它通过其地址包含在内的账户来查找指定的账户，
	// 或者可选地通过嵌入的 URL 字段中的任何位置元数据来查找。
	//
	// 如果钱包需要额外的身份验证来签署请求（例如
	// 解密账户的密码，或验证交易的 PIN 码），
	// 将返回一个 AuthNeededError 实例，其中包含有关用户需要哪些字段或操作的信息。
	// 用户可以通过 SignTextWithPassphrase 提供所需的详细信息来重试，或通过其他方式（例如在密钥库中解锁账户）。
	//
	// 此方法应返回“规范”格式的签名，v 为 0 或 1。
	SignText(account Account, text []byte) ([]byte, error)

	// SignTextWithPassphrase 与 Signtext 相同，但也需要密码
	SignTextWithPassphrase(account Account, passphrase string, hash []byte) ([]byte, error)

	// SignTx 请求钱包签署给定的交易。
	//
	// 它通过其地址包含在内的账户来查找指定的账户，
	// 或者可选地通过嵌入的 URL 字段中的任何位置元数据来查找。
	//
	// 如果钱包需要额外的身份验证来签署请求（例如
	// 解密账户的密码，或验证交易的 PIN 码），
	// 将返回一个 AuthNeededError 实例，其中包含有关用户需要哪些字段或操作的信息。
	// 用户可以通过 SignTxWithPassphrase 提供所需的详细信息来重试，或通过其他方式（例如在密钥库中解锁账户）。
	SignTx(account Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)

	// SignTxWithPassphrase 与 SignTx 相同，但也需要密码
	SignTxWithPassphrase(account Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
}

// Backend 是一个“钱包提供者”，可能包含一批账户，他们可以
// 签署交易并根据请求进行操作。
type Backend interface {
	// Wallets 检索后端当前知道的钱包列表。
	//
	// 返回的钱包默认情况下未打开。对于软件 HD 钱包，这
	// 意味着没有基础种子被解密，对于硬件钱包，没有实际
	// 连接被建立。
	//
	// 结果钱包列表将根据后端分配的内部 URL 按字母顺序排序。
	// 由于钱包（尤其是硬件）可能会来来去去，同一个钱包可能会在
	// 后续检索期间出现在列表中的不同位置。
	Wallets() []Wallet

	// Subscribe 创建一个异步订阅，以接收后端检测到钱包到达或离开时的通知。
	Subscribe(sink chan<- WalletEvent) event.Subscription
}

// TextHash 是一个辅助函数，用于计算给定消息的哈希，可以
// 安全地用于计算签名。
//
// 哈希计算为
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message})。
//
// 这为签名的消息提供了上下文，并防止签署交易。
func TextHash(data []byte) []byte {
	hash, _ := TextAndHash(data)
	return hash
}

// TextAndHash 是一个辅助函数，用于计算给定消息的哈希，可以
// 安全地用于计算签名。
//
// 哈希计算为
//
//	keccak256("\x19Ethereum Signed Message:\n"${message length}${message})。
//
// 这为签名的消息提供了上下文，并防止签署交易。
func TextAndHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(msg))
	return hasher.Sum(nil), msg
}

// WalletEventType 表示钱包订阅子系统可以触发的不同事件类型。
type WalletEventType int

const (
	// WalletArrived 在通过 USB 或通过密钥库中的文件系统事件检测到新钱包时触发。
	WalletArrived WalletEventType = iota

	// WalletOpened 在成功打开钱包以启动任何后台进程（例如自动密钥派生）时触发。
	WalletOpened

	// WalletDropped 在钱包被移除或断开连接时触发，无论是通过 USB
	// 还是由于密钥库中的文件系统事件。此事件表明钱包
	// 不再可用于操作。
	WalletDropped
)

// WalletEvent 是账户后端在检测到钱包到达或离开时触发的事件。
type WalletEvent struct {
	Wallet Wallet          // 到达或离开的钱包实例
	Kind   WalletEventType // 系统中发生的事件类型
}
