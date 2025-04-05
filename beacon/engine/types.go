// 版权所有 2022 The go-ethereum 作者
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

package engine // 引擎包实现以太坊执行层与共识层的交互协议

import (
	"fmt" // 格式化
	"math/big" // 大数运算
	"slices" // 切片操作

	"github.com/ethereum/go-ethereum/common" // 通用工具
	"github.com/ethereum/go-ethereum/common/hexutil" // 十六进制工具
	"github.com/ethereum/go-ethereum/core/types" // 核心类型
	"github.com/ethereum/go-ethereum/params" // 参数配置
	"github.com/ethereum/go-ethereum/trie" // 字典树
)

// PayloadVersion 表示用于请求开始构建payload的PayloadAttributes版本
type PayloadVersion byte

var (
	PayloadV1 PayloadVersion = 0x1 // 版本1 - 基础payload版本
	PayloadV2 PayloadVersion = 0x2 // 版本2 - 支持提款功能的payload版本
	PayloadV3 PayloadVersion = 0x3 // 版本3 - 支持EIP-4844 blob交易的payload版本
)

//go:generate go run github.com/fjl/gencodec -type PayloadAttributes -field-override payloadAttributesMarshaling -out gen_blockparams.go

// PayloadAttributes 描述了构建区块时的环境上下文
// 包含执行层构建新区块所需的所有参数
// 这些参数由共识层(CL)提供给执行层(EL)
type PayloadAttributes struct {
	Timestamp             uint64              `json:"timestamp"             gencodec:"required"` // 区块时间戳(Unix秒)
	Random                common.Hash         `json:"prevRandao"            gencodec:"required"` // 随机数(来自信标链的RANDAO混合值)
	SuggestedFeeRecipient common.Address      `json:"suggestedFeeRecipient" gencodec:"required"` // 建议的矿工地址(接收交易费用的账户)
	Withdrawals           []*types.Withdrawal `json:"withdrawals"` // 提款列表(从信标链到执行层的提款操作)
	BeaconRoot            *common.Hash        `json:"parentBeaconBlockRoot"` // 父信标区块根(可选，用于EIP-4788)
}

// payloadAttributesMarshaling 是 PayloadAttributes 的 JSON 类型覆盖
type payloadAttributesMarshaling struct {
	Timestamp hexutil.Uint64 // 时间戳(十六进制格式)
}

//go:generate go run github.com/fjl/gencodec -type ExecutableData -field-override executableDataMarshaling -out gen_ed.go

// ExecutableData 是执行EL payload所需的数据
// 包含执行层构建新区块所需的所有数据，由执行层提供给共识层
// 这些数据将被共识层用于验证和执行区块
type ExecutableData struct {
	ParentHash       common.Hash             `json:"parentHash"    gencodec:"required"` // 父区块哈希(区块链连接的关键字段)
	FeeRecipient     common.Address          `json:"feeRecipient"  gencodec:"required"` // 矿工地址(接收交易费用的账户)
	StateRoot        common.Hash             `json:"stateRoot"     gencodec:"required"` // 状态根(执行所有交易后的状态树根哈希)
	ReceiptsRoot     common.Hash             `json:"receiptsRoot"  gencodec:"required"` // 收据根(所有交易收据的Merkle根哈希)
	LogsBloom        []byte                  `json:"logsBloom"     gencodec:"required"` // 日志布隆过滤器(压缩的日志事件数据)
	Random           common.Hash             `json:"prevRandao"    gencodec:"required"` // 随机数(来自信标链的RANDAO混合值)
	Number           uint64                  `json:"blockNumber"   gencodec:"required"` // 区块号(区块高度)
	GasLimit         uint64                  `json:"gasLimit"      gencodec:"required"` // Gas限制(区块允许的最大Gas总量)
	GasUsed          uint64                  `json:"gasUsed"       gencodec:"required"` // 已用Gas(区块中交易实际消耗的Gas总量)
	Timestamp        uint64                  `json:"timestamp"     gencodec:"required"` // 时间戳(区块创建时的Unix秒)
	ExtraData        []byte                  `json:"extraData"     gencodec:"required"` // 额外数据(区块的额外信息字段)
	BaseFeePerGas    *big.Int                `json:"baseFeePerGas" gencodec:"required"` // 基础Gas费(EIP-1559引入的动态基础费用)
	BlockHash        common.Hash             `json:"blockHash"     gencodec:"required"` // 区块哈希(当前区块的Keccak256哈希)
	Transactions     [][]byte                `json:"transactions"  gencodec:"required"` // 交易列表(RLP编码的交易数据)
	Withdrawals      []*types.Withdrawal     `json:"withdrawals"` // 提款列表(从信标链到执行层的提款操作)
	BlobGasUsed      *uint64                 `json:"blobGasUsed"` // Blob Gas使用量(EIP-4844 blob交易使用的Gas)
	ExcessBlobGas    *uint64                 `json:"excessBlobGas"` // 超额Blob Gas(EIP-4844 blob交易的目标Gas调整)
	ExecutionWitness *types.ExecutionWitness `json:"executionWitness,omitempty"` // 执行见证(可选，用于无状态执行验证)
}

// executableDataMarshaling 是 ExecutableData 的 JSON 类型覆盖
type executableDataMarshaling struct {
	Number        hexutil.Uint64 // 区块号(十六进制格式)
	GasLimit      hexutil.Uint64 // Gas限制(十六进制格式)
	GasUsed       hexutil.Uint64 // 已用Gas(十六进制格式)
	Timestamp     hexutil.Uint64 // 时间戳(十六进制格式)
	BaseFeePerGas *hexutil.Big   // 基础Gas费(十六进制大数)
	ExtraData     hexutil.Bytes  // 额外数据(十六进制格式)
	LogsBloom     hexutil.Bytes  // 日志布隆过滤器(十六进制格式)
	Transactions  []hexutil.Bytes // 交易列表(十六进制格式)
	BlobGasUsed   *hexutil.Uint64 // Blob Gas使用量(十六进制格式)
	ExcessBlobGas *hexutil.Uint64 // 超额Blob Gas(十六进制格式)
}

// StatelessPayloadStatusV1 是无状态payload执行的结果
type StatelessPayloadStatusV1 struct {
	Status          string      `json:"status"` // 状态
	StateRoot       common.Hash `json:"stateRoot"` // 状态根
	ReceiptsRoot    common.Hash `json:"receiptsRoot"` // 收据根
	ValidationError *string     `json:"validationError"` // 验证错误
}

//go:generate go run github.com/fjl/gencodec -type ExecutionPayloadEnvelope -field-override executionPayloadEnvelopeMarshaling -out gen_epe.go

// ExecutionPayloadEnvelope 是执行payload的封装
type ExecutionPayloadEnvelope struct {
	ExecutionPayload *ExecutableData `json:"executionPayload"  gencodec:"required"` // 执行payload
	BlockValue       *big.Int        `json:"blockValue"  gencodec:"required"` // 区块价值
	BlobsBundle      *BlobsBundleV1  `json:"blobsBundle"` // Blobs包
	Requests         [][]byte        `json:"executionRequests"` // 执行请求
	Override         bool            `json:"shouldOverrideBuilder"` // 是否覆盖构建器
	Witness          *hexutil.Bytes  `json:"witness,omitempty"` // 见证数据
}

// BlobsBundleV1 包含Blob相关数据的集合
type BlobsBundleV1 struct {
	Commitments []hexutil.Bytes `json:"commitments"` // 承诺列表
	Proofs      []hexutil.Bytes `json:"proofs"`      // 证明列表
	Blobs       []hexutil.Bytes `json:"blobs"`       // Blob数据列表
}

// BlobAndProofV1 包含单个Blob及其证明
type BlobAndProofV1 struct {
	Blob  hexutil.Bytes `json:"blob"`  // Blob数据
	Proof hexutil.Bytes `json:"proof"` // 对应的证明
}

// executionPayloadEnvelopeMarshaling 是 ExecutionPayloadEnvelope 的 JSON 类型覆盖
type executionPayloadEnvelopeMarshaling struct {
	BlockValue *hexutil.Big   // 区块价值(十六进制大数)
	Requests   []hexutil.Bytes // 执行请求(十六进制格式)
}

// PayloadStatusV1 表示payload的状态信息
type PayloadStatusV1 struct {
	Status          string         `json:"status"`          // 状态
	Witness         *hexutil.Bytes `json:"witness"`         // 见证数据
	LatestValidHash *common.Hash   `json:"latestValidHash"` // 最新有效哈希
	ValidationError *string        `json:"validationError"` // 验证错误
}

// TransitionConfigurationV1 定义了过渡配置
type TransitionConfigurationV1 struct {
	TerminalTotalDifficulty *hexutil.Big   `json:"terminalTotalDifficulty"` // 终端总难度
	TerminalBlockHash       common.Hash    `json:"terminalBlockHash"`       // 终端区块哈希
	TerminalBlockNumber     hexutil.Uint64 `json:"terminalBlockNumber"`     // 终端区块号
}

// PayloadID 是payload构建过程的标识符
type PayloadID [8]byte

// Version 返回与标识符关联的payload版本
func (b PayloadID) Version() PayloadVersion {
	return PayloadVersion(b[0])
}

// Is 检查标识符是否匹配提供的任何payload版本
func (b PayloadID) Is(versions ...PayloadVersion) bool {
	return slices.Contains(versions, b.Version())
}

func (b PayloadID) String() string {
	return hexutil.Encode(b[:])
}

func (b PayloadID) MarshalText() ([]byte, error) {
	return hexutil.Bytes(b[:]).MarshalText()
}

func (b *PayloadID) UnmarshalText(input []byte) error {
	err := hexutil.UnmarshalFixedText("PayloadID", input, b[:])
	if err != nil {
		return fmt.Errorf("无效的payload id %q: %w", input, err)
	}
	return nil
}

type ForkChoiceResponse struct {
	PayloadStatus PayloadStatusV1 `json:"payloadStatus"`
	PayloadID     *PayloadID      `json:"payloadId"`
}

type ForkchoiceStateV1 struct {
	HeadBlockHash      common.Hash `json:"headBlockHash"`
	SafeBlockHash      common.Hash `json:"safeBlockHash"`
	FinalizedBlockHash common.Hash `json:"finalizedBlockHash"`
}

func encodeTransactions(txs []*types.Transaction) [][]byte {
	var enc = make([][]byte, len(txs))
	for i, tx := range txs {
		enc[i], _ = tx.MarshalBinary()
	}
	return enc
}

func decodeTransactions(enc [][]byte) ([]*types.Transaction, error) {
	var txs = make([]*types.Transaction, len(enc))
	for i, encTx := range enc {
		var tx types.Transaction
		if err := tx.UnmarshalBinary(encTx); err != nil {
			return nil, fmt.Errorf("invalid transaction %d: %v", i, err)
		}
		txs[i] = &tx
	}
	return txs, nil
}

// ExecutableDataToBlock constructs a block from executable data.
// It verifies that the following fields:
//
//		len(extraData) <= 32
//		uncleHash = emptyUncleHash
//		difficulty = 0
//	 	if versionedHashes != nil, versionedHashes match to blob transactions
//
// and that the blockhash of the constructed block matches the parameters. Nil
// Withdrawals value will propagate through the returned block. Empty
// Withdrawals value must be passed via non-nil, length 0 value in data.
func ExecutableDataToBlock(data ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, requests [][]byte) (*types.Block, error) {
	block, err := ExecutableDataToBlockNoHash(data, versionedHashes, beaconRoot, requests)
	if err != nil {
		return nil, err
	}
	if block.Hash() != data.BlockHash {
		return nil, fmt.Errorf("blockhash mismatch, want %x, got %x", data.BlockHash, block.Hash())
	}
	return block, nil
}

// ExecutableDataToBlockNoHash is analogous to ExecutableDataToBlock, but is used
// for stateless execution, so it skips checking if the executable data hashes to
// the requested hash (stateless has to *compute* the root hash, it's not given).
func ExecutableDataToBlockNoHash(data ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, requests [][]byte) (*types.Block, error) {
	txs, err := decodeTransactions(data.Transactions)
	if err != nil {
		return nil, err
	}
	if len(data.ExtraData) > int(params.MaximumExtraDataSize) {
		return nil, fmt.Errorf("invalid extradata length: %v", len(data.ExtraData))
	}
	if len(data.LogsBloom) != 256 {
		return nil, fmt.Errorf("invalid logsBloom length: %v", len(data.LogsBloom))
	}
	// Check that baseFeePerGas is not negative or too big
	if data.BaseFeePerGas != nil && (data.BaseFeePerGas.Sign() == -1 || data.BaseFeePerGas.BitLen() > 256) {
		return nil, fmt.Errorf("invalid baseFeePerGas: %v", data.BaseFeePerGas)
	}
	var blobHashes = make([]common.Hash, 0, len(txs))
	for _, tx := range txs {
		blobHashes = append(blobHashes, tx.BlobHashes()...)
	}
	if len(blobHashes) != len(versionedHashes) {
		return nil, fmt.Errorf("invalid number of versionedHashes: %v blobHashes: %v", versionedHashes, blobHashes)
	}
	for i := 0; i < len(blobHashes); i++ {
		if blobHashes[i] != versionedHashes[i] {
			return nil, fmt.Errorf("invalid versionedHash at %v: %v blobHashes: %v", i, versionedHashes, blobHashes)
		}
	}
	// Only set withdrawalsRoot if it is non-nil. This allows CLs to use
	// ExecutableData before withdrawals are enabled by marshaling
	// Withdrawals as the json null value.
	var withdrawalsRoot *common.Hash
	if data.Withdrawals != nil {
		h := types.DeriveSha(types.Withdrawals(data.Withdrawals), trie.NewStackTrie(nil))
		withdrawalsRoot = &h
	}

	var requestsHash *common.Hash
	if requests != nil {
		h := types.CalcRequestsHash(requests)
		requestsHash = &h
	}

	header := &types.Header{
		ParentHash:       data.ParentHash,
		UncleHash:        types.EmptyUncleHash,
		Coinbase:         data.FeeRecipient,
		Root:             data.StateRoot,
		TxHash:           types.DeriveSha(types.Transactions(txs), trie.NewStackTrie(nil)),
		ReceiptHash:      data.ReceiptsRoot,
		Bloom:            types.BytesToBloom(data.LogsBloom),
		Difficulty:       common.Big0,
		Number:           new(big.Int).SetUint64(data.Number),
		GasLimit:         data.GasLimit,
		GasUsed:          data.GasUsed,
		Time:             data.Timestamp,
		BaseFee:          data.BaseFeePerGas,
		Extra:            data.ExtraData,
		MixDigest:        data.Random,
		WithdrawalsHash:  withdrawalsRoot,
		ExcessBlobGas:    data.ExcessBlobGas,
		BlobGasUsed:      data.BlobGasUsed,
		ParentBeaconRoot: beaconRoot,
		RequestsHash:     requestsHash,
	}
	return types.NewBlockWithHeader(header).
			WithBody(types.Body{Transactions: txs, Uncles: nil, Withdrawals: data.Withdrawals}).
			WithWitness(data.ExecutionWitness),
		nil
}

// BlockToExecutableData constructs the ExecutableData structure by filling the
// fields from the given block. It assumes the given block is post-merge block.
func BlockToExecutableData(block *types.Block, fees *big.Int, sidecars []*types.BlobTxSidecar, requests [][]byte) *ExecutionPayloadEnvelope {
	data := &ExecutableData{
		BlockHash:        block.Hash(),
		ParentHash:       block.ParentHash(),
		FeeRecipient:     block.Coinbase(),
		StateRoot:        block.Root(),
		Number:           block.NumberU64(),
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		BaseFeePerGas:    block.BaseFee(),
		Timestamp:        block.Time(),
		ReceiptsRoot:     block.ReceiptHash(),
		LogsBloom:        block.Bloom().Bytes(),
		Transactions:     encodeTransactions(block.Transactions()),
		Random:           block.MixDigest(),
		ExtraData:        block.Extra(),
		Withdrawals:      block.Withdrawals(),
		BlobGasUsed:      block.BlobGasUsed(),
		ExcessBlobGas:    block.ExcessBlobGas(),
		ExecutionWitness: block.ExecutionWitness(),
	}

	// Add blobs.
	bundle := BlobsBundleV1{
		Commitments: make([]hexutil.Bytes, 0),
		Blobs:       make([]hexutil.Bytes, 0),
		Proofs:      make([]hexutil.Bytes, 0),
	}
	for _, sidecar := range sidecars {
		for j := range sidecar.Blobs {
			bundle.Blobs = append(bundle.Blobs, hexutil.Bytes(sidecar.Blobs[j][:]))
			bundle.Commitments = append(bundle.Commitments, hexutil.Bytes(sidecar.Commitments[j][:]))
			bundle.Proofs = append(bundle.Proofs, hexutil.Bytes(sidecar.Proofs[j][:]))
		}
	}

	return &ExecutionPayloadEnvelope{
		ExecutionPayload: data,
		BlockValue:       fees,
		BlobsBundle:      &bundle,
		Requests:         requests,
		Override:         false,
	}
}

// ExecutionPayloadBody is used in the response to GetPayloadBodiesByHash and GetPayloadBodiesByRange
type ExecutionPayloadBody struct {
	TransactionData []hexutil.Bytes     `json:"transactions"`
	Withdrawals     []*types.Withdrawal `json:"withdrawals"`
}

// Client identifiers to support ClientVersionV1.
const (
	ClientCode = "GE"
	ClientName = "go-ethereum"
)

// ClientVersionV1 contains information which identifies a client implementation.
type ClientVersionV1 struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func (v *ClientVersionV1) String() string {
	return fmt.Sprintf("%s-%s-%s-%s", v.Code, v.Name, v.Version, v.Commit)
}
