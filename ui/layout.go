package ui

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	colorGreen  = color.NRGBA{R: 0x2e, G: 0x7d, B: 0x32, A: 0xff}
	colorRed    = color.NRGBA{R: 0xc6, G: 0x28, B: 0x28, A: 0xff}
	colorOrange = color.NRGBA{R: 0xe6, G: 0x5c, B: 0x00, A: 0xff}
	colorGray   = color.NRGBA{R: 0x75, G: 0x75, B: 0x75, A: 0xff}
	consoleBg   = color.NRGBA{R: 0x1e, G: 0x1e, B: 0x2e, A: 0xff}
	consoleFg   = color.NRGBA{R: 0xcd, G: 0xd6, B: 0xf4, A: 0xff}
	sepColor    = color.NRGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff}
)

func (a *App) layout(gtx C) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(a.layoutToolbar),
		layout.Rigid(func(gtx C) D { return hSep(gtx) }),
		layout.Flexed(1, a.layoutBody),
		layout.Rigid(func(gtx C) D { return hSep(gtx) }),
		layout.Rigid(a.layoutInputBar),
		layout.Rigid(func(gtx C) D { return hSep(gtx) }),
		layout.Rigid(a.layoutSendBar),
	)
}

func (a *App) layoutToolbar(gtx C) D {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return material.Body2(a.theme, "Port:").Layout(gtx)
			}),
			layout.Rigid(spacerW(4)),
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(220)
				return material.Editor(a.theme, &a.portEditor, "/dev/...").Layout(gtx)
			}),
			layout.Rigid(spacerW(8)),
			layout.Rigid(func(gtx C) D {
				label := "Connect"
				bg := colorGreen
				if a.connected {
					label = "Disconnect"
					bg = colorOrange
				}
				btn := material.Button(a.theme, &a.connectBtn, label)
				btn.Background = bg
				return btn.Layout(gtx)
			}),
			layout.Rigid(spacerW(24)),
			layout.Rigid(func(gtx C) D {
				return material.Body2(a.theme, "File:").Layout(gtx)
			}),
			layout.Rigid(spacerW(4)),
			layout.Flexed(1, func(gtx C) D {
				return material.Editor(a.theme, &a.fileEditor, "path/to/file.gcode").Layout(gtx)
			}),
			layout.Rigid(spacerW(8)),
			layout.Rigid(func(gtx C) D {
				return material.Button(a.theme, &a.loadBtn, "Load").Layout(gtx)
			}),
		)
	})
}

func (a *App) layoutBody(gtx C) D {
	return layout.Flex{}.Layout(gtx,
		layout.Flexed(1, a.layoutConsole),
		layout.Rigid(func(gtx C) D { return vSep(gtx) }),
		layout.Rigid(a.layoutSidePanel),
	)
}

func (a *App) layoutConsole(gtx C) D {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, consoleBg)
			return D{Size: gtx.Constraints.Max}
		}),
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx C) D {
				return material.List(a.theme, &a.consoleList).Layout(gtx, len(a.logLines), func(gtx C, i int) D {
					l := material.Body2(a.theme, a.logLines[i])
					l.Color = consoleFg
					return l.Layout(gtx)
				})
			})
		}),
	)
}

func (a *App) layoutSidePanel(gtx C) D {
	gtx.Constraints.Min.X = gtx.Dp(260)
	gtx.Constraints.Max.X = gtx.Dp(260)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(a.layoutStatus),
		layout.Rigid(func(gtx C) D { return hSep(gtx) }),
		layout.Rigid(a.layoutJogPad),
	)
}

func (a *App) layoutStatus(gtx C) D {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx C) D {
		state := "Disconnected"
		stateColor := colorGray
		var x, y, z, feed, spindle float64
		if a.connected {
			state = "Connected"
			stateColor = colorGreen
		}
		if a.status != nil {
			state = a.status.State.String()
			feed = a.status.Feed
			spindle = a.status.Spindle
			if a.status.MPos != nil {
				x, y, z = a.status.MPos.X, a.status.MPos.Y, a.status.MPos.Z
			} else if a.status.WPos != nil {
				x, y, z = a.status.WPos.X, a.status.WPos.Y, a.status.WPos.Z
			}
		}

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				l := material.H6(a.theme, state)
				l.Color = stateColor
				return l.Layout(gtx)
			}),
			layout.Rigid(spacerH(4)),
			layout.Rigid(statusLine(a.theme, "X", fmt.Sprintf("%.3f mm", x))),
			layout.Rigid(statusLine(a.theme, "Y", fmt.Sprintf("%.3f mm", y))),
			layout.Rigid(statusLine(a.theme, "Z", fmt.Sprintf("%.3f mm", z))),
			layout.Rigid(spacerH(4)),
			layout.Rigid(statusLine(a.theme, "Feed", fmt.Sprintf("%.0f mm/min", feed))),
			layout.Rigid(statusLine(a.theme, "Spindle", fmt.Sprintf("%.0f RPM", spindle))),
		)
	})
}

func (a *App) layoutJogPad(gtx C) D {
	return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx C) D {
		bw := unit.Dp(56)
		gap := unit.Dp(4)

		btn := func(c *widget.Clickable, label string) layout.Widget {
			return func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(bw)
				return material.Button(a.theme, c, label).Layout(gtx)
			}
		}
		empty := func(gtx C) D {
			return D{Size: image.Pt(gtx.Dp(bw), 0)}
		}
		sp := spacerW(gap)

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				l := material.H6(a.theme, "Jog")
				return l.Layout(gtx)
			}),
			layout.Rigid(spacerH(8)),
			// Row 1: _, Y+, _, Z+
			layout.Rigid(func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(empty), layout.Rigid(sp),
					layout.Rigid(btn(&a.jogYP, "Y+")), layout.Rigid(sp),
					layout.Rigid(empty), layout.Rigid(sp),
					layout.Rigid(btn(&a.jogZP, "Z+")),
				)
			}),
			layout.Rigid(spacerH(gap)),
			// Row 2: X-, Home, X+
			layout.Rigid(func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(btn(&a.jogXM, "X-")), layout.Rigid(sp),
					layout.Rigid(func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Dp(bw)
						b := material.Button(a.theme, &a.jogHome, "H")
						b.Background = colorOrange
						return b.Layout(gtx)
					}), layout.Rigid(sp),
					layout.Rigid(btn(&a.jogXP, "X+")),
				)
			}),
			layout.Rigid(spacerH(gap)),
			// Row 3: _, Y-, _, Z-
			layout.Rigid(func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(empty), layout.Rigid(sp),
					layout.Rigid(btn(&a.jogYM, "Y-")), layout.Rigid(sp),
					layout.Rigid(empty), layout.Rigid(sp),
					layout.Rigid(btn(&a.jogZM, "Z-")),
				)
			}),
			layout.Rigid(spacerH(12)),
			// Step buttons
			layout.Rigid(func(gtx C) D {
				return material.Body2(a.theme, fmt.Sprintf("Step: %.1f mm", a.jogStep)).Layout(gtx)
			}),
			layout.Rigid(spacerH(4)),
			layout.Rigid(func(gtx C) D {
				steps := [4]string{"0.1", "1", "10", "100"}
				stepVals := [4]float64{0.1, 1, 10, 100}
				children := make([]layout.FlexChild, 0, 7)
				for i := range a.jogStepBtns {
					i := i
					if i > 0 {
						children = append(children, layout.Rigid(sp))
					}
					children = append(children, layout.Rigid(func(gtx C) D {
						b := material.Button(a.theme, &a.jogStepBtns[i], steps[i])
						if a.jogStep == stepVals[i] {
							b.Background = colorGreen
						} else {
							b.Background = colorGray
						}
						return b.Layout(gtx)
					}))
				}
				return layout.Flex{}.Layout(gtx, children...)
			}),
		)
	})
}

func (a *App) layoutInputBar(gtx C) D {
	return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return material.Body2(a.theme, ">").Layout(gtx)
			}),
			layout.Rigid(spacerW(4)),
			layout.Flexed(1, func(gtx C) D {
				return material.Editor(a.theme, &a.cmdEditor, "G-code command...").Layout(gtx)
			}),
			layout.Rigid(spacerW(8)),
			layout.Rigid(func(gtx C) D {
				return material.Button(a.theme, &a.sendCmdBtn, "Send").Layout(gtx)
			}),
		)
	})
}

func (a *App) layoutSendBar(gtx C) D {
	return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx C) D {
		if len(a.fileLines) == 0 {
			l := material.Body2(a.theme, "No file loaded")
			l.Color = colorGray
			return l.Layout(gtx)
		}

		if !a.sending {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					btn := material.Button(a.theme, &a.sendFileBtn, "Send File")
					btn.Background = colorGreen
					if !a.connected {
						btn.Background = colorGray
					}
					return btn.Layout(gtx)
				}),
				layout.Rigid(spacerW(12)),
				layout.Rigid(func(gtx C) D {
					return material.Body2(a.theme, fmt.Sprintf("%d lines ready", len(a.fileLines))).Layout(gtx)
				}),
			)
		}

		n := a.acked.Load()
		pct := float32(0)
		if a.total > 0 {
			pct = float32(n) / float32(a.total)
		}
		pauseLabel := "Pause"
		if a.paused {
			pauseLabel = "Resume"
		}

		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				btn := material.Button(a.theme, &a.pauseBtn, pauseLabel)
				btn.Background = colorOrange
				return btn.Layout(gtx)
			}),
			layout.Rigid(spacerW(4)),
			layout.Rigid(func(gtx C) D {
				btn := material.Button(a.theme, &a.stopBtn, "Stop")
				btn.Background = colorRed
				return btn.Layout(gtx)
			}),
			layout.Rigid(spacerW(12)),
			layout.Flexed(1, func(gtx C) D {
				return material.ProgressBar(a.theme, pct).Layout(gtx)
			}),
			layout.Rigid(spacerW(8)),
			layout.Rigid(func(gtx C) D {
				return material.Body2(a.theme, fmt.Sprintf("%d/%d (%d%%)", n, a.total, int(pct*100))).Layout(gtx)
			}),
		)
	})
}

func statusLine(th *material.Theme, label, value string) layout.Widget {
	return func(gtx C) D {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(70)
				l := material.Body2(th, label)
				l.Color = colorGray
				return l.Layout(gtx)
			}),
			layout.Rigid(func(gtx C) D {
				return material.Body1(th, value).Layout(gtx)
			}),
		)
	}
}

func spacerW(dp unit.Dp) layout.Widget {
	return func(gtx C) D {
		return D{Size: image.Pt(gtx.Dp(dp), 0)}
	}
}

func spacerH(dp unit.Dp) layout.Widget {
	return func(gtx C) D {
		return D{Size: image.Pt(0, gtx.Dp(dp))}
	}
}

func hSep(gtx C) D {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, sepColor)
	return D{Size: size}
}

func vSep(gtx C) D {
	size := image.Pt(gtx.Dp(1), gtx.Constraints.Max.Y)
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, sepColor)
	return D{Size: size}
}
