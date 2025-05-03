package oled

import (
	"machine"
	"time"
)

const (
	// 像素分辨率
	RES_HORIZONTAL = 128
	RES_VERTICAL   = 64

	// 字符长宽
	CHAR_WIDTH  = 8
	CHAR_HEIGHT = 16

	// 字符分辨率
	CHARS_HORIZONTAL = RES_HORIZONTAL / CHAR_WIDTH
	CHARS_VERTICAL   = RES_VERTICAL / CHAR_HEIGHT
)

var DisplayLines uint8 = 4 // 低于 4 的值会在下方留出保护区

const (
	ADDRESS       = 0x78 // OLED 地址
	WRITE_COMMAND = 0x00
	WRITE_DATA    = 0x40
	INIT_DELAY    = time.Millisecond * 50 // 上电延时
)

const (
	DEFAULT_SCL = machine.PB8
	DEFAULT_SDA = machine.PB9
)

var (
	SCL = DEFAULT_SCL
	SDA = DEFAULT_SDA
)

// w_SCL 写 [SCL] 引脚
//
// I2C 信号输出太快导致 OLED 不显示内容,
// 禁止内联且输出两次拖时间
//
//go:noinline
func w_SCL(b bool) {
	SCL.Set(b)
	SCL.Set(b)
}

// w_SDA 写 [SDA] 引脚
//
// I2C 信号输出太快导致 OLED 不显示内容,
// 禁止内联且输出两次拖时间
//
//go:noinline
func w_SDA(b bool) {
	SDA.Set(b)
	SDA.Set(b)
}

func i2cInit() {
	pc := machine.PinConfig{
		Mode: machine.PinOutput10MHz |
			machine.PinOutputModeGPOpenDrain,
	}
	SCL.Configure(pc)
	SDA.Configure(pc)

	w_SCL(true)
	w_SDA(true)
}

//go:noinline
func i2cStart() {
	w_SDA(true)
	w_SCL(true)
	w_SDA(false)
	w_SCL(false)
}

//go:noinline
func i2cStop() {
	w_SDA(false)
	w_SCL(true)
	w_SDA(true)
}

// i2cSendByte 发送一个字节
//
//go:noinline
func i2cSendByte(data byte) {
	for i := range 8 {
		w_SDA((data & (0x80 >> i)) != 0)
		w_SCL(true)
		w_SCL(false)
	}
	w_SCL(true) // 额外的一个时钟，不处理应答信号
	w_SCL(false)
}

// WriteCommand 写命令
//
//go:noinline
func WriteCommand(command byte) {
	i2cStart()
	i2cSendByte(ADDRESS)       // 从机地址
	i2cSendByte(WRITE_COMMAND) // 写命令
	i2cSendByte(command)
	i2cStop()
}

// WriteData 写数据
//
//go:noinline
func WriteData(data byte) {
	i2cStart()
	i2cSendByte(ADDRESS)    // 从机地址
	i2cSendByte(WRITE_DATA) // 写数据
	i2cSendByte(data)
	i2cStop()
}

// SetCursor 设置光标位置
//
//	y 以左上角为原点, 向下方向的坐标, 范围: 0 - 7
//	x 以左上角为原点, 向右方向的坐标, 范围: 0 - 127
func SetCursor(y, x uint8) {
	WriteCommand(0xB0 | y)                 // 设置 Y 位置
	WriteCommand(0x10 | ((x & 0xF0) >> 4)) // 设置 X 位置高 4 位
	WriteCommand(0x00 | (x & 0x0F))        // 设置 X 位置低 4 位
}

// Clear 清屏
func Clear() {
	for i := range uint8(8) {
		SetCursor(i, 0)
		for range RES_HORIZONTAL {
			WriteData(0x00)
		}
	}
}

// ClearLine 清空行
//
//	line 行位置, 范围: 0 - 3, 溢出归零
//	columnStart 起始列位置
//	columnEnd 结束列位置, 左闭右闭
func ClearLine(line, columnStart, columnEnd uint8) {
	if columnStart > columnEnd {
		columnStart, columnEnd = columnEnd, columnStart
	}
	columnEnd %= CHARS_HORIZONTAL
	columnEnd++ // 闭区间
	line %= CHARS_VERTICAL
	ops := CHAR_WIDTH * (columnEnd - columnStart)
	SetCursor(line*2, columnStart*CHAR_WIDTH)
	for range ops {
		WriteData(0x00)
	}
	SetCursor(line*2+1, columnStart*CHAR_WIDTH)
	for range ops {
		WriteData(0x00)
	}
}

// ClearColumn 清空列
//
//	column 列位置, 范围: 0 - 15, 溢出归零
//	lineStart 起始行位置
//	lineEnd 结束行位置, 左闭右闭
func ClearColumn(column, lineStart, lineEnd uint8) {
	if lineStart > lineEnd {
		lineStart, lineEnd = lineEnd, lineStart
	}
	lineEnd %= CHARS_VERTICAL
	lineEnd++
	lineEnd *= 2
	column %= CHARS_HORIZONTAL
	for ; lineStart <= lineEnd; lineStart++ {
		SetCursor(lineStart, column*CHAR_WIDTH)
		for range CHAR_WIDTH {
			WriteData(0x00)
		}
	}
}

// ShowChar 显示一个字符
//
//	line 行位置, 范围: 0 - 3, 溢出归零
//	column 列位置, 范围: 0 - 15, 溢出自动换行
//	char 要显示的字符
func ShowChar(line, column uint8, char byte) uint8 {
	if char < 32 || char > 126 { // 非可打印字符
		return ShowNumNoZero(line, column, uint32(char), 0)
	}
	overflow := column / CHARS_HORIZONTAL
	if overflow > 0 { // 换行
		line += overflow
		column -= overflow * CHARS_HORIZONTAL
	}
	line %= DisplayLines
	SetCursor(line*2, column*CHAR_WIDTH) // 设置光标位置在上半部分
	for i := range CHAR_WIDTH {
		WriteData(FONT[char-' '][i]) // 显示上半部分内容
	}
	SetCursor(line*2+1, column*CHAR_WIDTH) // 设置光标位置在下半部分
	for i := range CHAR_WIDTH {
		WriteData(FONT[char-' '][i+8]) // 显示下半部分内容
	}
	return 1
}

// ShowString 显示字符串
//
//	line 起始行位置, 范围: 0 - 3
//	column 起始列位置, 范围: 0 - 15
//	str 要显示的字符串, 支持换行符
func ShowString(line, column uint8, str string) (n uint8) {
	for i := range str {
		if str[i] == '\n' {
			j := (column + 15) &^ 15
			for ; column < j; column++ { // 清空换行位置
				/*n +=*/ ShowChar(line, column, ' ')
			}
			continue
		}
		n += ShowChar(line, column, str[i])
		column++
	}
	return n
}

// ShowString 显示布尔值
//
//	line 起始行位置, 范围: 0 - 3
//	column 起始列位置, 范围: 0 - 15
//	b 显示'T'/'F'
func ShowBool(line, column uint8, b bool) uint8 {
	if b {
		ShowChar(line, column, 'T')
	} else {
		ShowChar(line, column, 'F')
	}
	return 1
}

// Pow 返回 x 的 y 次方
func Pow(x, y uint32) (result uint32) {
	if x == 2 {
		return 1 << y
	}
	result = 1
	for range y {
		result *= x
	}
	return
}

// GetNumLen 计算给定进制下显示给定数字需要多少位
//
//	number 要计算的数字
//	base 进制
func GetNumLen(number, base uint32) (length uint8) {
	if number == 0 {
		return 1
	}
LOOP:
	length++
	number /= base
	if number > 0 {
		goto LOOP
	}
	return
}

// ShowNum 显示无符号十进制数
//
//	line 起始行位置, 范围: 0 - 3
//	column 起始列位置, 范围: 0 - 15
//	number 要显示的数字, 范围: 0 - 4294967295
//	length 要显示数字的长度, 范围: 1 - 10, 0: auto
func ShowNum(line, column uint8, number uint32, length uint8) (n uint8) {
	if length == 0 {
		length = GetNumLen(number, 10)
	}
	for length > 0 {
		n += ShowChar(line, column, byte(number/Pow(10, uint32(length-1))%10)+'0')
		column++
		length--
	}
	return
}

// ShowNum 显示无符号十进制数, 隐藏多余的 0
//
//	line 起始行位置, 范围: 0 - 3
//	column 起始列位置, 范围: 0 - 15
//	number 要显示的数字, 范围: 0 - 4294967295
//	length 要显示数字的长度, 范围: 1 - 10, 0: auto
func ShowNumNoZero(line, column uint8, number uint32, length uint8) (n uint8) {
	if number == 0 {
		return ShowChar(line, column+length-1, '0')
	}
	if length == 0 {
		length = GetNumLen(number, 10)
	}
	var char byte
	noZeroOnce := true
	for length > 0 {
		char = byte(number/Pow(10, uint32(length-1))%10) + '0'
		if noZeroOnce {
			if char == '0' {
				char = ' '
			} else {
				noZeroOnce = false
			}
		}
		n += ShowChar(line, column, char)
		column++
		length--
	}
	return
}

// ShowNum 显示有符号十进制数
//
//	line 起始行位置, 范围: 0 - 3
//	column 起始列位置, 范围: 0 - 15
//	number 要显示的数字, 范围: -2147483648 - 2147483647
//	length 要显示数字的长度, 范围: 1 - 10, 0: auto, 不包括正负号
func ShowSignedNum(line, column uint8, number int32, length uint8) (n uint8) {
	if number > 0 {
		n += ShowChar(line, column, '+')
	} else if number < 0 {
		n += ShowChar(line, column, '-')
		number = -number
	} // 0 隐藏正负号
	if length == 0 {
		length = GetNumLen(uint32(number), 10)
	}
	return n + ShowNum(line, column+1, uint32(number), length)
}

// ShowHexNum 显示十六进制数
//
//	line 起始行位置, 范围: 0 - 3
//	column 起始列位置, 范围: 0 - 15
//	number 要显示的数字, 范围: 0 - 0xFFFFFFFF
//	length 要显示数字的长度, 范围: 1 - 8, 0: auto
func ShowHexNum(line, column uint8, number uint32, length uint8) (n uint8) {
	if length == 0 {
		length = GetNumLen(number, 16)
	}
	var char, single byte
	for ; length > 0; length-- {
		single = byte(number / Pow(16, uint32(length-1)) % 16)
		if single < 10 {
			char = single + '0'
		} else {
			char = single - 10 + 'A'
		}
		n += ShowChar(line, column, char)
		column++
	}
	return n
}

// ShowHexNum 显示无符号二进制数
//
//	line 起始行位置, 范围: 0 - 3
//	column 起始列位置, 范围: 0 - 15
//	number 要显示的数字, 范围: 0 - 0xFFFFFFFF
//	length 要显示数字的长度, 范围: 1 - 32
func ShowBinNum(line, column byte, number uint32, length byte) {
	for ; length > 0; length-- {
		ShowChar(line, column, byte((number>>(length-1))&0x01)+'0')
		column++
	}
}

// Init 初始化 OLED,
// sclPin 和 sdaPin 传入 0 时使用默认引脚 [machine.PB8] [machine.PB9]
func Init(sclPin, sdaPin machine.Pin) {
	if sclPin != 0 || sdaPin != 0 {
		SCL = sclPin
		SDA = sdaPin
	}

	time.Sleep(INIT_DELAY) // 上电延时

	i2cInit() // 端口初始化

	WriteCommand(0xAE) // 关闭显示

	WriteCommand(0xD5) // 设置显示时钟分频比/振荡器频率
	WriteCommand(0x80)

	WriteCommand(0xA8) // 设置多路复用率
	WriteCommand(0x3F)

	WriteCommand(0xD3) // 设置显示偏移
	WriteCommand(0x00)

	WriteCommand(0x40) // 设置显示开始行

	WriteCommand(0xA1) // 设置左右方向, 0xA1正常 0xA0左右反置

	WriteCommand(0xC8) // 设置上下方向, 0xC8正常 0xC0上下反置

	WriteCommand(0xDA) // 设置COM引脚硬件配置
	WriteCommand(0x12)

	WriteCommand(0x81) // 设置对比度控制
	WriteCommand(0xCF)

	WriteCommand(0xD9) // 设置预充电周期
	WriteCommand(0xF1)

	WriteCommand(0xDB) // 设置VCOMH取消选择级别
	WriteCommand(0x30)

	WriteCommand(0xA4) // 设置整个显示打开/关闭

	WriteCommand(0xA6) // 设置正常/倒转显示

	WriteCommand(0x8D) // 设置充电泵
	WriteCommand(0x14)

	WriteCommand(0xAF) // 开启显示

	Clear() // OLED清屏
}
