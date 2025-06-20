package subcmd

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	tool "github.com/fpt/go-dev-mcp/internal/mcptool"
	"github.com/mark3labs/mcp-go/server"

	"github.com/google/subcommands"
)

const DefaultSSEServerAddr = "localhost:5000"

type ServeCmd struct {
	workdir string
	addr    string
	sse     bool
	debug   bool
	logFile string
}

func (*ServeCmd) Name() string     { return "serve" }
func (*ServeCmd) Synopsis() string { return "Serve files over the server." }
func (*ServeCmd) Usage() string {
	return `serve [flags]:
  Serve files over the server.
`
}

func (p *ServeCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.sse, "sse", false, "Use SSE")
	f.StringVar(&p.workdir, "workdir", ".", "Working directory")
	f.StringVar(&p.addr, "addr", DefaultSSEServerAddr, "SSE server address")
	f.BoolVar(&p.debug, "debug", os.Getenv("DEBUG") != "", "Enable debug mode")
	f.StringVar(&p.logFile, "logfile", os.Getenv("LOGFILE"), "Log file path")
}

func (p *ServeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	s := server.NewMCPServer(
		"go-dev-mcp server 🚀",
		"1.0.0",
	)

	ctx := context.Background()
	if p.debug && p.logFile != "" {
		f, err := os.OpenFile(p.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			slog.ErrorContext(ctx, "Error opening log file", "error", err)
			return subcommands.ExitFailure
		}
		defer f.Close()
		slog.SetDefault(slog.New(slog.NewTextHandler(f, nil)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	}

	if err := tool.Register(s, p.workdir); err != nil {
		slog.ErrorContext(ctx, "Error registering tools", "error", err)
		return subcommands.ExitFailure
	}

	if p.sse {
		// Start the SSE server
		ss := server.NewSSEServer(s)
		defer func() {
			if err := ss.Shutdown(ctx); err != nil {
				slog.ErrorContext(ctx, "Error shutting down SSE server", "error", err)
			}
		}()
		fmt.Printf("Starting SSE server on %s\n", p.addr)
		if err := ss.Start(p.addr); err != nil {
			slog.ErrorContext(ctx, "SSE server error", "error", err)
		}
	} else {
		// Start the stdio server
		if err := server.ServeStdio(s); err != nil {
			slog.ErrorContext(ctx, "Server error", "error", err)
		}
	}

	return subcommands.ExitSuccess
}
