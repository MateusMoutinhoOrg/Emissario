# Every database is encrypted at rest by Keep itself, under an EncryptKey
# derived from EMISSARIO_ENCRYPT_KEY, or kept in clear with --notencrypt for
# local debugging. This example walks every row of the start-up table of
# docs/Privacy against one --database folder, and checks that no value is
# left in clear on the disk.
#
# The keys are fixed so the output is the same at every run; a real key comes
# from `emissario generate-encrypt-key` and lives in a secret manager.

KEY=000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f
OTHER=ICEiIyQlJicoKSorLC0uLzAxMjM0NTY3ODk6Ozw9Pj8=

mkdir -p test-dir
cd test-dir
rm -rf data clear

run() {
  "$@" > /dev/null 2> err.txt
  echo "exit $? $(cat err.txt)"
}

echo "== an empty folder and a key: encrypted, encryption.json written"
run env EMISSARIO_ENCRYPT_KEY=$KEY emissario add-backoffice-user --username admin --email admin@example.com --role root
cat data/encryption.json

echo "== no key and no --notencrypt: refused"
run emissario add-backoffice-user --username a --email a@example.com

echo "== the key and --notencrypt: refused"
run env EMISSARIO_ENCRYPT_KEY=$KEY emissario add-backoffice-user --username a --email a@example.com --notencrypt

echo "== an encrypted folder and its key: runs"
run env EMISSARIO_ENCRYPT_KEY=$KEY emissario add-backoffice-user --username viewer --email viewer@example.com

echo "== an encrypted folder and another key: refused"
run env EMISSARIO_ENCRYPT_KEY=$OTHER emissario add-backoffice-user --username a --email a@example.com

echo "== an encrypted folder and --notencrypt: refused"
run emissario add-backoffice-user --username a --email a@example.com --notencrypt

echo "== a folder of emissario v0.0.6, sealed with XChaCha20-Poly1305: refused"
mkdir -p v1/backofficedb
echo '{"algorithm":"XChaCha20-Poly1305","key_id":"e395e1f8","version":1}' > v1/encryption.json
run env EMISSARIO_ENCRYPT_KEY=$KEY emissario add-backoffice-user --username a --email a@example.com --database v1

echo "== a passphrase: refused, never quoted"
run env EMISSARIO_ENCRYPT_KEY="correct horse battery staple" emissario add-backoffice-user --username a --email a@example.com --database other

echo "== an empty folder and --notencrypt: in clear, with a warning"
run emissario add-backoffice-user --username a --email a@example.com --database clear --notencrypt

echo "== a folder in clear and a key: refused"
run env EMISSARIO_ENCRYPT_KEY=$KEY emissario add-backoffice-user --username b --email b@example.com --database clear

echo "== help, version and generate-encrypt-key open no database"
emissario version > /dev/null && echo "version runs without a key"
emissario generate-encrypt-key 2> /dev/null | wc -c | tr -d ' '

echo "== nothing in clear under data/"
grep -rl -e admin@example.com -e viewer@example.com data || echo "no email found in data/"
echo "files shorter than the 28 bytes of nonce and tag AES-256-GCM adds to every value: $(find data -type f ! -name encryption.json -size -28c | wc -l | tr -d ' ')"

cd ..
mkdir -p assert-dir
cp test-dir/data/encryption.json assert-dir/
