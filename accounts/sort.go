// 版权所有 2018 go-ethereum 作者
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

package accounts

// AccountsByURL 基于URL字段为[]Account实现sort.Interface接口。
type AccountsByURL []Account

func (a AccountsByURL) Len() int           { return len(a) }
func (a AccountsByURL) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AccountsByURL) Less(i, j int) bool { return a[i].URL.Cmp(a[j].URL) < 0 }

// WalletsByURL 基于URL字段为[]Wallet实现sort.Interface接口。
type WalletsByURL []Wallet

func (w WalletsByURL) Len() int           { return len(w) }
func (w WalletsByURL) Swap(i, j int)      { w[i], w[j] = w[j], w[i] }
func (w WalletsByURL) Less(i, j int) bool { return w[i].URL().Cmp(w[j].URL()) < 0 }
