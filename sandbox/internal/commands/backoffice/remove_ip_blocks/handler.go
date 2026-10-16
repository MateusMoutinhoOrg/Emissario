package remove_ip_blocks

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/blocker"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
)

// Handle answers `remove-ip-blocks`: the block of --value, or every one with
// --all, is removed. It is the way back in for a root whose network a block
// turns away from /admin/login. A server running beside it goes by the change
// within the seconds its blocks are cached.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	if input.All == (input.Value != "") {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitUsage, "value", "give either --value or --all")
	}
	if input.All {
		removed, err := blocker.RemoveAll(sandbox)
		if err != nil {
			return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "failed to remove the ip blocks: "+err.Error())
		}
		response.Printf("%d ip blocks removed\n", removed)
		return nil
	}
	removed, refusal, err := blocker.RemoveValue(sandbox, input.Value)
	if err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "failed to remove the ip block: "+err.Error())
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitUsage, "value", refusal.Message)
	}
	if !removed {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "value", "no ip block holds "+input.Value)
	}
	response.Printf("ip block %s removed\n", input.Value)
	return nil
}
