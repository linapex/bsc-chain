// 版权 2015 The go-ethereum Authors
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

package common

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadJSON 读取给定文件并解析其JSON内容。
func LoadJSON(file string, val interface{}) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, val); err != nil {
		if syntaxerr, ok := err.(*json.SyntaxError); ok {
			line := findLine(content, syntaxerr.Offset)
			return fmt.Errorf("JSON syntax error at %v:%v: %v", file, line, err)
		}
		return fmt.Errorf("JSON unmarshal error in %v: %v", file, err)
	}
	return nil
}

// findLine 返回数据中给定偏移量对应的行号。
func findLine(data []byte, offset int64) (line int) {
	line = 1
	for i, r := range string(data) {
		if int64(i) >= offset {
			return
		}
		if r == '\n' {
			line++
		}
	}
	return
}
