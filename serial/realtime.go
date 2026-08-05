package serial

const (
	CmdStatusReport byte = '?'
	CmdFeedHold     byte = '!'
	CmdCycleResume  byte = '~'
	CmdSoftReset    byte = 0x18
	CmdJogCancel    byte = 0x85

	CmdFeedOvrReset   byte = 0x90
	CmdFeedOvrPlus10  byte = 0x91
	CmdFeedOvrMinus10 byte = 0x92
	CmdFeedOvrPlus1   byte = 0x93
	CmdFeedOvrMinus1  byte = 0x94

	CmdRapidOvrReset byte = 0x95
	CmdRapidOvr50    byte = 0x96
	CmdRapidOvr25    byte = 0x97

	CmdSpindleOvrReset   byte = 0x99
	CmdSpindleOvrPlus10  byte = 0x9A
	CmdSpindleOvrMinus10 byte = 0x9B
	CmdSpindleOvrPlus1   byte = 0x9C
	CmdSpindleOvrMinus1  byte = 0x9D
)
