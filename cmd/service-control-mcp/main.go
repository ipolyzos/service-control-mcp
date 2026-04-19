// Command service-control-mcp is the MCP stdio server for managing system services.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ipolyzos/service-control-mcp/internal/audit"
	"github.com/ipolyzos/service-control-mcp/internal/platform"
	"github.com/ipolyzos/service-control-mcp/internal/plugin"
	"github.com/ipolyzos/service-control-mcp/internal/runtime"

	// Register platform adapters.
	_ "github.com/ipolyzos/service-control-mcp/internal/platform/launchd"
	_ "github.com/ipolyzos/service-control-mcp/internal/platform/systemd"

	describe "github.com/ipolyzos/service-control-mcp/plugins/describe_service"
	getlogs "github.com/ipolyzos/service-control-mcp/plugins/get_service_logs"
	getstatus "github.com/ipolyzos/service-control-mcp/plugins/get_service_status"
	list "github.com/ipolyzos/service-control-mcp/plugins/list_services"
	reload "github.com/ipolyzos/service-control-mcp/plugins/reload_service"
	restart "github.com/ipolyzos/service-control-mcp/plugins/restart_service"
	start "github.com/ipolyzos/service-control-mcp/plugins/start_service"
	stop "github.com/ipolyzos/service-control-mcp/plugins/stop_service"
	validate "github.com/ipolyzos/service-control-mcp/plugins/validate_service_name"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "service-control-mcp: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	mgr, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("detect platform: %w", err)
	}

	auditLog := audit.NewLogger(os.Stderr)

	reg := plugin.NewRegistry()
	tools := []plugin.Tool{
		list.New(mgr),
		getstatus.New(mgr),
		describe.New(mgr),
		validate.New(),
		getlogs.New(mgr),
		start.New(mgr, auditLog),
		stop.New(mgr, auditLog),
		restart.New(mgr, auditLog),
		reload.New(mgr, auditLog),
	}
	for _, t := range tools {
		if err = reg.Register(t); err != nil {
			return fmt.Errorf("register tool %q: %w", t.Name(), err)
		}
	}

	srv := runtime.New(reg)
	return srv.Serve(context.Background())
}
