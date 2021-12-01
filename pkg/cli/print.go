// Copyright 2021 Danny Hermes
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

// Printf formats according to a format specifier and writes to the STDOUT
// attached to the current context.
func Printf(ctx context.Context, format string, a ...interface{}) (int, error) {
	w := GetStdout(ctx)
	return fmt.Fprintf(w, format, a...)
}

// DebugPrintf invokes `Printf()` if the debug flag is set on the current
// context and adds a `[DEBUG] ` prefix to every line printed.
func DebugPrintf(ctx context.Context, format string, a ...interface{}) (int, error) {
	d := GetDebug(ctx)
	if !d {
		return 0, nil
	}
	// Hack: write the entire output to a buffer, then add a `[DEBUG] ` prefix
	// to every line before passing along to `Printf()`.
	b := bytes.NewBuffer(nil)
	_, err := fmt.Fprintf(b, format, a...)
	if err != nil {
		return 0, err
	}
	return writeWithDebugPrefix(ctx, b.String())
}

func writeWithDebugPrefix(ctx context.Context, s string) (int, error) {
	parts := strings.Split(s, "\n")
	for i, part := range parts {
		parts[i] = fmt.Sprintf("[DEBUG] %s", part)
	}

	// Reset any trailing newline
	last := len(parts) - 1
	if parts[last] == "[DEBUG] " {
		parts[last] = ""
	}

	withPrefix := strings.Join(parts, "\n")
	return Printf(ctx, withPrefix)
}

// Println formats using the default formats amd writes to the STDOUT
// attached to the current context. Spaces are always added between operands
// and a newline is appended.
func Println(ctx context.Context, a ...interface{}) (int, error) {
	w := GetStdout(ctx)
	return fmt.Fprintln(w, a...)
}

// DebugPrintln invokes `Println()` if the debug flag is set on the current
// context and adds a `[DEBUG] ` prefix to every line printed.
func DebugPrintln(ctx context.Context, a ...interface{}) (int, error) {
	d := GetDebug(ctx)
	if !d {
		return 0, nil
	}
	// Hack: write the entire output to a buffer, then add a `[DEBUG] ` prefix
	// to every line before passing along to `Printf()`.
	b := bytes.NewBuffer(nil)
	_, err := fmt.Fprintln(b, a...)
	if err != nil {
		return 0, err
	}
	return writeWithDebugPrefix(ctx, b.String())
}
