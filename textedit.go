// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

//包名 walk
package walk
//引用
import (
	"syscall"
	"unsafe"
)
//引用
import (
	"github.com/lxn/win"
)
//结构 父类是widget，两个Event publisher ，一个字体颜色。
type TextEdit struct {
	WidgetBase
	readOnlyChangedPublisher EventPublisher
	textChangedPublisher     EventPublisher
	textColor                Color
}
//由父类容器传来创建 返回TextEdit的实例。
func NewTextEdit(parent Container) (*TextEdit, error) {
	return NewTextEditWithStyle(parent, 0)
}
//由父类容器传来创建，带格式，返回TextEdit实例。
func NewTextEditWithStyle(parent Container, style uint32) (*TextEdit, error) {
	te := new(TextEdit)

	if err := InitWidget(
		te,
		parent,
		"EDIT",
		win.WS_TABSTOP|win.WS_VISIBLE|win.ES_MULTILINE|win.ES_WANTRETURN|style,
		win.WS_EX_CLIENTEDGE); err != nil {
		return nil, err
	}

	te.GraphicsEffects().Add(InteractionEffect)
	te.GraphicsEffects().Add(FocusEffect)

	te.MustRegisterProperty("ReadOnly", NewProperty(
		func() interface{} {
			return te.ReadOnly()
		},
		func(v interface{}) error {
			return te.SetReadOnly(v.(bool))
		},
		te.readOnlyChangedPublisher.Event()))

	te.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return te.Text()
		},
		func(v interface{}) error {
			return te.SetText(v.(string))
		},
		te.textChangedPublisher.Event()))

	return te, nil
}

func (*TextEdit) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (te *TextEdit) MinSizeHint() Size {
	return te.dialogBaseUnitsToPixels(Size{20, 12})
}

func (te *TextEdit) SizeHint() Size {
	return Size{100, 100}
}
//Text方法可得到内容。字符串。
func (te *TextEdit) Text() string {
	return windowText(te.hWnd)
}
//内容长度
func (te *TextEdit) TextLength() int {
	return int(te.SendMessage(win.WM_GETTEXTLENGTH, 0, 0))
}
//设置内容
func (te *TextEdit) SetText(value string) (err error) {
	if value == te.Text() {
		return nil
	}

	err = setWindowText(te.hWnd, value)
	te.textChangedPublisher.Publish()
	return
}
//对齐方式获得，可得到uint32的ID。
func (te *TextEdit) Alignment() Alignment1D {
	switch win.GetWindowLong(te.hWnd, win.GWL_STYLE) & (win.ES_LEFT | win.ES_CENTER | win.ES_RIGHT) {
	case win.ES_CENTER:
		return AlignCenter

	case win.ES_RIGHT:
		return AlignFar
	}

	return AlignNear
}
//设置对齐方式，使用AlignCenter居中，AlignFar右对齐，默认是左对齐。
func (te *TextEdit) SetAlignment(alignment Alignment1D) error {
	var bit uint32

	switch alignment {
	case AlignCenter:
		bit = win.ES_CENTER

	case AlignFar:
		bit = win.ES_RIGHT

	default:
		bit = win.ES_LEFT
	}

	return te.ensureStyleBits(bit, true)
}
//最大长度
func (te *TextEdit) MaxLength() int {
	return int(te.SendMessage(win.EM_GETLIMITTEXT, 0, 0))
}
//设置对打允许长度。
func (te *TextEdit) SetMaxLength(value int) {
	te.SendMessage(win.EM_SETLIMITTEXT, uintptr(value), 0)
}
//内容选中
func (te *TextEdit) TextSelection() (start, end int) {
	te.SendMessage(win.EM_GETSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return
}
//设置选中内容。
func (te *TextEdit) SetTextSelection(start, end int) {
	te.SendMessage(win.EM_SETSEL, uintptr(start), uintptr(end))
}
//替换选择内容。
func (te *TextEdit) ReplaceSelectedText(text string, canUndo bool) {
	te.SendMessage(win.EM_REPLACESEL,
		uintptr(win.BoolToBOOL(canUndo)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
}
//追加内容
func (te *TextEdit) AppendText(value string) {
	s, e := te.TextSelection()
	l := te.TextLength()
	te.SetTextSelection(l, l)
	te.ReplaceSelectedText(value, false)
	te.SetTextSelection(s, e)
}
//只读属性
func (te *TextEdit) ReadOnly() bool {
	return te.hasStyleBits(win.ES_READONLY)
}
//设置只读属性
func (te *TextEdit) SetReadOnly(readOnly bool) error {
	if 0 == te.SendMessage(win.EM_SETREADONLY, uintptr(win.BoolToBOOL(readOnly)), 0) {
		return newError("SendMessage(EM_SETREADONLY)")
	}

	te.readOnlyChangedPublisher.Publish()

	return nil
}
//文字更新事件。
func (te *TextEdit) TextChanged() *Event {
	return te.textChangedPublisher.Event()
}
//文字颜色
func (te *TextEdit) TextColor() Color {
	return te.textColor
}
//设置文字颜色。
func (te *TextEdit) SetTextColor(c Color) {
	te.textColor = c

	te.Invalidate()
}
//用他来实现绑定。
func (te *TextEdit) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case win.EN_CHANGE:
			te.textChangedPublisher.Publish()
		}

	case win.WM_GETDLGCODE:
		if wParam == win.VK_RETURN {
			return win.DLGC_WANTALLKEYS
		}

		return win.DLGC_HASSETSEL | win.DLGC_WANTARROWS | win.DLGC_WANTCHARS

	case win.WM_KEYDOWN:
		if Key(wParam) == KeyA && ControlDown() {
			te.SetTextSelection(0, -1)
		}
	}

	return te.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}
