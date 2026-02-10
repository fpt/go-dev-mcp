package tool

import (
	"fmt"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Register(s *server.MCPServer, workdir string) error {
	// Add Tree Directory tool
	tool := mcp.NewTool(
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
		mcp.WithNumber(
			"limit",
			mcp.DefaultNumber(app.DefaultLinesPerPage),
			mcp.Description(
				fmt.Sprintf("Number of lines to read (default: %d)", app.DefaultLinesPerPage),
			),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(readGoDoc))

	// Add GoDoc search within documentation tool
	tool = mcp.NewTool(
		"search_within_godoc",
		mcp.WithDescription(
			"Search for keywords within a specific Go package documentation."+
				" Returns all matching lines with line numbers, similar to search_local_files.",
		),
		mcp.WithString(
			"package_url",
			mcp.Required(),
			mcp.Description(
				"Go package URL (e.g., 'golang.org/x/net/html', 'github.com/user/repo')",
			),
		),
		mcp.WithString(
			"keyword",
			mcp.Required(),
			mcp.Description("Keyword to search for within the documentation"),
		),
		mcp.WithNumber("max_matches",
			mcp.DefaultNumber(10),
			mcp.Description("Maximum number of matches to return (default: 10)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchWithinGoDoc))

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
			mcp.Required(),
			mcp.Description(
				"Programming language to filter results"+
					" (e.g., 'go', 'python', 'rust', 'typescript', 'c++', 'arduino', 'yaml', 'makefile')",
			),
		),
		mcp.WithString("repo",
			mcp.Description("GitHub repository in 'owner/repo' format to limit search scope"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchCodeGitHub))

	// Add GitHub get content tool
	tool = mcp.NewTool("get_github_content",
		mcp.WithDescription("Get content from GitHub with line-based paging for large files"),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("GitHub repository in 'owner/repo' format"),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to the file in the repository"),
		),
		mcp.WithNumber("offset",
			mcp.DefaultNumber(0),
			mcp.Description("Line number to start reading from (0-based)"),
		),
		mcp.WithNumber("limit",
			mcp.DefaultNumber(100),
			mcp.Description("Number of lines to read (default: 100, 0 for all lines)"),
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
			mcp.Required(),
			mcp.Description(
				"File extension to search without dot"+
					" (e.g., 'go', 'py', 'rs', 'ts', 'js', 'ino', 'c', 'h', 'txt', 'yaml')",
			),
		),
		mcp.WithNumber("max_matches",
			mcp.DefaultNumber(10),
			mcp.Description("Maximum number of matches to show per file (default: 10)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchLocalFiles))

	// Add Go Package Outline tool
	tool = mcp.NewTool(
		"outline_go_package",
		mcp.WithDescription(
			"Get a comprehensive outline of a Go package:"+
				" dependencies, exported declarations, and call graph."+
				" Analyzes all Go source files in the specified directory."+
				" Use skip_* flags to omit sections and reduce output size.",
		),
		mcp.WithString("directory",
			mcp.Required(),
			mcp.Description("Directory containing Go source files to analyze"),
		),
		mcp.WithBoolean("skip_dependencies",
			mcp.DefaultBool(false),
			mcp.Description("Skip the dependencies section"),
		),
		mcp.WithBoolean("skip_declarations",
			mcp.DefaultBool(false),
			mcp.Description("Skip the declarations section"),
		),
		mcp.WithBoolean("skip_call_graph",
			mcp.DefaultBool(false),
			mcp.Description("Skip the call graph section (largest section)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(outlineGoPackage))

	// Add Scan Markdown tool
	tool = mcp.NewTool(
		"scan_markdown",
		mcp.WithDescription(
			"Scan markdown files in a directory or analyze a single markdown file to extract headings with line numbers. "+
				"Uses goldmark parser to accurately extract heading hierarchy and positions.",
		),
		mcp.WithString(
			"path",
			mcp.Required(),
			mcp.Description(
				"Directory path to scan for markdown files or path to a single markdown file (absolute path)",
			),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(scanMarkdown))

	// Add Go Code Validation tool
	tool = mcp.NewTool(
		"validate_go_code",
		mcp.WithDescription(
			"Validate Go code using multiple static analysis tools including go vet, build checks, "+
				"formatting validation, and module tidiness. Provides comprehensive code quality assessment.",
		),
		mcp.WithString(
			"directory",
			mcp.Required(),
			mcp.Description("Directory containing Go code to validate (absolute path)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(validateGoCode))

	// Add Rust documentation search tool
	tool = mcp.NewTool("search_rustdoc",
		mcp.WithDescription("Search for Rust crates on docs.rs"),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description("Search text for crate name or description"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchRustDoc))

	// Add Rust documentation read tool
	tool = mcp.NewTool(
		"read_rustdoc",
		mcp.WithDescription(
			"Read Rust crate documentation from docs.rs with line-based paging and caching."+
				" Supports offset/limit for efficient exploration of large docs."+
				" Cached for 30min to speed up subsequent requests.",
		),
		mcp.WithString(
			"crate_url",
			mcp.Required(),
			mcp.Description(
				"Crate name or path (e.g., 'serde', 'serde/de', 'tokio/runtime')",
			),
		),
		mcp.WithNumber("offset",
			mcp.DefaultNumber(0),
			mcp.Description("Line number to start reading from (0-based)"),
		),
		mcp.WithNumber(
			"limit",
			mcp.DefaultNumber(app.DefaultLinesPerPage),
			mcp.Description(
				fmt.Sprintf("Number of lines to read (default: %d)", app.DefaultLinesPerPage),
			),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(readRustDoc))

	// Add Rust documentation search within tool
	tool = mcp.NewTool(
		"search_within_rustdoc",
		mcp.WithDescription(
			"Search for keywords within a specific Rust crate documentation."+
				" Returns all matching lines with line numbers, similar to search_local_files.",
		),
		mcp.WithString(
			"crate_url",
			mcp.Required(),
			mcp.Description(
				"Crate name or path (e.g., 'serde', 'tokio/runtime')",
			),
		),
		mcp.WithString(
			"keyword",
			mcp.Required(),
			mcp.Description("Keyword to search for within the documentation"),
		),
		mcp.WithNumber("max_matches",
			mcp.DefaultNumber(10),
			mcp.Description("Maximum number of matches to return (default: 10)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchWithinRustDoc))

	// Add Python documentation search tool
	tool = mcp.NewTool("search_pydoc",
		mcp.WithDescription(
			"Search Python standard library modules on docs.python.org."+
				" Searches the module index by name and description.",
		),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description("Search text to match against module names and descriptions"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchPyDoc))

	// Add Python documentation read tool
	tool = mcp.NewTool(
		"read_pydoc",
		mcp.WithDescription(
			"Read Python standard library module documentation from docs.python.org"+
				" with line-based paging and caching."+
				" Supports offset/limit for efficient exploration of large docs."+
				" Cached for 30min to speed up subsequent requests.",
		),
		mcp.WithString(
			"module_name",
			mcp.Required(),
			mcp.Description(
				"Python module name as used in imports (e.g., 'json', 'os.path', 'collections.abc')",
			),
		),
		mcp.WithNumber("offset",
			mcp.DefaultNumber(0),
			mcp.Description("Line number to start reading from (0-based)"),
		),
		mcp.WithNumber(
			"limit",
			mcp.DefaultNumber(app.DefaultLinesPerPage),
			mcp.Description(
				fmt.Sprintf("Number of lines to read (default: %d)", app.DefaultLinesPerPage),
			),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(readPyDoc))

	// Add Python documentation search within tool
	tool = mcp.NewTool(
		"search_within_pydoc",
		mcp.WithDescription(
			"Search for keywords within a specific Python module documentation."+
				" Returns all matching lines with line numbers, similar to search_local_files.",
		),
		mcp.WithString(
			"module_name",
			mcp.Required(),
			mcp.Description(
				"Python module name (e.g., 'json', 'os.path')",
			),
		),
		mcp.WithString(
			"keyword",
			mcp.Required(),
			mcp.Description("Keyword to search for within the documentation"),
		),
		mcp.WithNumber("max_matches",
			mcp.DefaultNumber(10),
			mcp.Description("Maximum number of matches to return (default: 10)"),
		),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchWithinPyDoc))

	return nil
}
