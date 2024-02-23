// SPDX-License-Identifier: Apache-2.0

package nuke

import "context"

type contextKey int

const (
	arenaContextKey contextKey = 0
)

// InjectContextArena returns a new context with the Arena injected into it.
func InjectContextArena(ctx context.Context, a Arena) context.Context {
	return context.WithValue(ctx, arenaContextKey, a)
}

// ExtractContextArena returns the Arena from the context.
func ExtractContextArena(ctx context.Context) Arena {
	if a, ok := ctx.Value(arenaContextKey).(Arena); ok {
		return a
	}
	return nil
}
