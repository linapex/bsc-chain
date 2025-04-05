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
// 如果没有，请参阅 <http://www.gnu.org/licenses/>。

package common

import (
	"math/big"

	"github.com/holiman/uint256"
)

// 常用的大整数定义
var (
	Big1   = big.NewInt(1)   // 整数1
	Big2   = big.NewInt(2)   // 整数2
	Big3   = big.NewInt(3)   // 整数3
	Big0   = big.NewInt(0)   // 整数0
	Big32  = big.NewInt(32)  // 整数32
	Big256 = big.NewInt(256) // 整数256
	Big257 = big.NewInt(257) // 整数257

	U2560 = uint256.NewInt(0) // 无符号256位整数0
)
