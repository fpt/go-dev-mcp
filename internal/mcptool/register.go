package tool

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Register(s *server.MCPServer, workdir string) error {
	// Add datetime tool
	tool := mcp.NewTool("get_current_datetime",
		mcp.WithDescription("Get current date and time"),
	)
	s.AddTool(tool, getCurrentDateTime)

	// Add Make target tool
	tool = mcp.NewTool("run_make",
		mcp.WithDescription("Run make target"),
		mcp.WithString("project_dir",
			mcp.DefaultString(workdir),
			mcp.Description("Project root directory (NOTE: must be an absolute path)"),
		),
		mcp.WithString("target",
			mcp.DefaultString("help"),
			mcp.Description("Command to execute"),
		),
	)
	s.AddTool(tool, runMakeTarget)

	// Add Tree Directory tool
	tool = mcp.NewTool("tree_dir",
		mcp.WithDescription("List files in directory"),
		mcp.WithString("root_dir",
			mcp.Required(),
			mcp.Description("Root directory"),
		),
	)
	s.AddTool(tool, treeDir)

	// Add GoDoc search tool
	tool = mcp.NewTool("search_godoc",
		mcp.WithDescription("Search Go documentation"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
	)
	s.AddTool(tool, searchGoDoc)

	// Add GoDoc read tool
	tool = mcp.NewTool("read_godoc",
		mcp.WithDescription("Read Go documentation"),
		mcp.WithString("package_url",
			mcp.Required(),
			mcp.Description("URL of the Go package (e.g., 'github.com/user/repo')"),
		),
	)
	s.AddTool(tool, readGoDoc)

	return nil
}
