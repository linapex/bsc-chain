// 版权 2017 The go-ethereum Authors
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

package consensus

import "errors"

var (
	// ErrUnknownAncestor 当验证一个区块需要一个未知的祖先时返回。
	ErrUnknownAncestor = errors.New("unknown ancestor")

	// ErrPrunedAncestor 当验证一个区块需要一个已知但状态不可用的祖先时返回。
	ErrPrunedAncestor = errors.New("pruned ancestor")

	// ErrFutureBlock 当一个区块的时间戳在当前节点看来属于未来时返回。
	ErrFutureBlock = errors.New("block in the future")

	// ErrInvalidNumber 如果一个区块的编号不等于其父区块编号加一时返回。
	ErrInvalidNumber = errors.New("invalid block number")

	// ErrInvalidTerminalBlock 如果一个区块相对于终端总难度无效时返回。
	ErrInvalidTerminalBlock = errors.New("invalid terminal block")
)
