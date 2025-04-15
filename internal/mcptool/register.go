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
		mcp.WithDescription("Search for Go package in pkg.go.dev"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search text which occurs in the package name, package path, synopsis, or README"),
		),
	)
	s.AddTool(tool, searchGoDoc)

	// Add GoDoc read tool
	tool = mcp.NewTool("read_godoc",
		mcp.WithDescription("Read Go documentation of given package from pkg.go.dev"),
		mcp.WithString("package_url",
			mcp.Required(),
			mcp.Description("URL of the Go package (e.g., 'github.com/user/repo')"),
		),
	)
	s.AddTool(tool, readGoDoc)

	// Add GitHub search code tool
	tool = mcp.NewTool("search_github_code",
		mcp.WithDescription("Search code in GitHub"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithString("language",
			mcp.Required(), // Just for safety.
			mcp.DefaultString("go"),
			mcp.Enum("go", "go module", "yaml", "markdown", "makefile"),
			mcp.Description("Programming language to filter results (Default: 'go')"),
		),
		mcp.WithString("repo",
			mcp.Description("GitHub repository in 'owner/repo' format"),
		),
	)
	s.AddTool(tool, searchCodeGitHub)

	// Add GitHub get content tool
	tool = mcp.NewTool("get_github_content",
		mcp.WithDescription("Get content from GitHub"),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository in 'owner/repo' format"),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file in the repository"),
		),
	)
	s.AddTool(tool, getGitHubContent)

	// Add GitHub tree tool
	tool = mcp.NewTool("tree_github_repo",
		mcp.WithDescription("Display tree structure of a GitHub repository"),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository in 'owner/repo' format"),
		),
		mcp.WithString("path",
			mcp.DefaultString(""),
			mcp.Description("Path in the repository (defaults to root)"),
		),
	)
	s.AddTool(tool, getGitHubTree)

	return nil
}
