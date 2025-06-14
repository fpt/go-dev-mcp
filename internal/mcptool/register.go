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
	s.AddTool(tool, mcp.NewTypedToolHandler(getCurrentDateTime))

	// Add Make target tool
	tool = mcp.NewTool(
		"run_make",
		mcp.WithDescription(
			"Run make targets with intelligent output filtering"+
				" that highlights errors/warnings and reduces token usage by 85-95%."+
				" Automatically detects build failures, compilation errors, and test results.",
		),
		mcp.WithString("work_dir",
			mcp.DefaultString(workdir),
			mcp.Description("Working directory containing Makefile (absolute path)"),
		),
		mcp.WithString("target",
			mcp.DefaultString("help"),
			mcp.Description("Make target to run (e.g., 'build', 'test', 'clean', 'help')"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(runMakeTarget))

	// Add Tree Directory tool
	tool = mcp.NewTool(
		"tree_dir",
		mcp.WithDescription(
			"Display directory tree structure with depth limiting"+
				" (default: 4 levels) to reduce token usage by 60-80%."+
				" Perfect for project exploration and understanding codebase structure.",
		),
		mcp.WithString("root_dir",
			mcp.Required(),
			mcp.Description("Root directory to scan (absolute path)"),
		),
		mcp.WithBoolean(
			"ignore_dot",
			mcp.DefaultBool(false),
			mcp.Description(
				"Ignore dot files and directories (except .git which is always ignored)",
			),
		),
		mcp.WithNumber("max_depth",
			mcp.DefaultNumber(4),
			mcp.Description("Maximum directory depth to traverse (default: 4 levels)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(treeDir))

	// Add GoDoc search tool
	tool = mcp.NewTool("search_godoc",
		mcp.WithDescription("Search for Go package in pkg.go.dev"),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description(
				"Search text which occurs in the package name, package path, synopsis, or README",
			),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchGoDoc))

	// Add GoDoc read tool
	tool = mcp.NewTool(
		"read_godoc",
		mcp.WithDescription(
			"Read Go documentation with line-based paging and caching."+
				" Supports offset/limit for efficient exploration of large docs."+
				" Cached for 30min to speed up subsequent requests.",
		),
		mcp.WithString(
			"package_url",
			mcp.Required(),
			mcp.Description(
				"Go package URL (e.g., 'golang.org/x/net/html', 'github.com/user/repo')",
			),
		),
		mcp.WithNumber("offset",
			mcp.DefaultNumber(0),
			mcp.Description("Line number to start reading from (0-based)"),
		),
		mcp.WithNumber("limit",
			mcp.DefaultNumber(50),
			mcp.Description("Number of lines to read (default: 50)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(readGoDoc))

	// Add GitHub search code tool
	tool = mcp.NewTool(
		"search_github_code",
		mcp.WithDescription(
			"Search code in GitHub repositories with compact formatting"+
				" (30-40% token reduction)."+
				" Returns repository/path format instead of verbose labels for efficient scanning.",
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Code search query (e.g., 'function main', 'import context')"),
		),
		mcp.WithString("language",
			mcp.Required(), // Just for safety.
			mcp.DefaultString("go"),
			mcp.Enum("go", "go module", "yaml", "markdown", "makefile"),
			mcp.Description("Programming language to filter results (Default: 'go')"),
		),
		mcp.WithString("repo",
			mcp.Description("GitHub repository in 'owner/repo' format to limit search scope"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchCodeGitHub))

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
	s.AddTool(tool, mcp.NewTypedToolHandler(getGitHubContent))

	// Add GitHub tree tool
	tool = mcp.NewTool(
		"tree_github_repo",
		mcp.WithDescription(
			"Display GitHub repository tree structure with depth limiting"+
				" (default: 3 levels for network efficiency)."+
				" Conservative token usage for remote repository exploration.",
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository in 'owner/repo' format"),
		),
		mcp.WithString("path",
			mcp.DefaultString(""),
			mcp.Description("Path in the repository (defaults to root)"),
		),
		mcp.WithBoolean(
			"ignore_dot",
			mcp.DefaultBool(false),
			mcp.Description(
				"Ignore dot files and directories (except .git which is always ignored)",
			),
		),
		mcp.WithNumber(
			"max_depth",
			mcp.DefaultNumber(3),
			mcp.Description(
				"Maximum directory depth to traverse (default: 3 levels for network efficiency)",
			),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(getGitHubTree))

	// Add Local search tool
	tool = mcp.NewTool(
		"search_local_files",
		mcp.WithDescription(
			"Search file contents in local directories with match limiting"+
				" (default: 10 matches per file) to reduce token usage by 50-70%."+
				" Shows line numbers and truncation indicators.",
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Directory path to search in (absolute path)"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Text to search for in file contents (case-sensitive exact match)"),
		),
		mcp.WithString("extension",
			mcp.DefaultString("go"),
			mcp.Description("File extension to search (e.g., 'go', 'py', 'js', 'txt')"),
		),
		mcp.WithNumber("max_matches",
			mcp.DefaultNumber(10),
			mcp.Description("Maximum number of matches to show per file (default: 10)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchLocalFiles))

	return nil
}
