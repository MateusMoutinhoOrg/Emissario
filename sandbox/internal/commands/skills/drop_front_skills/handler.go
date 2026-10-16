package drop_front_skills

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/skills"
)

// Handle answers `drop-front-skills`: the emissario-website skill is written
// into the website repository at --path, once for each agent of
// skills.Targets, and the path of every file written is printed, one a line.
// A repository with no frontend/ folder is still written to, with a warning on
// stderr: emissario serves that folder alone. It opens no database.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	io := sandbox.Deps.IoDeps
	if !io.IsDir(input.Path) {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitUsage, "path", input.Path+" is not a directory")
	}
	written, err := skills.Drop(sandbox, skills.KindFront, input.Path)
	for _, path := range written {
		response.Printf("%s\n", path)
	}
	if err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "failed to write the skills: "+err.Error())
	}
	if !io.IsDir(io.Join(input.Path, "frontend")) {
		sandbox.Deps.StdDeps.Eprintf("warning: %s holds no frontend/ folder: emissario serves the frontend/ folder of the website repository, and nothing else of it\n", input.Path)
	}
	return nil
}
