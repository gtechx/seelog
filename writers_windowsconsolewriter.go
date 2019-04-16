// Copyright (c) 2012 - Cloud Instruments Co., Ltd.
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// +build windows

package seelog

import (
	"bytes"
	"io"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	_ int = iota
	DiscardNonColorEscSeq
	OutputNonColorEscSeq
)

// consoleWriter is used to write to console
type consoleWriter struct {
	w             io.Writer
	mode          int
	state         csiState
	paramStartBuf bytes.Buffer
	paramBuf      bytes.Buffer
}

// Creates a new console writer. Returns error, if the console writer couldn't be created.
func NewConsoleWriter() (writer *consoleWriter, err error) {
	newWriter := new(consoleWriter)
	newWriter.w = os.Stdout

	return newWriter, nil
}

// Create folder and file on WriteLog/Write first call
//func (console *consoleWriter) Write(bytes []byte) (int, error) {
//	return fmt.Print(string(bytes))
//}

func (console *consoleWriter) String() string {
	return "Windows Console writer"
}

type (
	csiState    int
	parseResult int
)

const (
	outsideCsiCode csiState = iota
	firstCsiCode
	secondCsiCode
)

const (
	noConsole parseResult = iota
	changedColor
	unknown
)

const (
	firstCsiChar   byte = '\x1b'
	secondeCsiChar byte = '['
	separatorChar  byte = ';'
	sgrCode        byte = 'm'
)

const (
	foregroundBlue      = uint16(0x0001)
	foregroundGreen     = uint16(0x0002)
	foregroundRed       = uint16(0x0004)
	foregroundIntensity = uint16(0x0008)
	backgroundBlue      = uint16(0x0010)
	backgroundGreen     = uint16(0x0020)
	backgroundRed       = uint16(0x0040)
	backgroundIntensity = uint16(0x0080)
	underscore          = uint16(0x8000)

	foregroundMask = foregroundBlue | foregroundGreen | foregroundRed | foregroundIntensity
	backgroundMask = backgroundBlue | backgroundGreen | backgroundRed | backgroundIntensity
)

const (
	ansiReset        = "0"
	ansiIntensityOn  = "1"
	ansiIntensityOff = "21"
	ansiUnderlineOn  = "4"
	ansiUnderlineOff = "24"
	ansiBlinkOn      = "5"
	ansiBlinkOff     = "25"

	ansiForegroundBlack   = "30"
	ansiForegroundRed     = "31"
	ansiForegroundGreen   = "32"
	ansiForegroundYellow  = "33"
	ansiForegroundBlue    = "34"
	ansiForegroundMagenta = "35"
	ansiForegroundCyan    = "36"
	ansiForegroundWhite   = "37"
	ansiForegroundDefault = "39"

	ansiBackgroundBlack   = "40"
	ansiBackgroundRed     = "41"
	ansiBackgroundGreen   = "42"
	ansiBackgroundYellow  = "43"
	ansiBackgroundBlue    = "44"
	ansiBackgroundMagenta = "45"
	ansiBackgroundCyan    = "46"
	ansiBackgroundWhite   = "47"
	ansiBackgroundDefault = "49"

	ansiLightForegroundGray    = "90"
	ansiLightForegroundRed     = "91"
	ansiLightForegroundGreen   = "92"
	ansiLightForegroundYellow  = "93"
	ansiLightForegroundBlue    = "94"
	ansiLightForegroundMagenta = "95"
	ansiLightForegroundCyan    = "96"
	ansiLightForegroundWhite   = "97"

	ansiLightBackgroundGray    = "100"
	ansiLightBackgroundRed     = "101"
	ansiLightBackgroundGreen   = "102"
	ansiLightBackgroundYellow  = "103"
	ansiLightBackgroundBlue    = "104"
	ansiLightBackgroundMagenta = "105"
	ansiLightBackgroundCyan    = "106"
	ansiLightBackgroundWhite   = "107"
)

type drawType int

const (
	foreground drawType = iota
	background
)

type winColor struct {
	code     uint16
	drawType drawType
}

var colorMap = map[string]winColor{
	ansiForegroundBlack:   {0, foreground},
	ansiForegroundRed:     {foregroundRed, foreground},
	ansiForegroundGreen:   {foregroundGreen, foreground},
	ansiForegroundYellow:  {foregroundRed | foregroundGreen, foreground},
	ansiForegroundBlue:    {foregroundBlue, foreground},
	ansiForegroundMagenta: {foregroundRed | foregroundBlue, foreground},
	ansiForegroundCyan:    {foregroundGreen | foregroundBlue, foreground},
	ansiForegroundWhite:   {foregroundRed | foregroundGreen | foregroundBlue, foreground},
	ansiForegroundDefault: {foregroundRed | foregroundGreen | foregroundBlue, foreground},

	ansiBackgroundBlack:   {0, background},
	ansiBackgroundRed:     {backgroundRed, background},
	ansiBackgroundGreen:   {backgroundGreen, background},
	ansiBackgroundYellow:  {backgroundRed | backgroundGreen, background},
	ansiBackgroundBlue:    {backgroundBlue, background},
	ansiBackgroundMagenta: {backgroundRed | backgroundBlue, background},
	ansiBackgroundCyan:    {backgroundGreen | backgroundBlue, background},
	ansiBackgroundWhite:   {backgroundRed | backgroundGreen | backgroundBlue, background},
	ansiBackgroundDefault: {0, background},

	ansiLightForegroundGray:    {foregroundIntensity, foreground},
	ansiLightForegroundRed:     {foregroundIntensity | foregroundRed, foreground},
	ansiLightForegroundGreen:   {foregroundIntensity | foregroundGreen, foreground},
	ansiLightForegroundYellow:  {foregroundIntensity | foregroundRed | foregroundGreen, foreground},
	ansiLightForegroundBlue:    {foregroundIntensity | foregroundBlue, foreground},
	ansiLightForegroundMagenta: {foregroundIntensity | foregroundRed | foregroundBlue, foreground},
	ansiLightForegroundCyan:    {foregroundIntensity | foregroundGreen | foregroundBlue, foreground},
	ansiLightForegroundWhite:   {foregroundIntensity | foregroundRed | foregroundGreen | foregroundBlue, foreground},

	ansiLightBackgroundGray:    {backgroundIntensity, background},
	ansiLightBackgroundRed:     {backgroundIntensity | backgroundRed, background},
	ansiLightBackgroundGreen:   {backgroundIntensity | backgroundGreen, background},
	ansiLightBackgroundYellow:  {backgroundIntensity | backgroundRed | backgroundGreen, background},
	ansiLightBackgroundBlue:    {backgroundIntensity | backgroundBlue, background},
	ansiLightBackgroundMagenta: {backgroundIntensity | backgroundRed | backgroundBlue, background},
	ansiLightBackgroundCyan:    {backgroundIntensity | backgroundGreen | backgroundBlue, background},
	ansiLightBackgroundWhite:   {backgroundIntensity | backgroundRed | backgroundGreen | backgroundBlue, background},
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleTextAttribute    = kernel32.NewProc("SetConsoleTextAttribute")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	defaultAttr                    *textAttributes
)

func init() {
	screenInfo := getConsoleScreenBufferInfo(uintptr(syscall.Stdout))
	if screenInfo != nil {
		colorMap[ansiForegroundDefault] = winColor{
			screenInfo.WAttributes & (foregroundRed | foregroundGreen | foregroundBlue),
			foreground,
		}
		colorMap[ansiBackgroundDefault] = winColor{
			screenInfo.WAttributes & (backgroundRed | backgroundGreen | backgroundBlue),
			background,
		}
		defaultAttr = convertTextAttr(screenInfo.WAttributes)
	}
}

type coord struct {
	X, Y int16
}

type smallRect struct {
	Left, Top, Right, Bottom int16
}

type consoleScreenBufferInfo struct {
	DwSize              coord
	DwCursorPosition    coord
	WAttributes         uint16
	SrWindow            smallRect
	DwMaximumWindowSize coord
}

func getConsoleScreenBufferInfo(hConsoleOutput uintptr) *consoleScreenBufferInfo {
	var csbi consoleScreenBufferInfo
	ret, _, _ := procGetConsoleScreenBufferInfo.Call(
		hConsoleOutput,
		uintptr(unsafe.Pointer(&csbi)))
	if ret == 0 {
		return nil
	}
	return &csbi
}

func setConsoleTextAttribute(hConsoleOutput uintptr, wAttributes uint16) bool {
	ret, _, _ := procSetConsoleTextAttribute.Call(
		hConsoleOutput,
		uintptr(wAttributes))
	return ret != 0
}

type textAttributes struct {
	foregroundColor     uint16
	backgroundColor     uint16
	foregroundIntensity uint16
	backgroundIntensity uint16
	underscore          uint16
	otherAttributes     uint16
}

func convertTextAttr(winAttr uint16) *textAttributes {
	fgColor := winAttr & (foregroundRed | foregroundGreen | foregroundBlue)
	bgColor := winAttr & (backgroundRed | backgroundGreen | backgroundBlue)
	fgIntensity := winAttr & foregroundIntensity
	bgIntensity := winAttr & backgroundIntensity
	underline := winAttr & underscore
	otherAttributes := winAttr &^ (foregroundMask | backgroundMask | underscore)
	return &textAttributes{fgColor, bgColor, fgIntensity, bgIntensity, underline, otherAttributes}
}

func convertWinAttr(textAttr *textAttributes) uint16 {
	var winAttr uint16
	winAttr |= textAttr.foregroundColor
	winAttr |= textAttr.backgroundColor
	winAttr |= textAttr.foregroundIntensity
	winAttr |= textAttr.backgroundIntensity
	winAttr |= textAttr.underscore
	winAttr |= textAttr.otherAttributes
	return winAttr
}

func changeColor(param []byte) parseResult {
	screenInfo := getConsoleScreenBufferInfo(uintptr(syscall.Stdout))
	if screenInfo == nil {
		return noConsole
	}

	winAttr := convertTextAttr(screenInfo.WAttributes)
	strParam := string(param)
	if len(strParam) <= 0 {
		strParam = "0"
	}
	csiParam := strings.Split(strParam, string(separatorChar))
	for _, p := range csiParam {
		c, ok := colorMap[p]
		switch {
		case !ok:
			switch p {
			case ansiReset:
				winAttr.foregroundColor = defaultAttr.foregroundColor
				winAttr.backgroundColor = defaultAttr.backgroundColor
				winAttr.foregroundIntensity = defaultAttr.foregroundIntensity
				winAttr.backgroundIntensity = defaultAttr.backgroundIntensity
				winAttr.underscore = 0
				winAttr.otherAttributes = 0
			case ansiIntensityOn:
				winAttr.foregroundIntensity = foregroundIntensity
			case ansiIntensityOff:
				winAttr.foregroundIntensity = 0
			case ansiUnderlineOn:
				winAttr.underscore = underscore
			case ansiUnderlineOff:
				winAttr.underscore = 0
			case ansiBlinkOn:
				winAttr.backgroundIntensity = backgroundIntensity
			case ansiBlinkOff:
				winAttr.backgroundIntensity = 0
			default:
				// unknown code
			}
		case c.drawType == foreground:
			winAttr.foregroundColor = c.code
		case c.drawType == background:
			winAttr.backgroundColor = c.code
		}
	}
	winTextAttribute := convertWinAttr(winAttr)
	setConsoleTextAttribute(uintptr(syscall.Stdout), winTextAttribute)

	return changedColor
}

func parseEscapeSequence(command byte, param []byte) parseResult {
	if defaultAttr == nil {
		return noConsole
	}

	switch command {
	case sgrCode:
		return changeColor(param)
	default:
		return unknown
	}
}

func (cw *consoleWriter) flushBuffer() (int, error) {
	return cw.flushTo(cw.w)
}

func (cw *consoleWriter) resetBuffer() (int, error) {
	return cw.flushTo(nil)
}

func (cw *consoleWriter) flushTo(w io.Writer) (int, error) {
	var n1, n2 int
	var err error

	startBytes := cw.paramStartBuf.Bytes()
	cw.paramStartBuf.Reset()
	if w != nil {
		n1, err = cw.w.Write(startBytes)
		if err != nil {
			return n1, err
		}
	} else {
		n1 = len(startBytes)
	}
	paramBytes := cw.paramBuf.Bytes()
	cw.paramBuf.Reset()
	if w != nil {
		n2, err = cw.w.Write(paramBytes)
		if err != nil {
			return n1 + n2, err
		}
	} else {
		n2 = len(paramBytes)
	}
	return n1 + n2, nil
}

func isParameterChar(b byte) bool {
	return ('0' <= b && b <= '9') || b == separatorChar
}

func (cw *consoleWriter) Write(p []byte) (int, error) {
	var r, nw, first, last int
	if cw.mode != DiscardNonColorEscSeq {
		cw.state = outsideCsiCode
		cw.resetBuffer()
	}

	var err error
	for i, ch := range p {
		switch cw.state {
		case outsideCsiCode:
			if ch == firstCsiChar {
				cw.paramStartBuf.WriteByte(ch)
				cw.state = firstCsiCode
			}
		case firstCsiCode:
			switch ch {
			case firstCsiChar:
				cw.paramStartBuf.WriteByte(ch)
				break
			case secondeCsiChar:
				cw.paramStartBuf.WriteByte(ch)
				cw.state = secondCsiCode
				last = i - 1
			default:
				cw.resetBuffer()
				cw.state = outsideCsiCode
			}
		case secondCsiCode:
			if isParameterChar(ch) {
				cw.paramBuf.WriteByte(ch)
			} else {
				nw, err = cw.w.Write(p[first:last])
				r += nw
				if err != nil {
					return r, err
				}
				first = i + 1
				result := parseEscapeSequence(ch, cw.paramBuf.Bytes())
				if result == noConsole || (cw.mode == OutputNonColorEscSeq && result == unknown) {
					cw.paramBuf.WriteByte(ch)
					nw, err := cw.flushBuffer()
					if err != nil {
						return r, err
					}
					r += nw
				} else {
					n, _ := cw.resetBuffer()
					// Add one more to the size of the buffer for the last ch
					r += n + 1
				}

				cw.state = outsideCsiCode
			}
		default:
			cw.state = outsideCsiCode
		}
	}

	if cw.mode != DiscardNonColorEscSeq || cw.state == outsideCsiCode {
		nw, err = cw.w.Write(p[first:])
		r += nw
	}

	return r, err
}

