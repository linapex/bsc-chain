// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
	"github.com/ethereum/go-ethereum/params"
)

// BenchmarkInsertChain_empty_memdb 基准测试插入链条到空的内存数据库
func BenchmarkInsertChain_empty_memdb(b *testing.B) {
	benchInsertChain(b, false, nil)
}
// BenchmarkInsertChain_empty_diskdb 基准测试插入链条到空的磁盘数据库
func BenchmarkInsertChain_empty_diskdb(b *testing.B) {
	benchInsertChain(b, true, nil)
}
// BenchmarkInsertChain_valueTx_memdb 基准测试插入带有交易的链条到内存数据库
func BenchmarkInsertChain_valueTx_memdb(b *testing.B) {
	benchInsertChain(b, false, genValueTx(0))
}
// BenchmarkInsertChain_valueTx_diskdb 基准测试插入带有交易的链条到磁盘数据库
func BenchmarkInsertChain_valueTx_diskdb(b *testing.B) {
	benchInsertChain(b, true, genValueTx(0))
}
// BenchmarkInsertChain_valueTx_100kB_memdb 基准测试插入带有100kB交易的链条到内存数据库
func BenchmarkInsertChain_valueTx_100kB_memdb(b *testing.B) {
	benchInsertChain(b, false, genValueTx(100*1024))
}
// BenchmarkInsertChain_valueTx_100kB_diskdb 基准测试插入带有100kB交易的链条到磁盘数据库
func BenchmarkInsertChain_valueTx_100kB_diskdb(b *testing.B) {
	benchInsertChain(b, true, genValueTx(100*1024))
}
// BenchmarkInsertChain_uncles_memdb 基准测试插入带有叔块的链条到内存数据库
func BenchmarkInsertChain_uncles_memdb(b *testing.B) {
	benchInsertChain(b, false, genUncles)
}
// BenchmarkInsertChain_uncles_diskdb 基准测试插入带有叔块的链条到磁盘数据库
func BenchmarkInsertChain_uncles_diskdb(b *testing.B) {
	benchInsertChain(b, true, genUncles)
}
// BenchmarkInsertChain_ring200_memdb 基准测试插入带有200个环形交易的链条到内存数据库
func BenchmarkInsertChain_ring200_memdb(b *testing.B) {
	benchInsertChain(b, false, genTxRing(200))
}
// BenchmarkInsertChain_ring200_diskdb 基准测试插入带有200个环形交易的链条到磁盘数据库
func BenchmarkInsertChain_ring200_diskdb(b *testing.B) {
	benchInsertChain(b, true, genTxRing(200))
}
// BenchmarkInsertChain_ring1000_memdb 基准测试插入带有1000个环形交易的链条到内存数据库
func BenchmarkInsertChain_ring1000_memdb(b *testing.B) {
	benchInsertChain(b, false, genTxRing(1000))
}
// BenchmarkInsertChain_ring1000_diskdb 基准测试插入带有1000个环形交易的链条到磁盘数据库
func BenchmarkInsertChain_ring1000_diskdb(b *testing.B) {
	benchInsertChain(b, true, genTxRing(1000))
}

var (
	// This is the content of the genesis block used by the benchmarks.
	benchRootKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	benchRootAddr   = crypto.PubkeyToAddress(benchRootKey.PublicKey)
	benchRootFunds  = math.BigPow(2, 200)
)

// genValueTx returns a block generator that includes a single
// value-transfer transaction with n bytes of extra data in each
// block.
// genValueTx 返回一个块生成器，该生成器在每个块中包含一个带有n字节额外数据的值传输交易。
func genValueTx(nbytes int) func(int, *BlockGen) {
	// 我们可以重用所有交易的数据。
	// 在签名期间，方法 tx.WithSignature(s, sig)
	// 执行：
	// 	cpy := tx.inner.copy()
	//	cpy.setSignatureValues(signer.ChainID(), v, r, s)
	// 完成此操作后，调用者可以重用数据。
	data := make([]byte, nbytes)
	return func(i int, gen *BlockGen) {
		toaddr := common.Address{}
		gas, _ := IntrinsicGas(data, nil, nil, false, false, false, false)
		signer := gen.Signer()
		gasPrice := big.NewInt(0)
		if gen.header.BaseFee != nil {
			gasPrice = gen.header.BaseFee
		}
		tx, _ := types.SignNewTx(benchRootKey, signer, &types.LegacyTx{
			Nonce:    gen.TxNonce(benchRootAddr),
			To:       &toaddr,
			Value:    big.NewInt(1),
			Gas:      gas,
			Data:     data,
			GasPrice: gasPrice,
		})
		gen.AddTx(tx)
	}
}

var (
	ringKeys  = make([]*ecdsa.PrivateKey, 1000)
	ringAddrs = make([]common.Address, len(ringKeys))
)

func init() {
	ringKeys[0] = benchRootKey
	ringAddrs[0] = benchRootAddr
	for i := 1; i < len(ringKeys); i++ {
		ringKeys[i], _ = crypto.GenerateKey()
		ringAddrs[i] = crypto.PubkeyToAddress(ringKeys[i].PublicKey)
	}
}

// genTxRing returns a block generator that sends ether in a ring
// among n accounts. This is creates n entries in the state database
// and fills the blocks with many small transactions.
// genTxRing 返回一个块生成器，该生成器在n个账户之间以环形发送以太币。这会在状态数据库中创建n个条目，并用许多小交易填充块。
func genTxRing(naccounts int) func(int, *BlockGen) {
	from := 0
	availableFunds := new(big.Int).Set(benchRootFunds)
	return func(i int, gen *BlockGen) {
		block := gen.PrevBlock(i - 1)
		gas := block.GasLimit()
		gasPrice := big.NewInt(0)
		if gen.header.BaseFee != nil {
			gasPrice = gen.header.BaseFee
		}
		signer := gen.Signer()
		for {
			gas -= params.TxGas
			if gas < params.TxGas {
				break
			}
			to := (from + 1) % naccounts
			burn := new(big.Int).SetUint64(params.TxGas)
			burn.Mul(burn, gen.header.BaseFee)
			availableFunds.Sub(availableFunds, burn)
			if availableFunds.Cmp(big.NewInt(1)) < 0 {
				panic("not enough funds")
			}
			tx, err := types.SignNewTx(ringKeys[from], signer,
				&types.LegacyTx{
					Nonce:    gen.TxNonce(ringAddrs[from]),
					To:       &ringAddrs[to],
					Value:    availableFunds,
					Gas:      params.TxGas,
					GasPrice: gasPrice,
				})
			if err != nil {
				panic(err)
			}
			gen.AddTx(tx)
			from = to
		}
	}
}

// genUncles generates blocks with two uncle headers.
// genUncles 生成带有两个叔块头的块。
func genUncles(i int, gen *BlockGen) {
	if i >= 7 {
		b2 := gen.PrevBlock(i - 6).Header()
		b2.Extra = []byte("foo")
		gen.AddUncle(b2)
		b3 := gen.PrevBlock(i - 6).Header()
		b3.Extra = []byte("bar")
		gen.AddUncle(b3)
	}
}

// benchInsertChain 基准测试插入链条到数据库中。
func benchInsertChain(b *testing.B, disk bool, gen func(int, *BlockGen)) {
	// 创建内存中的数据库或临时目录中的数据库。
	var db ethdb.Database
	if !disk {
		db = rawdb.NewMemoryDatabase()
	} else {
		pdb, err := pebble.New(b.TempDir(), 128, 128, "", false)
		if err != nil {
			b.Fatalf("无法创建临时数据库: %v", err)
		}
		db = rawdb.NewDatabase(pdb)
		defer db.Close()
	}
	// 使用提供的块生成器函数生成 b.N 个块的链。
	gspec := &Genesis{
		Config: params.TestChainConfig,
		Alloc:  types.GenesisAlloc{benchRootAddr: {Balance: benchRootFunds}},
	}
	_, chain, _ := GenerateChainWithGenesis(gspec, ethash.NewFaker(), b.N, gen)

	// 计时插入新链。
	// 状态和块存储在同一个数据库中。
	chainman, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer chainman.Stop()
	b.ReportAllocs()
	b.ResetTimer()
	if i, err := chainman.InsertChain(chain); err != nil {
		b.Fatalf("插入错误 (块 %d): %v\n", i, err)
	}
}

func BenchmarkChainRead_header_10k(b *testing.B) {
	benchReadChain(b, false, 10000)
}
func BenchmarkChainRead_full_10k(b *testing.B) {
	benchReadChain(b, true, 10000)
}
func BenchmarkChainRead_header_100k(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping in short-mode")
	}
	benchReadChain(b, false, 100000)
}
func BenchmarkChainRead_full_100k(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping in short-mode")
	}
	benchReadChain(b, true, 100000)
}
func BenchmarkChainRead_header_500k(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping in short-mode")
	}
	benchReadChain(b, false, 500000)
}
func BenchmarkChainRead_full_500k(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping in short-mode")
	}
	benchReadChain(b, true, 500000)
}
func BenchmarkChainWrite_header_10k(b *testing.B) {
	benchWriteChain(b, false, 10000)
}
func BenchmarkChainWrite_full_10k(b *testing.B) {
	benchWriteChain(b, true, 10000)
}
func BenchmarkChainWrite_header_100k(b *testing.B) {
	benchWriteChain(b, false, 100000)
}
func BenchmarkChainWrite_full_100k(b *testing.B) {
	benchWriteChain(b, true, 100000)
}
func BenchmarkChainWrite_header_500k(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping in short-mode")
	}
	benchWriteChain(b, false, 500000)
}
func BenchmarkChainWrite_full_500k(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping in short-mode")
	}
	benchWriteChain(b, true, 500000)
}

// makeChainForBench writes a given number of headers or empty blocks/receipts
// into a database.
func makeChainForBench(db ethdb.Database, genesis *Genesis, full bool, count uint64) {
	var hash common.Hash
	for n := uint64(0); n < count; n++ {
		header := &types.Header{
			Coinbase:    common.Address{},
			Number:      big.NewInt(int64(n)),
			ParentHash:  hash,
			Difficulty:  big.NewInt(1),
			UncleHash:   types.EmptyUncleHash,
			TxHash:      types.EmptyTxsHash,
			ReceiptHash: types.EmptyReceiptsHash,
		}
		if n == 0 {
			header = genesis.ToBlock().Header()
		}
		hash = header.Hash()

		rawdb.WriteHeader(db, header)
		rawdb.WriteCanonicalHash(db, hash, n)
		rawdb.WriteTd(db, hash, n, big.NewInt(int64(n+1)))

		if n == 0 {
			rawdb.WriteChainConfig(db, hash, genesis.Config)
		}
		rawdb.WriteHeadHeaderHash(db, hash)

		if full || n == 0 {
			block := types.NewBlockWithHeader(header)
			rawdb.WriteBody(db, hash, n, block.Body())
			rawdb.WriteReceipts(db, hash, n, nil)
			rawdb.WriteHeadBlockHash(db, hash)
		}
	}
}

func benchWriteChain(b *testing.B, full bool, count uint64) {
	genesis := &Genesis{Config: params.AllEthashProtocolChanges}
	for i := 0; i < b.N; i++ {
		pdb, err := pebble.New(b.TempDir(), 1024, 128, "", false)
		if err != nil {
			b.Fatalf("error opening database: %v", err)
		}
		db := rawdb.NewDatabase(pdb)
		makeChainForBench(db, genesis, full, count)
		db.Close()
	}
}

func benchReadChain(b *testing.B, full bool, count uint64) {
	dir := b.TempDir()

	pdb, err := pebble.New(dir, 1024, 128, "", false)
	if err != nil {
		b.Fatalf("error opening database: %v", err)
	}
	db := rawdb.NewDatabase(pdb)

	genesis := &Genesis{Config: params.AllEthashProtocolChanges}
	makeChainForBench(db, genesis, full, count)
	db.Close()
	cacheConfig := *defaultCacheConfig
	cacheConfig.TrieDirtyDisabled = true

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pdb, err = pebble.New(dir, 1024, 128, "", false)
		if err != nil {
			b.Fatalf("error opening database: %v", err)
		}
		db = rawdb.NewDatabase(pdb)

		chain, err := NewBlockChain(db, &cacheConfig, genesis, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
		if err != nil {
			b.Fatalf("error creating chain: %v", err)
		}
		for n := uint64(0); n < count; n++ {
			header := chain.GetHeaderByNumber(n)
			if full {
				hash := header.Hash()
				rawdb.ReadBody(db, hash, n)
				rawdb.ReadReceipts(db, hash, n, header.Time, chain.Config())
			}
		}
		chain.Stop()
		db.Close()
	}
}
