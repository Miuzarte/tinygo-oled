# tinygo-oled

0.96寸OLED12864驱动 for tinygo, 移植修改自江协科技。

---

以下内容由 GitHub Copilot 补充, 部分经人工修改

---

## 功能概述

tinygo-oled 提供了一系列函数，用于控制 0.96 寸 OLED 显示屏。以下是各函数的简要介绍：

### 初始化与配置

- `Init(sclPin, sdaPin machine.Pin)`
  初始化 OLED 显示屏。如果 `sclPin` 和 `sdaPin` 为 0，则使用默认引脚 `machine.PB8` 和 `machine.PB9`。

### 基础操作

- `Clear()`
  清屏，清除 OLED 上的所有内容。

- `SetCursor(y, x uint8)`
  设置光标位置，`y` 为行坐标，`x` 为列坐标。

- `WriteCommand(command byte)`
  发送命令到 OLED。

- `WriteData(data byte)`
  发送数据到 OLED。

### 显示内容

除了 `ShowBinNum` 都会返回输出的字符数(不包括换行符)

传入 `ShowChar` 的 `column` 值越界时会自动换行, `char` 值越界(为非可打印字符)时会显示原始 uint8 值

- `ShowChar(line, column uint8, char byte) uint8`
  显示一个字符，`line` 为行位置，`column` 为列位置，`char` 为要显示的字符。

- `ShowString(line, column uint8, str string) uint8`
  显示字符串，支持换行符。`line` 和 `column` 分别为起始行和列位置。

- `ShowBool(line, column uint8, b bool) uint8`
  显示布尔值，`b` 为 `true` 时显示 `T`，为 `false` 时显示 `F`。

- `ShowNum(line, column uint8, number uint32, length uint8) uint8`
  显示无符号十进制数，`length` 为显示数字的长度，0 表示自动。

- `ShowNumNoZero(line, column uint8, number uint32, length uint8) uint8`
  显示无符号十进制数，隐藏多余的 0。

- `ShowSignedNum(line, column uint8, number int32, length uint8) uint8`
  显示有符号十进制数，`number` 为要显示的数字。

- `ShowHexNum(line, column uint8, number uint32, length uint8) uint8`
  显示十六进制数。

- `ShowBinNum(line, column byte, number uint32, length byte)`
  显示无符号二进制数。

### 清除内容

- `ClearLine(line, columnStart, columnEnd uint8)`
  清空指定行的内容，`columnStart` 和 `columnEnd` 为起始和结束列位置。

- `ClearColumn(column, lineStart, lineEnd uint8)`
  清空指定列的内容，`lineStart` 和 `lineEnd` 为起始和结束行位置。

### 工具函数

- `Pow(x, y uint32) (result uint32)`
  返回 `x` 的 `y` 次方。

- `GetNumLen(number, base uint32) (length uint8)`
  计算给定进制下显示给定数字需要多少位。

## 使用示例

```go
package main

import (
    "github.com/Miuzarte/tinygo-oled/oled"
    "machine"
)

func main() {
    oled.Init(machine.PB8, machine.PB9)
    oled.ShowString(0, 0, "Hello, OLED!")
}
```

## REF

C version: [Miuzarte/OLED12864](https://github.com/Miuzarte/OLED12864)

本项目移植修改自 [江协科技/STM32入门教程资料](https://jiangxiekeji.com/download.html#32)

本仓库基于 MIT 许可证
