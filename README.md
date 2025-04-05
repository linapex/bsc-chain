# BNB智能链

BNB智能链的目标是为BNB信标链带来可编程性和互操作性。为了拥抱现有的流行社区和先进技术，它将通过与以太坊上所有现有智能合约和以太坊工具保持兼容性来带来巨大的好处。为了实现这一目标，最简单的解决方案是基于go-ethereum进行开发，因为我们非常尊重以太坊的伟大工作。

BNB智能链基于go-ethereum分支开始其开发。因此，您可能会看到许多工具、二进制文件和文档都基于以太坊的，例如名称"geth"。

[![API参考](https://pkg.go.dev/badge/github.com/ethereum/go-ethereum)](https://pkg.go.dev/github.com/ethereum/go-ethereum?tab=doc)
[![Discord](https://img.shields.io/badge/discord-join%20chat-blue.svg)](https://discord.gg/z2VpC455eU)

但在EVM兼容的基础上，BNB智能链引入了一个由21个验证者组成的系统，采用权益证明授权(PoSA)共识机制，可以支持短区块时间和更低的费用。质押最多的验证者候选人将成为验证者并产生区块。双签检测和其他惩罚逻辑保证了安全性、稳定性和链的最终性。

## 系统架构

BNB智能链由以下核心组件构成：

- **核心区块链引擎**：负责处理交易、区块生产和状态管理
- **共识层**：实现PoSA共识机制，确保网络安全和交易最终性
- **P2P网络层**：管理节点之间的通信，维护网络连接
- **RPC接口**：提供标准化API，便于开发者与区块链交互
- **智能合约虚拟机**：执行EVM兼容的智能合约，支持丰富的DApp生态

## 节点类型

网络中包含三种类型的节点：

- **验证者节点**：21个活跃节点，负责生产区块和验证交易
- **全节点**：存储完整区块链数据，可独立验证所有交易
- **轻客户端**：只下载区块头，适用于资源受限的设备

## 产品定位

**BNB智能链**将是：

- **自主主权区块链**：通过选举的验证者提供安全和保障。
- **EVM兼容**：支持所有现有的以太坊工具，同时具有更快的最终性和更低的交易费用。
- **分布式链上治理**：权益证明授权带来去中心化和社区参与。作为原生代币，BNB将同时作为智能合约执行的燃料和质押的代币。

更多详情请参阅[白皮书](https://github.com/bnb-chain/whitepaper/blob/master/WHITEPAPER.md)。

## 主要特点

### 权益证明授权
尽管工作量证明(PoW)已被证明是实现去中心化网络的实用机制，但它对环境不友好，并且还需要大量参与者来维护安全性。

权威证明(PoA)提供了对51%攻击的一些防御，提高了效率并容忍一定程度的拜占庭节点(恶意或被黑客攻击的节点)。
同时，PoA协议最受批评的是不如PoW去中心化，因为验证者(即轮流产生区块的节点)拥有所有权限，容易受到腐败和安全攻击。

其他区块链，如EOS和Cosmos，都引入了不同类型的委托权益证明(DPoS)，允许代币持有者投票并选举验证者集合。这增加了去中心化程度并有利于社区治理。

为了将DPoS和PoA结合用于共识，BNB智能链实现了一种名为Parlia的新型共识引擎，它：

1. 区块由有限的验证者集合产生。
2. 验证者以PoA方式轮流产生区块，类似于以太坊的Clique共识引擎。
3. 验证者集合基于BNB智能链上的质押治理进行选举和淘汰。
4. Parlia共识引擎将与一组[系统合约](https://docs.bnbchain.org/bnb-smart-chain/staking/overview/#system-contracts)交互，以实现活跃度惩罚、收益分配和验证者集合更新功能。

## 原生代币

BNB将在BNB智能链上运行，方式与ETH在以太坊上运行相同，因此它仍然是BSC的`原生代币`。这意味着，
BNB将用于：

1. 在BSC上部署或调用智能合约时支付`燃料费`

## 构建源码

以下许多内容与go-ethereum相同或相似。

有关先决条件和详细构建说明，请阅读[安装说明](https://geth.ethereum.org/docs/getting-started/installing-geth)。

构建`geth`需要Go(1.22版本或更高)和C编译器(GCC 5或更高)。您可以使用您喜欢的包管理器安装它们。安装依赖项后，运行

```shell
make geth
```

或者，要构建完整的工具套件：

```shell
make all
```

如果在使用自行构建的二进制文件运行节点时遇到以下错误：
```shell
Caught SIGILL in blst_cgo_init, consult <blst>/bindinds/go/README.md.
```
请尝试添加以下环境变量并重新构建：
```shell
export CGO_CFLAGS="-O -D__BLST_PORTABLE__" 
export CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
```

## 可执行文件

BSC项目附带了几个包装器/可执行文件，位于`cmd`目录中。

|  命令   | 描述                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| :--------: | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`geth`** | 主要BNB智能链客户端二进制文件。它是进入BSC网络(主网、测试网或私有网)的入口点，能够作为全节点(默认)、归档节点(保留所有历史状态)或轻节点(实时检索数据)运行。它具有与go-ethereum相同甚至更多的RPC和其他接口，可以被其他进程用作通过在HTTP、WebSocket和/或IPC传输之上公开的JSON RPC端点进入BSC网络的网关。`geth --help`和[CLI页面](https://geth.ethereum.org/docs/interface/command-line-options)提供了命令行选项。 |
|   `clef`   | 独立签名工具，可用作`geth`的后端签名者。                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
|  `devp2p`  | 用于在网络层与节点交互的工具，无需运行完整的区块链。                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|  `abigen`  | 源代码生成器，用于将以太坊合约定义转换为易于使用、编译时类型安全的Go包。它可以处理普通的[以太坊合约ABI](https://docs.soliditylang.org/en/develop/abi-spec.html)，如果合约字节码也可用，则功能会扩展。但是，它也接受Solidity源文件，使开发更加流畅。请参阅我们的[原生DApps](https://geth.ethereum.org/docs/dapp/native-bindings)页面了解详情。                                                                                               |
| `bootnode` | 我们的以太坊客户端实现的精简版本，仅参与网络节点发现协议，但不运行任何更高级别的应用协议。它可以用作轻量级引导节点，帮助在私有网络中查找对等节点。                                                                                                                                                                                                                                                                                                                                                                                                            |
|   `evm`    | EVM(以太坊虚拟机)的开发者实用版本，能够在可配置的环境和执行模式下运行字节码片段。其目的是允许对EVM操作码进行隔离、细粒度的调试(例如`evm --code 60ff60ff --debug run`)。                                                                                                                                                                                                                                                                                                                                                                                                            |
| `rlpdump`  | 开发者实用工具，用于将二进制RLP([递归长度前缀](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp))转储(以太坊协议在网络和共识方面使用的数据编码)转换为更友好的层次表示(例如`rlpdump --hex CE0183FFFFFFC4C304050583616263`)。                                                                                                                                                                                                                                                                                 |

## 运行`geth`

详细介绍所有可能的命令行标志超出了本文范围(请参阅我们的[CLI Wiki页面](https://geth.ethereum.org/docs/fundamentals/command-line-options))，但我们列举了一些常见的参数组合，以帮助您快速了解如何运行自己的`geth`实例。

### 硬件要求

在主网上运行全节点的硬件必须满足某些要求：
- 运行最新版本的Mac OS X、Linux或Windows的VPS。
- 重要：3 TB(2023年12月)的可用磁盘空间，固态硬盘(SSD)，gp3，8k IOPS，500 MB/S吞吐量，读取延迟<1ms。(如果节点以快照同步方式启动，则需要NVMe SSD)
- 16核CPU和64 GB内存(RAM)
- 建议在AWS上使用m5zn.6xlarge或r7iz.4xlarge实例类型，在Google云上使用c2-standard-16。
- 具有5 MB/S上传/下载速度的宽带互联网连接

测试网的要求：
- 运行最新版本的Mac OS X、Linux或Windows的VPS。
- 测试网需要500G存储空间。
- 4核CPU和16GB内存(RAM)。

### 运行全节点的步骤

#### 1. 下载预构建二进制文件
```shell
# Linux
wget $(curl -s https://api.github.com/repos/bnb-chain/bsc/releases/latest |grep browser_ |grep geth_linux |cut -d\" -f4)
mv geth_linux geth
chmod -v u+x geth

# MacOS
wget $(curl -s https://api.github.com/repos/bnb-chain/bsc/releases/latest |grep browser_ |grep geth_mac |cut -d\" -f4)
mv geth_macos geth
chmod -v u+x geth
```

#### 2. 下载配置文件
```shell
//== 主网
wget $(curl -s https://api.github.com/repos/bnb-chain/bsc/releases/latest |grep browser_ |grep mainnet |cut -d\" -f4)
unzip mainnet.zip

//== 测试网
wget $(curl -s https://api.github.com/repos/bnb-chain/bsc/releases/latest |grep browser_ |grep testnet |cut -d\" -f4)
unzip testnet.zip
```

#### 3. 下载快照
从[这里](https://github.com/bnb-chain/bsc-snapshots)下载最新的链数据快照。按照指南组织您的文件。

#### 4. 启动全节点
```shell
## 默认情况下将使用基于路径的存储方案运行，并启用内联状态修剪，保留最新90000个区块的历史状态。
./geth --config ./config.toml --datadir ./node  --cache 8000 --rpc.allow-unprotected-txs --history.transactions 0

## 如果您希望获得高性能并且不太关心状态一致性，建议使用`--tries-verify-mode none`运行全节点。
./geth --config ./config.toml --datadir ./node  --cache 8000 --rpc.allow-unprotected-txs --history.transactions 0 --tries-verify-mode none
```

#### 5. 监控节点状态

默认情况下，从**./node/bsc.log**监控日志。当节点开始同步时，应该能够看到以下输出：
```shell
t=2022-09-08T13:00:27+0000 lvl=info msg="Imported new chain segment"             blocks=1    txs=177   mgas=17.317   elapsed=31.131ms    mgasps=556.259  number=21,153,429 hash=0x42e6b54ba7106387f0650defc62c9ace3160b427702dab7bd1c5abb83a32d8db dirty="0.00 B"
t=2022-09-08T13:00:29+0000 lvl=info msg="Imported new chain segment"             blocks=1    txs=251   mgas=39.638   elapsed=68.827ms    mgasps=575.900  number=21,153,430 hash=0xa3397b273b31b013e43487689782f20c03f47525b4cd4107c1715af45a88796e dirty="0.00 B"
t=2022-09-08T13:00:33+0000 lvl=info msg="Imported new chain segment"             blocks=1    txs=197   mgas=19.364   elapsed=34.663ms    mgasps=558.632  number=21,153,431 hash=0x0c7872b698f28cb5c36a8a3e1e315b1d31bda6109b15467a9735a12380e2ad14 dirty="0.00 B"
```

#### 6. 与全节点交互
启动`geth`内置的交互式[JavaScript控制台](https://geth.ethereum.org/docs/interface/javascript-console)(通过尾部的`console`子命令)，通过它您可以使用[`web3`方法](https://web3js.readthedocs.io/en/)(注意：`geth`中捆绑的`web3`版本非常旧，与官方文档不同步)以及`geth`自己的[管理API](https://geth.ethereum.org/docs/rpc/server)进行交互。此工具是可选的，如果您不使用它，您始终可以使用`geth attach`附加到已经运行的`geth`实例。

#### 7. 更多

关于[运行节点](https://docs.bnbchain.org/bnb-smart-chain/developers/node_operators/full_node/)和[成为验证者](https://docs.bnbchain.org/bnb-smart-chain/validator/create-val/)的更多详情

*注意：尽管一些内部保护措施防止交易在主网和测试网之间交叉，但您应该始终为游戏和真实资金使用单独的账户。除非您手动移动账户，否则`geth`默认会正确分离这两个网络，不会在它们之间提供任何账户。*

### 配置

作为向`geth`二进制文件传递众多标志的替代方法，您也可以通过以下方式传递配置文件：

```shell
$ geth --config /path/to/your_config.toml
```

要了解文件应该是什么样子，您可以使用`dumpconfig`子命令导出现有配置：

```shell
$ geth --your-favourite-flags dumpconfig
```

### 以编程方式与`geth`节点交互

作为开发者，迟早您会希望通过自己的程序而不是通过控制台手动与`geth`和BSC网络交互。为了帮助实现这一点，`geth`内置了对基于JSON-RPC的API([标准API](https://ethereum.github.io/execution-apis/api-documentation/)和[`geth`特定API](https://geth.ethereum.org/docs/interacting-with-geth/rpc))的支持。这些可以通过HTTP、WebSockets和IPC(在基于UNIX的平台上是UNIX套接字，在Windows上是命名管道)公开。

IPC接口默认启用，并公开`geth`支持的所有API，而HTTP和WS接口需要手动启用，并且由于安全原因只公开API的一个子集。这些可以按照您的预期打开/关闭和配置。

基于HTTP的JSON-RPC API选项：

  * `--http` 启用HTTP-RPC服务器
  * `--http.addr` HTTP-RPC服务器监听接口(默认：`localhost`)
  * `--http.port` HTTP-RPC服务器监听端口(默认：`8545`)
  * `--http.api` 通过HTTP-RPC接口提供的API(默认：`eth,net,web3`)
  * `--http.corsdomain` 接受跨源请求的域的逗号分隔列表(浏览器强制)
  * `--ws` 启用WS-RPC服务器
  * `--ws.addr` WS-RPC服务器监听接口(默认：`localhost`)
  * `--ws.port` WS-RPC服务器监听端口(默认：`8546`)
  * `--ws.api` 通过WS-RPC接口提供的API(默认：`eth,net,web3`)
  * `--ws.origins` 接受WebSocket请求的来源
  * `--ipcdisable` 禁用IPC-RPC服务器
  * `--ipcpath` 数据目录内IPC套接字/管道的文件名(显式路径会转义它)

您需要使用自己的编程环境的功能(库、工具等)通过HTTP、WS或IPC连接到配置了上述标志的`geth`节点，并且您需要在所有传输上使用[JSON-RPC](https://www.jsonrpc.org/specification)。您可以为多个请求重用相同的连接！

**注意：在打开基于HTTP/WS的传输之前，请了解其安全影响！互联网上的黑客正在积极尝试破坏具有公开API的BSC节点！此外，所有浏览器标签都可以访问本地运行的Web服务器，因此恶意网页可能会尝试破坏本地可用的API！**

### 运行私有网络
- [BSC-Deploy](https://github.com/bnb-chain/node-deploy/)：用于设置BNB智能链的部署工具。

## 运行引导节点

引导节点是非常轻量级的节点，不在NAT后面，只运行发现协议。当您启动节点时，它应该记录您的enode，这是一个公共标识符，其他人可以用它连接到您的节点。

首先，引导节点需要一个密钥，可以使用以下命令创建，该命令将密钥保存到boot.key：

```
bootnode -genkey boot.key
```

然后可以使用此密钥生成引导节点，如下所示：

```
bootnode -nodekey boot.key -addr :30311 -network bsc
```

传递给-addr的端口选择是任意的。
引导节点命令将以下日志返回到终端，确认它正在运行：

```
enode://3063d1c9e1b824cfbb7c7b6abafa34faec6bb4e7e06941d218d760acdd7963b274278c5c3e63914bd6d1b58504c59ec5522c56f883baceb8538674b92da48a96@127.0.0.1:0?discport=30311
Note: you're using cmd/bootnode, a developer tool.
We recommend using a regular node as bootstrap node for production deployments.
INFO [08-21|11:11:30.687] New local node record                    seq=1,692,616,290,684 id=2c9af1742f8f85ce ip=<nil> udp=0 tcp=0
INFO [08-21|12:11:30.753] New local node record                    seq=1,692,616,290,685 id=2c9af1742f8f85ce ip=54.217.128.118 udp=30311 tcp=0
INFO [09-01|02:46:26.234] New local node record                    seq=1,692,616,290,686 id=2c9af1742f8f85ce ip=34.250.32.100  udp=30311 tcp=0
```

## 贡献

感谢您考虑帮助改进源代码！我们欢迎来自互联网上任何人的贡献，并对即使是最小的修复也表示感谢！

如果您想为bsc做贡献，请fork、修复、提交并发送拉取请求，供维护者审查并合并到主代码库中。如果您希望提交更复杂的更改，请先在[我们的discord频道](https://discord.gg/bnbchain)上与核心开发者联系，以确保这些更改符合项目的总体理念，或者获得一些早期反馈，这可以使您的工作更轻松，以及我们的审查和合并程序更快更简单。

请确保您的贡献遵守我们的编码准则：

 * 代码必须遵守官方Go[格式化](https://golang.org/doc/effective_go.html#formatting)准则(即使用[gofmt](https://golang.org/cmd/gofmt/))。
 * 代码必须按照官方Go[注释](https://golang.org/doc/effective_go.html#commentary)准则进行文档化。
 * 拉取请求需要基于`master`分支并针对其打开。
 * 提交消息应该以它们修改的包为前缀。
   * 例如，"eth, rpc: make trace configs optional"

请参阅[开发者指南](https://geth.ethereum.org/docs/developers/geth-developer/dev-guide)了解有关配置环境、管理项目依赖项和测试程序的更多详情。

## 许可证

bsc库(即`cmd`目录外的所有代码)采用[GNU较宽松通用公共许可证v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html)许可，也包含在我们的存储库中的`COPYING.LESSER`文件中。

bsc二进制文件(即`cmd`目录内的所有代码)采用[GNU通用公共许可证v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html)许可，也包含在我们的存储库中的`COPYING`文件中。