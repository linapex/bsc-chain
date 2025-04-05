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

// Package consensus 实现了不同的以太坊共识引擎。
package consensus

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	SystemAddress = common.HexToAddress("0xffffFFFfFFffffffffffffffFfFFFfffFFFfFFfE")
)

// ChainHeaderReader 定义了一小组方法，用于在区块头验证期间访问本地区块链。
type ChainHeaderReader interface {
	// Config 获取区块链的链配置。
	Config() *params.ChainConfig

	// GenesisHeader 获取链的创世区块头。
	GenesisHeader() *types.Header

	// CurrentHeader 从本地链获取当前区块头。
	CurrentHeader() *types.Header

	// GetHeader 通过哈希和编号从数据库获取区块头。
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber 通过编号从数据库获取区块头。
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash 通过哈希从数据库获取区块头。
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetTd 通过哈希和编号从数据库获取总难度。
	GetTd(hash common.Hash, number uint64) *big.Int

	// GetHighestVerifiedHeader 获取已验证的最高区块头。
	GetHighestVerifiedHeader() *types.Header

	// GetVerifiedBlockByHash 获取已验证的最高区块。
	GetVerifiedBlockByHash(hash common.Hash) *types.Header

	// ChasingHead 返回对等节点中最优链的头部。
	ChasingHead() *types.Header
}

type VotePool interface {
	FetchVoteByBlockHash(blockHash common.Hash) []*types.VoteEnvelope
}

// ChainReader 定义了一小组方法，用于在区块头和/或叔块验证期间访问本地区块链。
type ChainReader interface {
	ChainHeaderReader

	// GetBlock 通过哈希和编号从数据库获取区块。
	GetBlock(hash common.Hash, number uint64) *types.Block
}

// Engine is an algorithm agnostic consensus engine.
type Engine interface {
	// Author retrieves the Ethereum address of the account that minted the given
	// block, which may be different from the header's coinbase if a consensus
	// engine is based on signatures.
	Author(header *types.Header) (common.Address, error)

	// VerifyHeader 检查区块头是否符合给定引擎的共识规则。
	VerifyHeader(chain ChainHeaderReader, header *types.Header) error

	// VerifyHeaders 类似于VerifyHeader，但并发验证一批区块头。
	// 该方法返回一个退出通道用于中止操作，
	// 和一个结果通道用于获取异步验证结果（顺序与输入切片相同）。
	VerifyHeaders(chain ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error)

	// VerifyUncles 验证给定区块的叔块是否符合给定引擎的共识规则。
	VerifyUncles(chain ChainReader, block *types.Block) error

	// VerifyRequests 验证Requests和header.RequestsHash之间的一致性。
	VerifyRequests(header *types.Header, Requests [][]byte) error

	// NextInTurnValidator 返回下一个轮次的验证者地址
	NextInTurnValidator(chain ChainHeaderReader, header *types.Header) (common.Address, error)

	// Prepare 根据特定引擎的规则初始化区块头的共识字段。
	// 这些更改是内联执行的。
	Prepare(chain ChainHeaderReader, header *types.Header) error

	// Finalize 执行任何交易后的状态修改（例如区块奖励或处理提款）
	// 但不组装区块。
	//
	// 注意：状态数据库可能会更新以反映在最终化时发生的任何共识规则
	// （例如区块奖励）。
	Finalize(chain ChainHeaderReader, header *types.Header, state vm.StateDB, txs *[]*types.Transaction,
		uncles []*types.Header, withdrawals []*types.Withdrawal, receipts *[]*types.Receipt, systemTxs *[]*types.Transaction, usedGas *uint64, tracer *tracing.Hooks) error

	// FinalizeAndAssemble 执行任何交易后的状态修改（例如区块奖励或处理提款）
	// 并组装最终区块。
	//
	// 注意：区块头和状态数据库可能会更新以反映在最终化时发生的任何共识规则
	// （例如区块奖励）。
	FinalizeAndAssemble(chain ChainHeaderReader, header *types.Header, state *state.StateDB, body *types.Body, receipts []*types.Receipt, tracer *tracing.Hooks) (*types.Block, []*types.Receipt, error)

	// Seal 为给定的输入区块生成一个新的密封请求，并将结果推送到给定的通道。
	//
	// 注意：该方法立即返回并将异步发送结果。
	// 根据共识算法，可能会返回多个结果。
	Seal(chain ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error

	// SealHash 返回区块在密封前的哈希值。
	SealHash(header *types.Header) common.Hash

	// CalcDifficulty 是难度调整算法。它返回新区块应具有的难度值。
	CalcDifficulty(chain ChainHeaderReader, time uint64, parent *types.Header) *big.Int

	// APIs 返回此共识引擎提供的RPC API。
	APIs(chain ChainHeaderReader) []rpc.API

	// Delay 返回矿工可以提交交易的最大持续时间
	Delay(chain ChainReader, header *types.Header, leftOver *time.Duration) *time.Duration

	// Close 终止共识引擎维护的任何后台线程。
	Close() error
}

type PoSA interface {
	Engine

	// IsSystemTransaction 检查交易是否是系统交易
	IsSystemTransaction(tx *types.Transaction, header *types.Header) (bool, error)
	// IsSystemContract 检查地址是否是系统合约
	IsSystemContract(to *common.Address) bool
	// EnoughDistance 检查区块是否有足够的距离
	EnoughDistance(chain ChainReader, header *types.Header) bool
	// IsLocalBlock 检查区块是否是本地区块
	IsLocalBlock(header *types.Header) bool
	// GetJustifiedNumberAndHash 获取已证明的区块编号和哈希
	GetJustifiedNumberAndHash(chain ChainHeaderReader, headers []*types.Header) (uint64, common.Hash, error)
	// GetFinalizedHeader 获取已最终化的区块头
	GetFinalizedHeader(chain ChainHeaderReader, header *types.Header) *types.Header
	// VerifyVote 验证投票的有效性
	VerifyVote(chain ChainHeaderReader, vote *types.VoteEnvelope) error
	// IsActiveValidatorAt 检查在指定区块高度是否是活跃验证者
	IsActiveValidatorAt(chain ChainHeaderReader, header *types.Header, checkVoteKeyFn func(bLSPublicKey *types.BLSPublicKey) bool) bool
	// NextProposalBlock 获取下一个提案区块的信息
	NextProposalBlock(chain ChainHeaderReader, header *types.Header, proposer common.Address) (uint64, uint64, error)
}
