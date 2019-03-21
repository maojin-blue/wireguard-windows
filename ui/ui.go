/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"time"

	"github.com/lxn/walk"
	"golang.zx2c4.com/wireguard/windows/service"
)

const nagMessage = `It looks like you're still using this WireGuard pre-alpha build. Great!

We're glad you like it, and we'd appreciate you sharing both your successes and your tribulations with us via team@wireguard.com or #wireguard-windows on Freenode.

But because this is pre-release software, we're not confident it's something you should yet be using, except for testing and reporting bugs. Check back with us for a newer version.

Would you like to quit WireGuard now? If not, you'll be nagged again in two minutes about the same thing.`

func nag() {
	if walk.MsgBox(nil, "THANKS FOR REPORTING BUGS COME AGAIN ANOTHER DAY", nagMessage, walk.MsgBoxIconError|walk.MsgBoxYesNo|0x00001000) != walk.DlgCmdNo {
		onQuit()
	}
	time.AfterFunc(time.Minute*2, nag)
}

func RunUI() {
	icon, err := walk.NewIconFromResourceId(1)
	if err != nil {
		panic(err)
	}
	defer icon.Dispose()

	mtw, err := NewManageTunnelsWindow(icon)
	if err != nil {
		panic(err)
	}
	defer mtw.Dispose()

	tray, err := NewTray(mtw, icon)
	if err != nil {
		panic(err)
	}
	defer tray.Dispose()

	// Bind to updates
	service.IPCClientRegisterTunnelChange(func(tunnel *service.Tunnel, state service.TunnelState, err error) {
		if err == nil {
			return
		}

		if mtw.Visible() {
			errMsg := err.Error()
			if len(errMsg) > 0 && errMsg[len(errMsg)-1] != '.' {
				errMsg += "."
			}
			walk.MsgBox(mtw, "Tunnel Error", errMsg()+"\n\nPlease consult the Windows Event Log for more information.", walk.MsgBoxIconWarning)
		} else {
			tray.ShowError("WireGuard Tunnel Error", err.Error())
		}
	})
	service.IPCClientRegisterTunnelChange(func(tunnel *service.Tunnel, state service.TunnelState, err error) {
		tray.setTunnelStateWithNotification(tunnel, state, err == nil)
	})
	service.IPCClientRegisterTunnelChange(func(tunnel *service.Tunnel, state service.TunnelState, err error) {
		mtw.setTunnelState(tunnel, state)
	})

	// Fetch current state
	go func() {
		tunnels, err := service.IPCClientTunnels()
		if err != nil {
			return
		}
		for _, tunnel := range tunnels {
			state, err := tunnel.State()
			if err != nil {
				continue
			}
			tray.setTunnelStateWithNotification(&tunnel, state, false)
			mtw.setTunnelState(&tunnel, state)
		}
	}()

	time.AfterFunc(time.Minute*15, nag)
	mtw.Run()
}
