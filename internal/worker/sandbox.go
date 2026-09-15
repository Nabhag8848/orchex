package worker

import (
	"context"
	"encoding/json"
	"log/slog"
)

// SandboxInvoker executes source code in the function sandbox.
type SandboxInvoker interface {
	Invoke(context.Context, string, json.RawMessage, int) (json.RawMessage, error)
}

// CheckSandbox performs a best-effort startup invocation and logs the result.
// Call it in a goroutine so the check does not delay worker startup.
func CheckSandbox(ctx context.Context, client SandboxInvoker) {
	result, err := client.Invoke(ctx, "return { ping: true };", json.RawMessage(`{"data":{}}`), 5000)
	if err != nil {
		slog.Warn("sandbox startup invoke failed", "error", err)
		return
	}
	slog.Debug("sandbox startup invoke succeeded", "result", string(result))
}
