# CliExamples

Every example of the emissario cli. Each one is a shell session that runs with its own
directory as the working directory and writes only into its own `test-dir`, so it can be read
as documentation and copied line by line. The script types `emissario`, which `run-examples`
resolves to the code in this tree. It ends by copying out of `test-dir` into `assert-dir` the
paths it asserts — `mkdir -p assert-dir/<path>` then `cp -R test-dir/<path>/. assert-dir/<path>/`,
each keeping the place it holds in the tree.

`agnos run-examples` runs them all and checks each against the `result.yaml` beside it — the
golden holding the output, the exit code and the sha256 of every `assert-dir` file, written by
`run-examples` and never by hand. [Workflow](../Workflow/doc.md) has the commands that add and
remove one; the lib side is [LibExamples](../LibExamples/doc.md).

No example is declared yet: `examples/cli/` is created by the first
`agnos add-cli-example`.

