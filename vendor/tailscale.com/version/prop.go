// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package version

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
)

// IsMobile reports whether this is a mobile client build.
func IsMobile() bool {
	return runtime.GOOS == "android" || runtime.GOOS == "ios"
}

// OS returns runtime.GOOS, except instead of returning "darwin" it
// returns "iOS" or "macOS".
func OS() string {
	if runtime.GOOS == "ios" {
		return "iOS"
	}
	if runtime.GOOS == "darwin" {
		return "macOS"
	}
	return runtime.GOOS
}

// IsSandboxedMacOS reports whether this process is a sandboxed macOS
// process. It is true for the Mac App Store and macsys (System
// Extension) version on macOS, and false for tailscaled-on-macOS.
func IsSandboxedMacOS() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	if IsMacSysExt() {
		return true
	}
	exe, _ := os.Executable()
	return strings.HasSuffix(exe, "/Contents/MacOS/Tailscale")
}

var isMacSysExt atomic.Value

// IsMacSysExt whether this binary is from the standalone "System
// Extension" (a.k.a. "macsys") version of Tailscale for macOS.
func IsMacSysExt() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	if b, ok := isMacSysExt.Load().(bool); ok {
		return b
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	v := filepath.Base(exe) == "io.tailscale.ipn.macsys.network-extension"
	isMacSysExt.Store(v)
	return v
}
