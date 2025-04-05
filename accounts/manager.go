// 版权所有 2017 The go-ethereum 作者
// 本文件是 go-ethereum 库的一部分。

// go-ethereum 库是免费的软件：您可以根据自由软件基金会发布的 GNU 较低版本通用公共许可证的条款，重新分发和/或修改它，版本 3 或
// （或任何更高版本）。

// go-ethereum 库分发的目的是希望它有用，
// 但没有任何保证；甚至没有隐含的
// 适销性或适合特定用途的保证。有关更多详细信息，请参阅
// GNU 较低版本通用公共许可证。

// 您应该已经收到与 go-ethereum 库一起的 GNU 较低版本通用公共许可证的副本。如果没有，请参阅 <http://www.gnu.org/licenses/>。

package accounts

import (
	"errors"
	"fmt"
)

// Manager 表示以太坊账户管理器。
type Manager struct {
	accounts []Account
}

// NewManager 创建一个新的账户管理器。
func NewManager() *Manager {
	return &Manager{
		accounts: []Account{},
	}
}

// AddAccount 添加一个新的账户到管理器。
func (m *Manager) AddAccount(account Account) error {
	for _, acc := range m.accounts {
		if acc.Address == account.Address {
			return errors.New("账户已存在")
		}
	}
	m.accounts = append(m.accounts, account)
	return nil
}

// RemoveAccount 从管理器中移除一个账户。
func (m *Manager) RemoveAccount(address string) error {
	for i, acc := range m.accounts {
		if acc.Address == address {
			m.accounts = append(m.accounts[:i], m.accounts[i+1:]...)
			return nil
		}
	}
	return errors.New("账户不存在")
}

// ListAccounts 列出所有管理器中的账户。
func (m *Manager) ListAccounts() []Account {
	return m.accounts
}

// FindAccount 根据地址查找账户。
func (m *Manager) FindAccount(address string) (*Account, error) {
	for _, acc := range m.accounts {
		if acc.Address == address {
			return &acc, nil
		}
	}
	return nil, errors.New("账户不存在")
}
