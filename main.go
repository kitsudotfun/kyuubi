package main

import (
	"context"
	"net/http"

	"github.com/kitsudotfun/kyuubi/api"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/cron"
)

func main() {
	// clean up database nightly
	cron.ScheduleTaskNonBlock(func(ctx context.Context) error { return api.CleanServers() })

	mux := http.NewServeMux()
	api.RegisterHandlers(mux)
	workers.Serve(mux)
}
