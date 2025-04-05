// 版权所有 2022 go-ethereum 作者
// 这个文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是免费软件：您可以根据
// 自由软件基金会发布的 GNU 宽通用公共许可证的条款重新分发和/或修改它，
// 可以选择使用版本3，或任何更新的版本。
//
// go-ethereum 库的发布是希望它能够有用，
// 但不提供任何保证；甚至不提供针对特定用途的适销性或适用性的暗示保证。
// 更多详细信息，请参阅 GNU 宽通用公共许可证。
//
// 您应该已经收到了 GNU 宽通用公共许可证的副本
// 和 go-ethereum 库一起。如果没有，请参阅 <http://www.gnu.org/licenses/>。

package miner

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// BuildPayloadArgs 包含构建payload所需的参数。
// 更多细节请查看engine-api规范文档。
// https://github.com/ethereum/execution-apis/blob/main/src/engine/cancun.md#payloadattributesv3
type BuildPayloadArgs struct {
	Parent       common.Hash           // The parent block to build payload on top
	Timestamp    uint64                // The provided timestamp of generated payload
	FeeRecipient common.Address        // The provided recipient address for collecting transaction fee
	Random       common.Hash           // The provided randomness value
	Withdrawals  types.Withdrawals     // The provided withdrawals
	BeaconRoot   *common.Hash          // The provided beaconRoot (Cancun)
	Version      engine.PayloadVersion // Versioning byte for payload id calculation.
}

// Id 通过对payload参数的组件进行哈希计算来生成一个8字节的标识符。
func (args *BuildPayloadArgs) Id() engine.PayloadID {
	hasher := sha256.New()
	hasher.Write(args.Parent[:])
	binary.Write(hasher, binary.BigEndian, args.Timestamp)
	hasher.Write(args.Random[:])
	hasher.Write(args.FeeRecipient[:])
	rlp.Encode(hasher, args.Withdrawals)
	if args.BeaconRoot != nil {
		hasher.Write(args.BeaconRoot[:])
	}
	var out engine.PayloadID
	copy(out[:], hasher.Sum(nil)[:8])
	out[0] = byte(args.Version)
	return out
}

// Payload封装了构建好的payload(等待密封的区块)。根据engine-api规范，
// EL应该先构建一个初始版本的payload(包含空交易集)，然后持续更新它以最大化收益。
// 因此，这里空区块总是可用的，完整区块将在之后设置/更新。
type Payload struct {
	id            engine.PayloadID
	empty         *types.Block
	emptyWitness  *stateless.Witness
	full          *types.Block
	fullWitness   *stateless.Witness
	sidecars      []*types.BlobTxSidecar
	emptyRequests [][]byte
	requests      [][]byte
	fullFees      *big.Int
	stop          chan struct{}
	lock          sync.Mutex
	cond          *sync.Cond
}

// newPayload 初始化payload对象。
func newPayload(empty *types.Block, emptyRequests [][]byte, witness *stateless.Witness, id engine.PayloadID) *Payload {
	payload := &Payload{
		id:            id,
		empty:         empty,
		emptyRequests: emptyRequests,
		emptyWitness:  witness,
		stop:          make(chan struct{}),
	}
	log.Info("Starting work on payload", "id", payload.id)
	payload.cond = sync.NewCond(&payload.lock)
	return payload
}

// update 使用最新构建的版本更新完整区块。
func (payload *Payload) update(r *newPayloadResult, elapsed time.Duration) {
	payload.lock.Lock()
	defer payload.lock.Unlock()

	select {
	case <-payload.stop:
		return // 拒绝过期的更新
	default:
	}
	// 确保新提供的完整区块具有更高的交易费用。
	// 在合并后阶段，不再有叔块奖励，交易费用（除了MEV收入外）
	// 是唯一的比较指标。
	if payload.full == nil || r.fees.Cmp(payload.fullFees) > 0 {
		payload.full = r.block
		payload.fullFees = r.fees
		payload.sidecars = r.sidecars
		payload.requests = r.requests
		payload.fullWitness = r.witness

		feesInEther := new(big.Float).Quo(new(big.Float).SetInt(r.fees), big.NewFloat(params.Ether))
		log.Info("Updated payload",
			"id", payload.id,
			"number", r.block.NumberU64(),
			"hash", r.block.Hash(),
			"txs", len(r.block.Transactions()),
			"withdrawals", len(r.block.Withdrawals()),
			"gas", r.block.GasUsed(),
			"fees", feesInEther,
			"root", r.block.Root(),
			"elapsed", common.PrettyDuration(elapsed),
		)
	}
	payload.cond.Broadcast() // 发送信号通知完整区块已更新
}

// Resolve 返回最新构建的payload并终止后台更新线程。可安全多次调用。
func (payload *Payload) Resolve() *engine.ExecutionPayloadEnvelope {
	payload.lock.Lock()
	defer payload.lock.Unlock()

	select {
	case <-payload.stop:
	default:
		close(payload.stop)
	}
	if payload.full != nil {
		envelope := engine.BlockToExecutableData(payload.full, payload.fullFees, payload.sidecars, payload.requests)
		if payload.fullWitness != nil {
			envelope.Witness = new(hexutil.Bytes)
			*envelope.Witness, _ = rlp.EncodeToBytes(payload.fullWitness) // 不会失败
		}
		return envelope
	}
	envelope := engine.BlockToExecutableData(payload.empty, big.NewInt(0), nil, payload.emptyRequests)
	if payload.emptyWitness != nil {
		envelope.Witness = new(hexutil.Bytes)
		*envelope.Witness, _ = rlp.EncodeToBytes(payload.emptyWitness) // 不会失败
	}
	return envelope
}

// ResolveEmpty 基本与Resolve相同，但只期望空区块。仅用于测试。
func (payload *Payload) ResolveEmpty() *engine.ExecutionPayloadEnvelope {
	payload.lock.Lock()
	defer payload.lock.Unlock()

	envelope := engine.BlockToExecutableData(payload.empty, big.NewInt(0), nil, payload.emptyRequests)
	if payload.emptyWitness != nil {
		envelope.Witness = new(hexutil.Bytes)
		*envelope.Witness, _ = rlp.EncodeToBytes(payload.emptyWitness) // 不会失败
	}
	return envelope
}

// ResolveFull 基本与Resolve相同，但只期望完整区块。
// 在ResolveFull返回前不要调用Resolve，否则可能永久阻塞。
func (payload *Payload) ResolveFull() *engine.ExecutionPayloadEnvelope {
	payload.lock.Lock()
	defer payload.lock.Unlock()

	if payload.full == nil {
		select {
		case <-payload.stop:
			return nil
		default:
		}
		// 等待完整payload的构建。注意如果同时调用了Resolve
		// 会终止后台构建进程，这可能导致永久阻塞。
		payload.cond.Wait()
	}
	// 终止后台payload构建
	select {
	case <-payload.stop:
	default:
		close(payload.stop)
	}
	envelope := engine.BlockToExecutableData(payload.full, payload.fullFees, payload.sidecars, payload.requests)
	if payload.fullWitness != nil {
		envelope.Witness = new(hexutil.Bytes)
		*envelope.Witness, _ = rlp.EncodeToBytes(payload.fullWitness) // 不会失败
	}
	return envelope
}

// buildPayload builds the payload according to the provided parameters.
func (w *worker) buildPayload(args *BuildPayloadArgs, witness bool) (*Payload, error) {
	// 构建不包含任何交易的初始版本。它应该足够快可以运行。
	// 空payload至少可以确保不会因为错过slot而无法交付。
	emptyParams := &generateParams{
		timestamp:   args.Timestamp,
		forceTime:   true,
		parentHash:  args.Parent,
		coinbase:    args.FeeRecipient,
		random:      args.Random,
		withdrawals: args.Withdrawals,
		beaconRoot:  args.BeaconRoot,
		noTxs:       true,
	}
	empty := w.getSealingBlock(emptyParams)
	if empty.err != nil {
		return nil, empty.err
	}
	// Construct a payload object for return.
	payload := newPayload(empty.block, empty.requests, empty.witness, args.Id())

	// 启动一个后台更新payload的例程。这个策略可以最大化包含最高手续费交易的收益。
	go func() {
		// Setup the timer for re-building the payload. The initial clock is kept
		// for triggering process immediately.
		timer := time.NewTimer(0)
		defer timer.Stop()

		// Setup the timer for terminating the process if SECONDS_PER_SLOT (12s in
		// the Mainnet configuration) have passed since the point in time identified
		// by the timestamp parameter.
		endTimer := time.NewTimer(time.Second * 12)

		fullParams := &generateParams{
			timestamp:   args.Timestamp,
			forceTime:   true,
			parentHash:  args.Parent,
			coinbase:    args.FeeRecipient,
			random:      args.Random,
			withdrawals: args.Withdrawals,
			beaconRoot:  args.BeaconRoot,
			noTxs:       false,
		}

		for {
			select {
			case <-timer.C:
				start := time.Now()
				r := w.getSealingBlock(fullParams)
				if r.err == nil {
					payload.update(r, time.Since(start))
				} else {
					log.Info("Error while generating work", "id", payload.id, "err", r.err)
				}
				timer.Reset(w.recommit)
			case <-payload.stop:
				log.Info("Stopping work on payload", "id", payload.id, "reason", "delivery")
				return
			case <-endTimer.C:
				log.Info("Stopping work on payload", "id", payload.id, "reason", "timeout")
				return
			}
		}
	}()
	return payload, nil
}
