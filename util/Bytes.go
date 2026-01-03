package util

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"time"
)

type Bytes []byte

func (b Bytes) Bool() bool {
	if len(b) == 0 {
		return false
	}
	return b[0] != 0
}

func (b Bytes) Int() int {
	// 根据当前平台 int 的位数决定读取长度
	if intSize == 64 {
		if len(b) < 8 {
			return 0
		}
		var x int64
		binary.Read(bytes.NewReader(b[:8]), EndianOrder, &x)
		return int(x)
	} else {
		if len(b) < 4 {
			return 0
		}
		var x int32
		binary.Read(bytes.NewReader(b[:4]), EndianOrder, &x)
		return int(x)
	}
}

// 将[]byte转换为int8
func (b Bytes) Int8() int8 {
	if len(b) < 1 {
		return 0
	}
	return int8(b[0])
}

// 将[]byte转换为uint8
func (b Bytes) Uint8() uint8 {
	if len(b) < 1 {
		return 0
	}
	return b[0]
}

// 将[]byte转换为int16
func (b Bytes) Int16() int16 {
	if len(b) < 2 {
		return 0
	}
	var x int16
	binary.Read(bytes.NewReader(b[:2]), EndianOrder, &x)
	return x
}

// 将[]byte转换为uint16
func (b Bytes) Uint16() uint16 {
	if len(b) < 2 {
		return 0
	}
	var x uint16
	binary.Read(bytes.NewReader(b[:2]), EndianOrder, &x)
	return x
}

// 将[]byte转换为int32
func (b Bytes) Int32() int32 {
	if len(b) < 4 {
		return 0
	}
	var x int32
	binary.Read(bytes.NewReader(b[:4]), EndianOrder, &x)
	return x
}

// 将[]byte转换为uint32
func (b Bytes) Uint32() uint32 {
	if len(b) < 4 {
		return 0
	}
	var x uint32
	binary.Read(bytes.NewReader(b[:4]), EndianOrder, &x)
	return x
}

// 将[]byte转换为int64
func (b Bytes) Int64() int64 {
	if len(b) < 8 {
		return 0
	}
	var x int64
	binary.Read(bytes.NewReader(b[:8]), EndianOrder, &x)
	return x
}

// 将[]byte转换为uint64
func (b Bytes) Uint64() uint64 {
	if len(b) < 8 {
		return 0
	}
	var x uint64
	binary.Read(bytes.NewReader(b[:8]), EndianOrder, &x)
	return x
}

func (b Bytes) String() string {
	return string(b)
}

// 将[]byte转换为time.Time类型
// []byte是原来time.Time的Unix时间戳（秒或毫秒）或时间等转换的字符串
// 支持多种常见时间格式和Unix时间戳（秒或毫秒）
func (b Bytes) Time() time.Time {
	str := string(b)
	// 先尝试将[]byte转换为Unix时间戳（秒或毫秒）
	// 尝试解析为int64时间戳
	if timestamp, err := StrToInt64(str); err == nil {
		// 判断是秒还是毫秒：长度10位通常是秒，13位通常是毫秒
		if len(str) == 13 {
			// 毫秒时间戳
			return time.UnixMilli(timestamp)
		} else {
			// 秒时间戳
			return time.Unix(timestamp, 0)
		}
	}
	// 定义支持的时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"15:04:05",
		"2006-01-02 15:04:05.999",
		"2006-01-02T15:04:05.999",
	}

	// 尝试使用各种格式解析
	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return t
		}
	}

	// 如果所有格式都解析失败，返回零时间
	return time.Time{}
}

// 根据i的类型转换为对应的类型
// 只有字符串和整型可以进行排序，故除整型外，其他类型都返回i.(string)
func (b Bytes) ToAny(i any) any {
	sv := b.String()
	switch v := i.(type) {
	case int:
		return b.Int()
	case bool:
		return b.Bool()
	case int8:
		return b.Int8()
	case uint8:
		return b.Uint8()
	case int16:
		return b.Int16()
	case uint16:
		return b.Uint16()
	case int32:
		return b.Int32()
	case uint32:
		return b.Uint32()
	case int64:
		return b.Int64()
	case uint64:
		return b.Uint64()
	case string:
		return b.String()
	case time.Time:
		return b.Time()
	/*
		case time.Time:
				tm, _ := time.Parse(time.DateTime, sv)
				return time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond(), time.Local)

	*/
	case float32:
		f, _ := strconv.ParseFloat(sv, 32)
		return float32(f)
	case float64:
		f, _ := strconv.ParseFloat(sv, 64)
		return f
	case complex64:
		c, _ := strconv.ParseComplex(strconv.FormatComplex(complex128(v), 'f', -1, 64), 64)
		return complex64(c)
	case complex128:
		c, _ := strconv.ParseComplex(sv, 128)
		return c
	default: //默认使用json解析
		var v2 any
		err := json.Unmarshal(b, &v2)
		if err != nil {
			return sv
		}
		return v2
	}
}

// 对数据进行转义，左右括号进行转义，遇到括号时,重复写入两次,作为转义
func (b Bytes) Escape() []byte {
	//如果 b==SPLIT,则无需转义
	if bytes.Equal(b, []byte(SPLIT)) {
		return b
	}
	blen := len(b)
	var buffer bytes.Buffer
	split := SPLIT[0]
	for i := range blen {
		c := b[i]
		switch c {
		case split:
			buffer.Write([]byte{split, split}) // 遇到分隔符时，重复写入两次
		default:
			buffer.WriteByte(c)
		}
	}
	return buffer.Bytes()
}

// 反转义
func (b Bytes) UnEscape() []byte {
	blen := len(b)
	var buffer bytes.Buffer
	split := SPLIT[0]
	for i := 0; i < len(b); i++ {
		c := b[i]
		if i+1 < blen {
			switch c {
			case split:
				if b[i+1] == split { //连续2个分隔符，表示数据中有分隔符
					buffer.WriteByte(split) //只需添加1个分隔符到数据中
					i++                     //跳过下一个分隔符
					continue
				}
			}
		}
		buffer.WriteByte(c)
	}
	return buffer.Bytes()
}

// 合并多个
func (b Bytes) Jion(ib ...[]byte) []byte {
	for _, v := range ib {
		b = bytes.Join([][]byte{b, v}, []byte{})
	}
	return b //原本的 b值，并不能被改变，所有需要返回一个新的值
}

// 根据分隔符SPLIT进行分割，同时反转义
func (b Bytes) Split(sp ...byte) [][]byte {
	rs := [][]byte{}
	s := []byte{}
	var split byte
	if len(sp) > 0 {
		split = sp[0]
	} else {
		split = SPLIT[0]
	}

	blen := len(b)

	for i := 0; i < blen; i++ {
		c := b[i]
		if c == split {
			//是连续分隔符，则是转义，只需添加一个分隔符到当前切片进行还原,并跳过下一个分隔符
			if i+1 < blen && b[i+1] == split {
				s = append(s, c)
				i++
			} else {
				// 单个分隔符，分割切片
				if len(s) > 0 {
					rs = append(rs, s)
					s = []byte{}
				}
			}
		} else {
			// 普通字符，添加到当前切片
			s = append(s, c)
		}
	}

	// 添加最后一个切片（如果有内容）
	if len(s) > 0 {
		rs = append(rs, s)
	}

	return rs
}
