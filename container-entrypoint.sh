#!/bin/sh
# The binaries path must be absolute: pvtr resolves the plugin binary as
# filepath.Join(binaries-path, binaryPath), and a bare relative name like
# "github-repo" would be looked up in $PATH instead of this directory.
#
# Keep the console quiet during the run; the full log is written to
# evaluation_results and printed below. Output is captured (not discarded)
# so it can be surfaced if pvtr crashes before writing any logs.
OUTPUT=$(./pvtr run --binaries-path /.privateer/bin --config /.privateer/config.yml 2>&1)
STATUS=$?

for file in evaluation_results/**/*.log; do echo "$file"; cat "$file"; echo; done

# Exit 0 (all checks passed) and exit 1 (some checks failed) both mean the
# scan ran to completion; the container reports results rather than failing.
# Any other exit code means pvtr crashed, so surface its output and propagate.
if [ "$STATUS" -ne 0 ] && [ "$STATUS" -ne 1 ]; then
  echo "ERROR: pvtr run exited with code $STATUS" >&2
  echo "$OUTPUT" >&2
  exit "$STATUS"
fi

exit 0
