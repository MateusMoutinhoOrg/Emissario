# rotate-encrypt-key moves every database from the key in the variable
# --old-key-env names to the one in EMISSARIO_ENCRYPT_KEY: it takes a
# snapshot, reseals every file Keep wrote, removes every index entry — its
# name is a digest keyed by the old key — and rebuilds every index under the
# new one. This example rotates one --database folder, checks that only the
# new key opens it, and resumes a rotation left unfinished.
#
# The keys are fixed so the output is the same at every run; a real key comes
# from `emissario generate-encrypt-key` and lives in a secret manager.

OLD=000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f
NEW=ICEiIyQlJicoKSorLC0uLzAxMjM0NTY3ODk6Ozw9Pj8=
THIRD=QEFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaW1xdXl8=

mkdir -p test-dir
cd test-dir
rm -rf data third

# run prints what a command answered and its exit status, the instant in the
# name of the snapshot a rotation takes spelled out.
run() {
  "$@" > out.txt 2>&1
  status=$?
  sed 's/pre-rotation-[0-9]*-[0-9]*/pre-rotation-<instant>/' out.txt
  echo "exit $status"
}

# keyid prints the key_id the encryption.json of a folder records.
keyid() {
  sed -n 's/.*"key_id":"\([0-9a-f]*\)".*/\1/p' "$1/encryption.json"
}

echo "== a folder encrypted with the old key"
env EMISSARIO_ENCRYPT_KEY=$OLD emissario add-backoffice-user --username admin --email admin@example.com --role root > /dev/null
echo "exit $?"
cat data/encryption.json

echo "== the same key twice: refused"
run env OLD_KEY=$OLD EMISSARIO_ENCRYPT_KEY=$OLD emissario rotate-encrypt-key --old-key-env OLD_KEY

echo "== an old key the folder was not encrypted with: refused"
run env OLD_KEY=$THIRD EMISSARIO_ENCRYPT_KEY=$NEW emissario rotate-encrypt-key --old-key-env OLD_KEY

echo "== the rotation"
run env OLD_KEY=$OLD EMISSARIO_ENCRYPT_KEY=$NEW emissario rotate-encrypt-key --old-key-env OLD_KEY
cat data/encryption.json

echo "== the old key is refused, the new one opens every database"
run env EMISSARIO_ENCRYPT_KEY=$OLD emissario remove-ip-blocks --all
run env EMISSARIO_ENCRYPT_KEY=$NEW emissario remove-ip-blocks --all

echo "== run again: nothing to do"
run env OLD_KEY=$OLD EMISSARIO_ENCRYPT_KEY=$NEW emissario rotate-encrypt-key --old-key-env OLD_KEY

echo "== a rotation to a third key that stopped part-way"
env EMISSARIO_ENCRYPT_KEY=$THIRD emissario remove-ip-blocks --all --database third > /dev/null
printf '{"algorithm":"AES-256-GCM","key_id":"%s","rotating_to":"%s","version":2}\n' "$(keyid data)" "$(keyid third)" > data/encryption.json
run env EMISSARIO_ENCRYPT_KEY=$NEW emissario remove-ip-blocks --all

echo "== run again with both keys, it finishes"
run env OLD_KEY=$NEW EMISSARIO_ENCRYPT_KEY=$THIRD emissario rotate-encrypt-key --old-key-env OLD_KEY
cat data/encryption.json
run env EMISSARIO_ENCRYPT_KEY=$THIRD emissario remove-ip-blocks --all

echo "== nothing in clear under data/"
grep -rl admin@example.com data || echo "no email found in data/"

cd ..
mkdir -p assert-dir
cp test-dir/data/encryption.json assert-dir/
