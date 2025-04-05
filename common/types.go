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
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/sha3"
)

// 哈希和地址的字节长度。
const (
	// HashLength 是哈希的预期长度
	HashLength = 32
	// AddressLength 是地址的预期长度
	AddressLength = 20
)

var (
	hashT    = reflect.TypeOf(Hash{})
	addressT = reflect.TypeOf(Address{})

	// MaxAddress 表示可能的最大地址值
	MaxAddress = HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")

	// MaxHash 表示可能的最大哈希值
	MaxHash = HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
)

// Hash 表示任意数据的32字节Keccak256哈希
type Hash [HashLength]byte

// BytesToHash 将字节数组b转换为哈希
// 如果b的长度大于哈希长度，将从左侧截断
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// BigToHash 将大整数b的字节表示转换为哈希
// 如果b的长度大于哈希长度，将从左侧截断
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

// HexToHash 将十六进制字符串s转换为哈希
// 如果s的长度大于哈希长度，将从左侧截断
func HexToHash(s string) Hash { return BytesToHash(FromHex(s)) }

// Cmp 比较两个哈希值
func (h Hash) Cmp(other Hash) int {
	return bytes.Compare(h[:], other[:])
}

// Bytes 获取哈希的底层字节表示
func (h Hash) Bytes() []byte { return h[:] }

// Big 将哈希转换为大整数
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }

// Hex 将哈希转换为十六进制字符串
func (h Hash) Hex() string { return hexutil.Encode(h[:]) }

// TerminalString 实现了log.TerminalStringer接口，为日志记录时的控制台输出格式化字符串
func (h Hash) TerminalString() string {
	return fmt.Sprintf("%x..%x", h[:3], h[29:])
}

// String 实现了stringer接口，在完整日志记录到文件时也会被日志记录器使用
func (h Hash) String() string {
	return h.Hex()
}

// Format 实现了fmt.Formatter接口
// Hash支持%v、%s、%q、%x、%X和%d格式动词
func (h Hash) Format(s fmt.State, c rune) {
	hexb := make([]byte, 2+len(h)*2)
	copy(hexb, "0x")
	hex.Encode(hexb[2:], h[:])

	switch c {
	case 'x', 'X':
		if !s.Flag('#') {
			hexb = hexb[2:]
		}
		if c == 'X' {
			hexb = bytes.ToUpper(hexb)
		}
		fallthrough
	case 'v', 's':
		s.Write(hexb)
	case 'q':
		q := []byte{'"'}
		s.Write(q)
		s.Write(hexb)
		s.Write(q)
	case 'd':
		fmt.Fprint(s, ([len(h)]byte)(h))
	default:
		fmt.Fprintf(s, "%%!%c(hash=%x)", c, h)
	}
}

// UnmarshalText 解析十六进制格式的哈希
func (h *Hash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash", input, h[:])
}

// UnmarshalJSON 解析十六进制格式的哈希
func (h *Hash) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText 返回哈希h的十六进制表示
func (h Hash) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// SetBytes 将哈希设置为b的值
// 如果b的长度大于哈希长度，将从左侧截断
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// Generate 实现了testing/quick.Generator接口
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}

// Scan 实现了database/sql.Scanner接口
func (h *Hash) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into Hash", src)
	}
	if len(srcB) != HashLength {
		return fmt.Errorf("can't scan []byte of len %d into Hash, want %d", len(srcB), HashLength)
	}
	copy(h[:], srcB)
	return nil
}

// Value 实现了database/sql/driver.Valuer接口
func (h Hash) Value() (driver.Value, error) {
	return h[:], nil
}

// ImplementsGraphQLType 如果Hash实现了指定的GraphQL类型则返回true
func (Hash) ImplementsGraphQLType(name string) bool { return name == "Bytes32" }

// UnmarshalGraphQL 解析提供的GraphQL查询数据
func (h *Hash) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		err = h.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Hash", input)
	}
	return err
}

// UnprefixedHash allows marshaling a Hash without 0x prefix.
type UnprefixedHash Hash

// UnmarshalText decodes the hash from hex. The 0x prefix is optional.
func (h *UnprefixedHash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedHash", input, h[:])
}

// MarshalText encodes the hash as hex.
func (h UnprefixedHash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

/////////// Address

// Address represents the 20 byte address of an Ethereum account.
type Address [AddressLength]byte

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// BigToAddress returns Address with byte values of b.
// If b is larger than len(h), b will be cropped from the left.
func BigToAddress(b *big.Int) Address { return BytesToAddress(b.Bytes()) }

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) Address { return BytesToAddress(FromHex(s)) }

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// Ethereum address or not.
func IsHexAddress(s string) bool {
	if has0xPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*AddressLength && isHex(s)
}

// Cmp compares two addresses.
func (a Address) Cmp(other Address) int {
	return bytes.Compare(a[:], other[:])
}

// Bytes gets the string representation of the underlying address.
func (a Address) Bytes() []byte { return a[:] }

// Big converts an address to a big integer.
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	return string(a.checksumHex())
}

// String implements fmt.Stringer.
func (a Address) String() string {
	return a.Hex()
}

func (a *Address) checksumHex() []byte {
	buf := a.hex()

	// compute checksum
	sha := sha3.NewLegacyKeccak256()
	sha.Write(buf[2:])
	hash := sha.Sum(nil)
	for i := 2; i < len(buf); i++ {
		hashByte := hash[(i-2)/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if buf[i] > '9' && hashByte > 7 {
			buf[i] -= 32
		}
	}
	return buf[:]
}

func (a Address) hex() []byte {
	var buf [len(a)*2 + 2]byte
	copy(buf[:2], "0x")
	hex.Encode(buf[2:], a[:])
	return buf[:]
}

// Format implements fmt.Formatter.
// Address supports the %v, %s, %q, %x, %X and %d format verbs.
func (a Address) Format(s fmt.State, c rune) {
	switch c {
	case 'v', 's':
		s.Write(a.checksumHex())
	case 'q':
		q := []byte{'"'}
		s.Write(q)
		s.Write(a.checksumHex())
		s.Write(q)
	case 'x', 'X':
		// %x disables the checksum.
		hex := a.hex()
		if !s.Flag('#') {
			hex = hex[2:]
		}
		if c == 'X' {
			hex = bytes.ToUpper(hex)
		}
		s.Write(hex)
	case 'd':
		fmt.Fprint(s, ([len(a)]byte)(a))
	default:
		fmt.Fprintf(s, "%%!%c(address=%x)", c, a)
	}
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a), b will be cropped from the left.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}

// UnmarshalText 解析十六进制格式的哈希
func (a *Address) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Address", input, a[:])
}

// UnmarshalJSON 解析十六进制格式的哈希
func (a *Address) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(addressT, input, a[:])
}

// Scan 实现了database/sql.Scanner接口
func (a *Address) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into Address", src)
	}
	if len(srcB) != AddressLength {
		return fmt.Errorf("can't scan []byte of len %d into Address, want %d", len(srcB), AddressLength)
	}
	copy(a[:], srcB)
	return nil
}

// Value 实现了database/sql/driver.Valuer接口
func (a Address) Value() (driver.Value, error) {
	return a[:], nil
}

// ImplementsGraphQLType 如果Hash实现了指定的GraphQL类型则返回true
func (a Address) ImplementsGraphQLType(name string) bool { return name == "Address" }

// UnmarshalGraphQL 解析提供的GraphQL查询数据
func (a *Address) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		err = a.UnmarshalText([]byte(input))
	default:
		err = fmt.Errorf("unexpected type %T for Address", input)
	}
	return err
}

// UnprefixedAddress allows marshaling an Address without 0x prefix.
type UnprefixedAddress Address

// UnmarshalText decodes the address from hex. The 0x prefix is optional.
func (a *UnprefixedAddress) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedAddress", input, a[:])
}

// MarshalText encodes the address as hex.
func (a UnprefixedAddress) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(a[:])), nil
}

// MixedcaseAddress retains the original string, which may or may not be
// correctly checksummed
type MixedcaseAddress struct {
	addr     Address
	original string
}

// NewMixedcaseAddress constructor (mainly for testing)
func NewMixedcaseAddress(addr Address) MixedcaseAddress {
	return MixedcaseAddress{addr: addr, original: addr.Hex()}
}

// NewMixedcaseAddressFromString is mainly meant for unit-testing
func NewMixedcaseAddressFromString(hexaddr string) (*MixedcaseAddress, error) {
	if !IsHexAddress(hexaddr) {
		return nil, errors.New("invalid address")
	}
	a := FromHex(hexaddr)
	return &MixedcaseAddress{addr: BytesToAddress(a), original: hexaddr}, nil
}

// UnmarshalJSON parses MixedcaseAddress
func (ma *MixedcaseAddress) UnmarshalJSON(input []byte) error {
	if err := hexutil.UnmarshalFixedJSON(addressT, input, ma.addr[:]); err != nil {
		return err
	}
	return json.Unmarshal(input, &ma.original)
}

// MarshalJSON marshals the original value
func (ma MixedcaseAddress) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(ma.original, "0x") || strings.HasPrefix(ma.original, "0X") {
		return json.Marshal(fmt.Sprintf("0x%s", ma.original[2:]))
	}
	return json.Marshal(fmt.Sprintf("0x%s", ma.original))
}

// Address returns the address
func (ma *MixedcaseAddress) Address() Address {
	return ma.addr
}

// String implements fmt.Stringer
func (ma *MixedcaseAddress) String() string {
	if ma.ValidChecksum() {
		return fmt.Sprintf("%s [chksum ok]", ma.original)
	}
	return fmt.Sprintf("%s [chksum INVALID]", ma.original)
}

// ValidChecksum returns true if the address has valid checksum
func (ma *MixedcaseAddress) ValidChecksum() bool {
	return ma.original == ma.addr.Hex()
}

// Original returns the mixed-case input string
func (ma *MixedcaseAddress) Original() string {
	return ma.original
}

// AddressEIP55 is an alias of Address with a customized json marshaller
type AddressEIP55 Address

// String returns the hex representation of the address in the manner of EIP55.
func (addr AddressEIP55) String() string {
	return Address(addr).Hex()
}

// MarshalJSON marshals the address in the manner of EIP55.
func (addr AddressEIP55) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

type Decimal uint64

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

// UnmarshalJSON 解析十六进制格式的哈希
func (d *Decimal) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return &json.UnmarshalTypeError{Value: "non-string", Type: reflect.TypeOf(uint64(0))}
	}
	if i, err := strconv.ParseUint(string(input[1:len(input)-1]), 10, 64); err == nil {
		*d = Decimal(i)
		return nil
	} else {
		return err
	}
}

type PrettyBytes []byte

// TerminalString 实现了log.TerminalStringer接口，为日志记录时的控制台输出格式化字符串
func (b PrettyBytes) TerminalString() string {
	if len(b) < 7 {
		return fmt.Sprintf("%x", b)
	}
	return fmt.Sprintf("%#x...%x (%dB)", b[:3], b[len(b)-3:], len(b))
}
