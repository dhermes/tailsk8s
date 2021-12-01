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
	"context"
	"io"
	"os"
)

type stdoutKey struct{}

type stderrKey struct{}

type debugKey struct{}

// WithStdout adds a STDOUT writer to a context.
func WithStdout(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, stdoutKey{}, w)
}

// GetStdout enables STDOUT to be specified on a context; if not provided,
// falls back to `os.Stdout`.
func GetStdout(ctx context.Context) io.Writer {
	w, ok := ctx.Value(stdoutKey{}).(io.Writer)
	if ok && w != nil {
		return w
	}
	return os.Stdout
}

// WithStderr adds a STDERR writer to a context.
func WithStderr(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, stderrKey{}, w)
}

// GetStderr enables STDERR to be specified on a context; if not provided,
// falls back to `os.Stderr`.
func GetStderr(ctx context.Context) io.Writer {
	w, ok := ctx.Value(stderrKey{}).(io.Writer)
	if ok && w != nil {
		return w
	}
	return os.Stderr
}

// WithDebug sets a "debug mode" flag on a context.
func WithDebug(ctx context.Context, d bool) context.Context {
	return context.WithValue(ctx, debugKey{}, d)
}

// GetDebug gets a "debug mode" flag from a context, or returns `false` if
// not set.
func GetDebug(ctx context.Context) bool {
	b, _ := ctx.Value(debugKey{}).(bool)
	return b
}
