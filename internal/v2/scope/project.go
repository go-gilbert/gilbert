package scope

const (
	projectWorkDirKey      = "cwd"
	projectWorkspaceDirKey = "workspaceDir"
	projectWorkflowFileKey = "workflowFile"
)

type ProjectInfoOpts struct {
	WorkDir      string
	WorkspaceDir string
	WorkflowFile string
}

type ProjectInfo map[string]string

func NewProjectInfo(opts ProjectInfoOpts) ProjectInfo {
	return ProjectInfo{
		projectWorkDirKey:      opts.WorkDir,
		projectWorkspaceDirKey: opts.WorkspaceDir,
		projectWorkflowFileKey: opts.WorkflowFile,
	}
}

func (pi ProjectInfo) WorkDir() string {
	return pi[projectWorkDirKey]
}

func (pi ProjectInfo) WorkspaceDir() string {
	return pi[projectWorkspaceDirKey]
}

func (pi ProjectInfo) WorkflowFile() string {
	return pi[projectWorkflowFileKey]
}

func (pi ProjectInfo) Values() map[string]string {
	return pi
}
