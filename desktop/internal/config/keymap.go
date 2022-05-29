package config

import (
	"github.com/micmonay/keybd_event"
)

var (
	// JSKeyMap return keyboard mapping between JS codes and Robotgo keys
	JSKeyMap map[int]int
)

func init() {
	JSKeyMap = make(map[int]int, 210)

	JSKeyMap[8] = keybd_event.VK_BACKSPACE
	JSKeyMap[9] = keybd_event.VK_TAB
	JSKeyMap[18] = keybd_event.VK_KANA
	JSKeyMap[13] = keybd_event.VK_ENTER
	JSKeyMap[20] = keybd_event.VK_CAPSLOCK
	JSKeyMap[27] = keybd_event.VK_ESC
	JSKeyMap[32] = keybd_event.VK_SPACE
	JSKeyMap[33] = keybd_event.VK_PAGEUP
	JSKeyMap[34] = keybd_event.VK_PAGEDOWN
	JSKeyMap[35] = keybd_event.VK_END
	JSKeyMap[36] = keybd_event.VK_HOME
	JSKeyMap[37] = keybd_event.VK_LEFT
	JSKeyMap[38] = keybd_event.VK_UP
	JSKeyMap[39] = keybd_event.VK_RIGHT
	JSKeyMap[40] = keybd_event.VK_DOWN
	JSKeyMap[44] = keybd_event.VK_PRINT
	JSKeyMap[45] = keybd_event.VK_INSERT
	JSKeyMap[46] = keybd_event.VK_DELETE
	JSKeyMap[48] = keybd_event.VK_0
	JSKeyMap[49] = keybd_event.VK_1
	JSKeyMap[50] = keybd_event.VK_2
	JSKeyMap[51] = keybd_event.VK_3
	JSKeyMap[52] = keybd_event.VK_4
	JSKeyMap[53] = keybd_event.VK_5
	JSKeyMap[54] = keybd_event.VK_6
	JSKeyMap[55] = keybd_event.VK_7
	JSKeyMap[56] = keybd_event.VK_8
	JSKeyMap[57] = keybd_event.VK_9
	JSKeyMap[65] = keybd_event.VK_A
	JSKeyMap[66] = keybd_event.VK_B
	JSKeyMap[67] = keybd_event.VK_C
	JSKeyMap[68] = keybd_event.VK_D
	JSKeyMap[69] = keybd_event.VK_E
	JSKeyMap[70] = keybd_event.VK_F
	JSKeyMap[71] = keybd_event.VK_G
	JSKeyMap[72] = keybd_event.VK_H
	JSKeyMap[73] = keybd_event.VK_I
	JSKeyMap[74] = keybd_event.VK_J
	JSKeyMap[75] = keybd_event.VK_K
	JSKeyMap[76] = keybd_event.VK_L
	JSKeyMap[77] = keybd_event.VK_M
	JSKeyMap[78] = keybd_event.VK_N
	JSKeyMap[79] = keybd_event.VK_O
	JSKeyMap[80] = keybd_event.VK_P
	JSKeyMap[81] = keybd_event.VK_Q
	JSKeyMap[82] = keybd_event.VK_R
	JSKeyMap[83] = keybd_event.VK_S
	JSKeyMap[84] = keybd_event.VK_T
	JSKeyMap[85] = keybd_event.VK_U
	JSKeyMap[86] = keybd_event.VK_V
	JSKeyMap[87] = keybd_event.VK_W
	JSKeyMap[88] = keybd_event.VK_X
	JSKeyMap[89] = keybd_event.VK_Y
	JSKeyMap[90] = keybd_event.VK_Z

	JSKeyMap[91] = keybd_event.VK_LMENU
	JSKeyMap[93] = keybd_event.VK_RMENU

	JSKeyMap[96] = keybd_event.VK_KP0
	JSKeyMap[97] = keybd_event.VK_KP1
	JSKeyMap[98] = keybd_event.VK_KP2
	JSKeyMap[99] = keybd_event.VK_KP3
	JSKeyMap[100] = keybd_event.VK_KP4
	JSKeyMap[101] = keybd_event.VK_KP5
	JSKeyMap[102] = keybd_event.VK_KP6
	JSKeyMap[103] = keybd_event.VK_KP7
	JSKeyMap[104] = keybd_event.VK_KP8
	JSKeyMap[105] = keybd_event.VK_KP9
	JSKeyMap[107] = keybd_event.VK_KPPLUS
	JSKeyMap[109] = keybd_event.VK_KPMINUS
	JSKeyMap[110] = keybd_event.VK_KPDOT

	JSKeyMap[112] = keybd_event.VK_F1
	JSKeyMap[113] = keybd_event.VK_F2
	JSKeyMap[114] = keybd_event.VK_F3
	JSKeyMap[115] = keybd_event.VK_F4
	JSKeyMap[116] = keybd_event.VK_F5
	JSKeyMap[117] = keybd_event.VK_F6
	JSKeyMap[118] = keybd_event.VK_F7
	JSKeyMap[119] = keybd_event.VK_F8
	JSKeyMap[120] = keybd_event.VK_F9
	JSKeyMap[121] = keybd_event.VK_F10
	JSKeyMap[122] = keybd_event.VK_F11
	JSKeyMap[123] = keybd_event.VK_F12

	JSKeyMap[144] = keybd_event.VK_NUMLOCK

	JSKeyMap[186] = keybd_event.VK_SEMICOLON
	JSKeyMap[187] = keybd_event.VK_EQUAL
	JSKeyMap[188] = keybd_event.VK_COMMA
	JSKeyMap[189] = keybd_event.VK_MINUS
	JSKeyMap[190] = keybd_event.VK_DOT
	JSKeyMap[191] = keybd_event.VK_SLASH
	JSKeyMap[192] = keybd_event.VK_GRAVE
	JSKeyMap[219] = keybd_event.VK_LEFTBRACE
	JSKeyMap[220] = keybd_event.VK_BACKSLASH
	JSKeyMap[221] = keybd_event.VK_RIGHTBRACE
	JSKeyMap[222] = keybd_event.VK_APOSTROPHE
}
