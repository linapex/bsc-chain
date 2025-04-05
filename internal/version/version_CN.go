// Copyright 2022 The go-ethereum Authors
// 本文件是go-ethereum库的一部分。
//
// go-ethereum库是免费软件：您可以根据GNU较宽松通用公共许可证的条款
// 重新分发和/或修改它，该许可证由自由软件基金会发布，
// 许可证版本3或（由您选择）任何更高版本。
//
// go-ethereum库的分发希望它会有用，
// 但不提供任何保证；甚至没有对适销性或特定用途适用性的暗示保证。
// 有关更多详细信息，请参阅GNU较宽松通用公共许可证。
//
// 您应该已经收到了GNU较宽松通用公共许可证的副本
// 与go-ethereum库一起。如果没有，请参阅<http://www.gnu.org/licenses/>。

// 版本包实现构建版本信息的读取。
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/ethereum/go-ethereum/version"
)

const ourPath = "github.com/ethereum/go-ethereum" // 我们模块的路径

// Family保存主要版本.次要版本的文本版本字符串
var Family = fmt.Sprintf("%d.%d", version.Major, version.Minor)

// Semantic保存主要版本.次要版本.补丁版本的文本版本字符串。
var Semantic = fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)

// WithMeta保存包括元数据在内的文本版本字符串。
var WithMeta = func() string {
	v := Semantic
	if version.Meta != "" {
		v += "-" + version.Meta
	}
	return v
}()

func WithCommit(gitCommit, gitDate string) string {
	vsn := WithMeta
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	if (version.Meta != "stable") && (gitDate != "") {
		vsn += "-" + gitDate
	}
	return vsn
}

// Archive保存用于Geth归档的文本版本字符串。例如
// 对于稳定版本是"1.8.11-dea1ce05"，对于不稳定
// 版本是"1.8.13-unstable-21c059b6"。
func Archive(gitCommit string) string {
	vsn := Semantic
	if version.Meta != "stable" {
		vsn += "-" + version.Meta
	}
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	return vsn
}

// ClientName根据以太坊p2p网络中的常见
// 约定创建软件名称/版本标识符。
func ClientName(clientIdentifier string) string {
	git, _ := VCS()
	return fmt.Sprintf("%s/v%v/%v-%v/%v",
		strings.Title(clientIdentifier),
		WithCommit(git.Commit, git.Date),
		runtime.GOOS, runtime.GOARCH,
		runtime.Version(),
	)
}

// Info返回关于当前二进制文件的构建和平台信息。
//
// 如果当前执行的包是由我们的go-ethereum
// 模块路径前缀的，它将打印出提交和日期VCS信息。否则，
// 它将假设它是由第三方导入的，并将返回导入的
// 版本以及它是否被另一个模块替换。
func Info() (version, vcs string) {
	version = WithMeta
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return version, ""
	}
	version = versionInfo(buildInfo)
	if status, ok := VCS(); ok {
		modified := ""
		if status.Dirty {
			modified = " (dirty)"
		}
		commit := status.Commit
		if len(commit) > 8 {
			commit = commit[:8]
		}
		vcs = commit + "-" + status.Date + modified
	}
	return version, vcs
}

// versionInfo返回当前执行的
// 实现的版本信息。
//
// 根据代码的实例化方式，它返回不同数量的
// 信息。如果它无法确定哪个模块与我们的
// 包相关，它会回退到params包中的硬编码值。
func versionInfo(info *debug.BuildInfo) string {
	// 如果主包来自我们的仓库，版本前缀为"geth"。
	if strings.HasPrefix(info.Path, ourPath) {
		return fmt.Sprintf("geth %s", info.Main.Version)
	}
	// 不是我们的主包，所以明确打印出模块路径和
	// 版本。
	var version string
	if info.Main.Path != "" && info.Main.Version != "" {
		// 这些在使用"go run"调用时可能为空。
		version = fmt.Sprintf("%s@%s ", info.Main.Path, info.Main.Version)
	}
	mod := findModule(info, ourPath)
	if mod == nil {
		// 如果我们的模块路径没有被导入，不清楚他们
		// 运行的是我们代码的哪个版本。回退到硬编码
		// 版本。
		return version + fmt.Sprintf("geth %s", WithMeta)
	}
	// 我们的包是主模块的依赖项。返回路径和
	// 两者的版本数据。
	version += fmt.Sprintf("%s@%s", mod.Path, mod.Version)
	if mod.Replace != nil {
		// 如果我们的包被其他东西替换，也注明这一点。
		version += fmt.Sprintf(" (replaced by %s@%s)", mod.Replace.Path, mod.Replace.Version)
	}
	return version
}

// findModule返回路径处的模块。
func findModule(info *debug.BuildInfo, path string) *debug.Module {
	if info.Path == ourPath {
		return &info.Main
	}
	for _, mod := range info.Deps {
		if mod.Path == path {
			return mod
		}
	}
	return nil
}