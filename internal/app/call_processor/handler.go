package processor

import (
	"log/slog"

	"github.com/webitel/call_audit/model"
)

type App struct {
	Cfg    *Config
	State  *UUIDState
	Runner *Runner
}

func NewApp(cfg *Config) *App {
	return &App{
		Cfg:    cfg,
		State:  NewUUIDState(),
		Runner: NewRunner(cfg),
	}
}

func (a *App) Process(job *model.CallJob) map[string]any {
	if job.Params.CallID == "" || job.Params.CallID == "0" {
		return map[string]any{
			"status": "error",
			"error":  "call_id is required",
		}
	}

	if !a.State.TryAdd(job.Params.CallID) {
		slog.Warn("Call ID is already being processed", slog.String("call_id", job.Params.CallID))
		return map[string]any{
			"status":  "already_processing",
			"call_id": job.Params.CallID,
		}
	}

	slog.Info("Accepted Call ID", slog.String("call_id", job.Params.CallID))
	go func() {
		if err := a.Runner.ProcessUUID(job); err != nil {
			slog.Error("Failed to process Call ID", slog.String("call_id", job.Params.CallID), slog.String("error", err.Error()))
		}
		a.State.Remove(job.Params.CallID)

		slog.Info("Finished processing Call ID", slog.String("call_id", job.Params.CallID))
	}()

	return map[string]any{
		"status":  "accepted",
		"call_id": job.Params.CallID,
	}
}
