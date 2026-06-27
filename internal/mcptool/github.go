package tool

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fpt/go-dev-mcp/internal/app"
	"github.com/fpt/go-dev-mcp/internal/infra"
	"github.com/google/go-github/v74/github"
	"github.com/mark3labs/mcp-go/mcp"
)

// SearchCodeGitHubArgs represents arguments for GitHub code search
type SearchCodeGitHubArgs struct {
	Query    string `json:"query"`
	Language string `json:"language"`
	Repo     string `json:"repo"`
}

// GitHubContentArgs represents arguments for GitHub content retrieval
type GitHubContentArgs struct {
	Repo   string `json:"repo"`
	Path   string `json:"path"`
	Ref    string `json:"ref,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// GitHubTreeArgs represents arguments for GitHub tree display
type GitHubTreeArgs struct {
	Repo      string `json:"repo"`
	Path      string `json:"path"`
	IgnoreDot bool   `json:"ignore_dot"`
	MaxDepth  int    `json:"max_depth,omitempty"`
}

func searchCodeGitHub(
	ctx context.Context,
	request mcp.CallToolRequest,
	args SearchCodeGitHubArgs,
) (*mcp.CallToolResult, error) {
	if args.Query == "" {
		return mcp.NewToolResultError("Missing search query"), nil
	}
	if args.Language == "" {
		return mcp.NewToolResultError("Missing language"), nil
	}
	if args.Repo != "" {
		// Validate the repo format
		parts := strings.Split(args.Repo, "/")
		if len(parts) != 2 {
			return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
		}
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		slog.ErrorContext(ctx, "searchCodeGitHub", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	result, err := app.GitHubSearchCode(ctx, gh, args.Query, &args.Language, &args.Repo)
	if err != nil {
		slog.ErrorContext(ctx, "searchCodeGitHub", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error searching code: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}

func getGitHubContent(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubContentArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}

	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner := parts[0]
	repo := parts[1]

	if args.Path == "" {
		return mcp.NewToolResultError("Missing path"), nil
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubContent", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	content, err := gh.GetContent(ctx, owner, repo, args.Path, args.Ref)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubContent", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting content: %v", err)), nil
	}

	// Apply offset/limit if specified
	if args.Offset > 0 || args.Limit > 0 {
		content = paginateContent(content, args.Offset, args.Limit)
	}

	return mcp.NewToolResultText(content), nil
}

func getGitHubTree(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubTreeArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}

	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner := parts[0]
	repo := parts[1]

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	// Create a string builder for the tree output
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s/%s:%s\n", owner, repo, args.Path))

	// Set default max depth if not specified
	maxDepth := args.MaxDepth
	if maxDepth == 0 {
		maxDepth = 3 // More conservative default for GitHub trees due to network overhead
	}

	// Generate the tree using our PrintGitHubTree function
	if err := app.PrintGitHubTree(
		ctx,
		&b,
		gh,
		owner,
		repo,
		args.Path,
		args.IgnoreDot,
		maxDepth,
	); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error generating tree: %v", err)), nil
	}

	return mcp.NewToolResultText(b.String()), nil
}

// GitHubDiffArgs represents arguments for getting a diff between two refs
type GitHubDiffArgs struct {
	Repo string `json:"repo"`
	Base string `json:"base"`
	Head string `json:"head"`
}

func getGitHubDiff(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubDiffArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner, repo := parts[0], parts[1]
	if args.Base == "" || args.Head == "" {
		return mcp.NewToolResultError("Missing base or head ref"), nil
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	diff, err := gh.CompareRefs(ctx, owner, repo, args.Base, args.Head)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubDiff", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting diff: %v", err)), nil
	}

	return mcp.NewToolResultText(diff), nil
}

// GitHubListIssuesArgs represents arguments for listing issues
type GitHubListIssuesArgs struct {
	Repo   string `json:"repo"`
	State  string `json:"state,omitempty"`
	Labels string `json:"labels,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func listGitHubIssues(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubListIssuesArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner, repo := parts[0], parts[1]

	state := args.State
	if state == "" {
		state = "open"
	}
	limit := args.Limit
	if limit == 0 {
		limit = 30
	}

	var labels []string
	if args.Labels != "" {
		for _, l := range strings.Split(args.Labels, ",") {
			if l := strings.TrimSpace(l); l != "" {
				labels = append(labels, l)
			}
		}
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	issues, _, err := gh.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		State:       state,
		Labels:      labels,
		ListOptions: github.ListOptions{PerPage: min(limit, 100)},
	})
	if err != nil {
		slog.ErrorContext(ctx, "listGitHubIssues", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error listing issues: %v", err)), nil
	}

	var sb strings.Builder
	count := 0
	for _, issue := range issues {
		if issue.PullRequestLinks != nil {
			continue // skip PRs
		}
		labelNames := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			labelNames[i] = l.GetName()
		}
		labelsStr := ""
		if len(labelNames) > 0 {
			labelsStr = " [" + strings.Join(labelNames, ", ") + "]"
		}
		sb.WriteString(
			fmt.Sprintf(
				"#%d [%s]%s %s\n",
				issue.GetNumber(),
				issue.GetState(),
				labelsStr,
				issue.GetTitle(),
			),
		)
		count++
		if count >= limit {
			break
		}
	}
	if count == 0 {
		sb.WriteString("No issues found.\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// GitHubGetIssueArgs represents arguments for getting a single issue
type GitHubGetIssueArgs struct {
	Repo   string `json:"repo"`
	Number int    `json:"number"`
}

func getGitHubIssue(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubGetIssueArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner, repo := parts[0], parts[1]
	if args.Number == 0 {
		return mcp.NewToolResultError("Missing issue number"), nil
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	issue, _, err := gh.Issues.Get(ctx, owner, repo, args.Number)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubIssue", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting issue: %v", err)), nil
	}

	comments, _, err := gh.Issues.ListComments(
		ctx, owner, repo, args.Number,
		&github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}},
	)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubIssue:comments", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting comments: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(
		fmt.Sprintf("#%d [%s] %s\n", issue.GetNumber(), issue.GetState(), issue.GetTitle()),
	)
	sb.WriteString(
		fmt.Sprintf(
			"Author: %s | Created: %s\n",
			issue.GetUser().GetLogin(),
			issue.GetCreatedAt().String(),
		),
	)
	sb.WriteString(fmt.Sprintf("URL: %s\n\n", issue.GetHTMLURL()))
	sb.WriteString(issue.GetBody())
	sb.WriteString("\n")

	if len(comments) > 0 {
		sb.WriteString(fmt.Sprintf("\n--- %d comment(s) ---\n", len(comments)))
		for _, c := range comments {
			sb.WriteString(
				fmt.Sprintf(
					"\n@%s (%s):\n%s\n",
					c.GetUser().GetLogin(),
					c.GetCreatedAt().String(),
					c.GetBody(),
				),
			)
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// GitHubListPullsArgs represents arguments for listing pull requests
type GitHubListPullsArgs struct {
	Repo  string `json:"repo"`
	State string `json:"state,omitempty"`
	Base  string `json:"base,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

func listGitHubPulls(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubListPullsArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner, repo := parts[0], parts[1]

	state := args.State
	if state == "" {
		state = "open"
	}
	limit := args.Limit
	if limit == 0 {
		limit = 30
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	pulls, _, err := gh.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		State:       state,
		Base:        args.Base,
		ListOptions: github.ListOptions{PerPage: min(limit, 100)},
	})
	if err != nil {
		slog.ErrorContext(ctx, "listGitHubPulls", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error listing pull requests: %v", err)), nil
	}

	var sb strings.Builder
	for i, pr := range pulls {
		if i >= limit {
			break
		}
		draft := ""
		if pr.GetDraft() {
			draft = " [draft]"
		}
		sb.WriteString(fmt.Sprintf("#%d [%s]%s %s (%s←%s) by %s\n",
			pr.GetNumber(), pr.GetState(), draft, pr.GetTitle(),
			pr.GetBase().GetRef(), pr.GetHead().GetRef(), pr.GetUser().GetLogin()))
	}
	if len(pulls) == 0 {
		sb.WriteString("No pull requests found.\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// GitHubGetPullArgs represents arguments for getting a single pull request
type GitHubGetPullArgs struct {
	Repo   string `json:"repo"`
	Number int    `json:"number"`
}

func getGitHubPull(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubGetPullArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner, repo := parts[0], parts[1]
	if args.Number == 0 {
		return mcp.NewToolResultError("Missing pull request number"), nil
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	pr, _, err := gh.PullRequests.Get(ctx, owner, repo, args.Number)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubPull", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting pull request: %v", err)), nil
	}

	files, _, err := gh.PullRequests.ListFiles(
		ctx, owner, repo, args.Number,
		&github.ListOptions{PerPage: 100},
	)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubPull:files", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting changed files: %v", err)), nil
	}

	var sb strings.Builder
	draft := ""
	if pr.GetDraft() {
		draft = " [draft]"
	}
	sb.WriteString(
		fmt.Sprintf("#%d [%s]%s %s\n", pr.GetNumber(), pr.GetState(), draft, pr.GetTitle()),
	)
	sb.WriteString(
		fmt.Sprintf(
			"Author: %s | %s←%s\n",
			pr.GetUser().GetLogin(),
			pr.GetBase().GetRef(),
			pr.GetHead().GetRef(),
		),
	)
	sb.WriteString(
		fmt.Sprintf("Created: %s | URL: %s\n", pr.GetCreatedAt().String(), pr.GetHTMLURL()),
	)
	sb.WriteString(
		fmt.Sprintf(
			"Changes: +%d -%d across %d file(s)\n\n",
			pr.GetAdditions(),
			pr.GetDeletions(),
			pr.GetChangedFiles(),
		),
	)
	sb.WriteString(pr.GetBody())
	sb.WriteString("\n")

	if len(files) > 0 {
		sb.WriteString("\n--- Changed files ---\n")
		for _, f := range files {
			sb.WriteString(
				fmt.Sprintf(
					"%s %s (+%d -%d)\n",
					f.GetStatus(),
					f.GetFilename(),
					f.GetAdditions(),
					f.GetDeletions(),
				),
			)
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// GitHubListWorkflowRunsArgs represents arguments for listing workflow runs
type GitHubListWorkflowRunsArgs struct {
	Repo   string `json:"repo"`
	Branch string `json:"branch,omitempty"`
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func listGitHubWorkflowRuns(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubListWorkflowRunsArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner, repo := parts[0], parts[1]

	limit := args.Limit
	if limit == 0 {
		limit = 20
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	runs, _, err := gh.Actions.ListRepositoryWorkflowRuns(
		ctx,
		owner,
		repo,
		&github.ListWorkflowRunsOptions{
			Branch:      args.Branch,
			Status:      args.Status,
			ListOptions: github.ListOptions{PerPage: min(limit, 100)},
		},
	)
	if err != nil {
		slog.ErrorContext(ctx, "listGitHubWorkflowRuns", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error listing workflow runs: %v", err)), nil
	}

	var sb strings.Builder
	for i, run := range runs.WorkflowRuns {
		if i >= limit {
			break
		}
		conclusion := run.GetConclusion()
		if conclusion == "" {
			conclusion = run.GetStatus()
		}
		sb.WriteString(fmt.Sprintf(
			"run#%d [%s] %s on %s (%s) sha:%s\n",
			run.GetRunNumber(),
			conclusion,
			run.GetName(),
			run.GetHeadBranch(),
			run.GetEvent(),
			run.GetHeadSHA()[:min(7, len(run.GetHeadSHA()))],
		))
	}
	if runs.GetTotalCount() == 0 {
		sb.WriteString("No workflow runs found.\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// GitHubGetWorkflowRunArgs represents arguments for getting a workflow run
type GitHubGetWorkflowRunArgs struct {
	Repo  string `json:"repo"`
	RunID int64  `json:"run_id"`
}

func getGitHubWorkflowRun(
	ctx context.Context,
	request mcp.CallToolRequest,
	args GitHubGetWorkflowRunArgs,
) (*mcp.CallToolResult, error) {
	if args.Repo == "" {
		return mcp.NewToolResultError("Missing repo"), nil
	}
	parts := strings.Split(args.Repo, "/")
	if len(parts) != 2 {
		return mcp.NewToolResultError("Invalid repo format, expected 'owner/repo'"), nil
	}
	owner, repo := parts[0], parts[1]
	if args.RunID == 0 {
		return mcp.NewToolResultError("Missing run_id"), nil
	}

	gh, err := infra.NewGitHubClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating GitHub client: %v", err)), nil
	}

	run, _, err := gh.Actions.GetWorkflowRunByID(ctx, owner, repo, args.RunID)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubWorkflowRun", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting workflow run: %v", err)), nil
	}

	jobs, _, err := gh.Actions.ListWorkflowJobs(
		ctx, owner, repo, args.RunID,
		&github.ListWorkflowJobsOptions{
			Filter:      "latest",
			ListOptions: github.ListOptions{PerPage: 100},
		},
	)
	if err != nil {
		slog.ErrorContext(ctx, "getGitHubWorkflowRun:jobs", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Error getting workflow jobs: %v", err)), nil
	}

	var sb strings.Builder
	conclusion := run.GetConclusion()
	if conclusion == "" {
		conclusion = run.GetStatus()
	}
	sb.WriteString(fmt.Sprintf("run#%d [%s] %s\n", run.GetRunNumber(), conclusion, run.GetName()))
	sb.WriteString(
		fmt.Sprintf(
			"Branch: %s | Event: %s | SHA: %s\n",
			run.GetHeadBranch(),
			run.GetEvent(),
			run.GetHeadSHA(),
		),
	)
	sb.WriteString(
		fmt.Sprintf("Created: %s | URL: %s\n", run.GetCreatedAt().String(), run.GetHTMLURL()),
	)

	if len(jobs.Jobs) > 0 {
		sb.WriteString("\n--- Jobs ---\n")
		for _, job := range jobs.Jobs {
			jobConclusion := job.GetConclusion()
			if jobConclusion == "" {
				jobConclusion = job.GetStatus()
			}
			sb.WriteString(fmt.Sprintf("\n%s [%s]\n", job.GetName(), jobConclusion))
			for _, step := range job.Steps {
				stepConclusion := step.GetConclusion()
				if stepConclusion == "" {
					stepConclusion = step.GetStatus()
				}
				sb.WriteString(
					fmt.Sprintf(
						"  %d. %s [%s]\n",
						step.GetNumber(),
						step.GetName(),
						stepConclusion,
					),
				)
			}
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// paginateContent applies offset/limit to content by lines, similar to how read_godoc works
func paginateContent(content string, offset, limit int) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Apply offset
	if offset >= totalLines {
		return fmt.Sprintf("(offset %d exceeds file length of %d lines)", offset, totalLines)
	}
	if offset > 0 {
		lines = lines[offset:]
	}

	// Apply limit (0 means no limit)
	if limit > 0 && limit < len(lines) {
		lines = lines[:limit]
		// Add truncation indicator
		lines = append(lines, fmt.Sprintf("... (showing lines %d-%d of %d total lines)",
			offset+1, offset+limit, totalLines))
	}

	return strings.Join(lines, "\n")
}
