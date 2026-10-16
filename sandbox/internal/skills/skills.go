package skills

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
)

// A skill is a folder of Markdown that teaches a coding agent — Claude Code,
// Codex, Antigravity — the side of emissario one repository works with: the
// website emissario serves (drop-front-skills) or the instructions its agent
// reads (drop-agents-skills). Each one is embedded under
// assets/skills/<kind>/<name>/, rendered as a text/template from Vars, and
// written into the folder every agent reads its skills from, in the Agent
// Skills layout: <name>/SKILL.md beside its references/. A skill holds the
// public interface that repository needs and nothing else: what its pages
// send and receive, or what the model is given; never how the server does it.

const (
	// KindFront is the skills of the website repository.
	KindFront = "front"
	// KindAgents is the skills of the agent instructions repository.
	KindAgents = "agents"
	// assetsDir is the folder of the embedded assets every kind lives under.
	assetsDir = "skills"
)

// Targets are the folders, relative to the repository, the agents read the
// skills of a project from: .claude/skills for Claude Code, which reads no
// other, and .agents/skills, the shared one Codex and Antigravity read. The
// same skill is written into each.
var Targets = []string{".claude/skills", ".agents/skills"}

// Vars is what a skill template reads: the limits it quotes, taken from the
// code that enforces them, and the mandatory prompts of the agent, so a skill
// never drifts from the server it describes.
type Vars struct {
	ContactPath           string
	ContactMaxBytes       int
	ContactMaxMaxBytes    int
	ContactRatePerMinute  int
	AgentMaxMessages      int
	AgentMaxContentLength int
	AgentMaxBodyKiB       int
	AgentMaxToolRounds    int
	AgentDeadlineSeconds  int
	AgentInstructionKiB   int
	AgentInstructionFiles int
	AgentUrlResponseKiB   int
	AgentRatePerMinute    int
	InstructionsFolder    string
	ReadmeFile            string
	SummaryFile           string
	SystemPrompt          string
	SummaryPrompt         string
}

// file is one rendered file of a skill, its path relative to the kind:
// <name>/SKILL.md, <name>/references/….
type file struct {
	path    string
	content []byte
}

// Drop writes the skills of kind into the repository at root, under every
// folder of Targets, and answers the path of each file written, relative to
// root, in the order written. Every file is rendered before anything is
// touched; then the folder of each skill is removed and written again, so a
// second run changes nothing and a newer emissario leaves no stale file.
// Nothing else under root is touched.
func Drop(sandbox *api.Sandbox, kind string, root string) ([]string, error) {
	files, names, err := render(sandbox, kind)
	if err != nil {
		return nil, err
	}
	io := sandbox.Deps.IoDeps
	written := []string{}
	for _, target := range Targets {
		for _, name := range names {
			io.RemoveDir(io.Join(root, target, name))
		}
		for _, rendered := range files {
			relative := target + "/" + rendered.path
			if err := io.WriteFile(io.Join(root, relative), rendered.content); err != nil {
				return written, sandbox.Deps.StdDeps.Errorf("%s could not be written: %s", relative, err.Error())
			}
			written = append(written, relative)
		}
	}
	return written, nil
}

// render answers every file of kind, rendered, and the names of its skills —
// the first segment of their paths — in order.
func render(sandbox *api.Sandbox, kind string) ([]file, []string, error) {
	vars, err := templateVars(sandbox)
	if err != nil {
		return nil, nil, err
	}
	dir := assetsDir + "/" + kind
	paths, err := sandbox.Deps.EmbedDeps.ListFilesRecursively(dir)
	if err != nil || len(paths) == 0 {
		return nil, nil, sandbox.Deps.StdDeps.Errorf("no skill of kind %s is embedded", kind)
	}
	files := []file{}
	names := []string{}
	seen := map[string]bool{}
	for _, path := range paths {
		content, err := sandbox.Deps.EmbedDeps.RenderTemplate(dir+"/"+path, vars)
		if err != nil {
			return nil, nil, sandbox.Deps.StdDeps.Errorf("the skill file %s could not be rendered: %s", path, err.Error())
		}
		files = append(files, file{path: path, content: content})
		name := sandbox.Deps.StringsDeps.Split(path, "/")[0]
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return files, names, nil
}

// templateVars fills Vars from the constants of the agent and the contact
// requests, and the embedded mandatory prompts of the agent.
func templateVars(sandbox *api.Sandbox) (Vars, error) {
	system, err := sandbox.Deps.EmbedDeps.ReadFile(agent.SystemPromptAsset)
	if err != nil {
		return Vars{}, sandbox.Deps.StdDeps.Errorf("the embedded %s could not be read: %s", agent.SystemPromptAsset, err.Error())
	}
	summary, err := sandbox.Deps.EmbedDeps.ReadFile(agent.SummaryPromptAsset)
	if err != nil {
		return Vars{}, sandbox.Deps.StdDeps.Errorf("the embedded %s could not be read: %s", agent.SummaryPromptAsset, err.Error())
	}
	return Vars{
		ContactPath:           contact.PublicPath,
		ContactMaxBytes:       contact.DefaultMaxBytes,
		ContactMaxMaxBytes:    contact.MaxMaxBytes,
		ContactRatePerMinute:  contact.DefaultRatePerMinute,
		AgentMaxMessages:      agent.MaxMessages,
		AgentMaxContentLength: agent.MaxContentLength,
		AgentMaxBodyKiB:       agent.MaxBodyBytes / 1024,
		AgentMaxToolRounds:    agent.MaxToolRounds,
		AgentDeadlineSeconds:  agent.DeadlineMilliseconds / 1000,
		AgentInstructionKiB:   agent.MaxInstructionBytes / 1024,
		AgentInstructionFiles: agent.MaxInstructionFiles,
		AgentUrlResponseKiB:   agent.MaxUrlResponseBytes / 1024,
		AgentRatePerMinute:    agent.DefaultRatePerMinute,
		InstructionsFolder:    agent.InstructionsFolder,
		ReadmeFile:            agent.ReadmeFile,
		SummaryFile:           agent.SummaryFile,
		SystemPrompt:          sandbox.Deps.StringsDeps.TrimSpace(string(system)),
		SummaryPrompt:         sandbox.Deps.StringsDeps.TrimSpace(string(summary)),
	}, nil
}
