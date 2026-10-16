# drop-front-skills writes the emissario-website skill into a website
# repository, and drop-agents-skills the emissario-agent skill into an agent
# instructions repository, once for each coding agent: .claude/skills/ (Claude
# Code), .agents/skills/ (Codex) and .agent/skills/ (Antigravity). This example
# drops both into two repositories, twice, and checks that a run opens no
# database, needs no key, rewrites only its own folders and refuses a --path
# that is not a directory.

mkdir -p test-dir
cd test-dir
rm -rf site agent

mkdir -p site/frontend site/.claude/skills/my-own-skill
echo '<h1>Hello</h1>' > site/frontend/index.html
echo 'a skill of the repository itself' > site/.claude/skills/my-own-skill/SKILL.md

mkdir -p agent/instructions
echo '# Assistant of Example' > agent/instructions/README.md

echo "== the website repository: one skill per agent"
env -u EMISSARIO_ENCRYPT_KEY emissario drop-front-skills --path site
echo "exit $?"

echo "== the agent repository"
env -u EMISSARIO_ENCRYPT_KEY emissario drop-agents-skills --path agent
echo "exit $?"

echo "== a second run changes nothing but a file it did not write, which it removes"
echo 'stale' > site/.agents/skills/emissario-website/stale.md
rm -rf before && mkdir before && cp -R site agent before/
emissario drop-front-skills --path site > /dev/null
emissario drop-agents-skills --path agent > /dev/null
diff -r before/site site
diff -r before/agent agent && echo "agent unchanged"
rm -rf before

echo "== a skill of the repository itself is left as it was"
cat site/.claude/skills/my-own-skill/SKILL.md

echo "== a repository that does not look like one: written, with a warning"
mkdir -p empty
emissario drop-agents-skills --path empty 2>&1 > /dev/null
echo "exit $?"

echo "== a --path that is not a directory: refused"
emissario drop-front-skills --path missing > /dev/null 2> err.txt
echo "exit $? $(cat err.txt)"

echo "== no database was opened"
test -e data && echo "data/ created" || echo "no data/ folder"

cd ..
mkdir -p assert-dir/site assert-dir/agent
cp -R test-dir/site/. assert-dir/site/
cp -R test-dir/agent/. assert-dir/agent/
