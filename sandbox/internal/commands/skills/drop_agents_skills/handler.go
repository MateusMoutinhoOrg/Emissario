package drop_agents_skills

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/skills"
)

// Handle answers `drop-agents-skills`: the emissario-agent skill is written
// into the agent instructions repository at --path, once for each agent of
// skills.Targets, and the path of every file written is printed, one a line.
// A repository with no instructions/README.md is still written to, with a
// warning on stderr: until it holds one, the agent answers 503
// no-instructions. It opens no database.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	io := sandbox.Deps.IoDeps
	if !io.IsDir(input.Path) {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitUsage, "path", input.Path+" is not a directory")
	}
	written, err := skills.Drop(sandbox, skills.KindAgents, input.Path)
	for _, path := range written {
		response.Printf("%s\n", path)
	}
	if err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "failed to write the skills: "+err.Error())
	}
	if !io.IsFile(io.Join(input.Path, agent.InstructionsFolder, agent.ReadmeFile)) {
		sandbox.Deps.StdDeps.Eprintf("warning: %s holds no %s/%s: until it does, the agent answers 503 no-instructions\n", input.Path, agent.InstructionsFolder, agent.ReadmeFile)
	}
	return nil
}
