# Localize a failure using individual Go coverage profiles

Development source supports a read-only spectrum comparison:

```powershell
fabric diagnose RUN --spectrum C:\private-evidence\spectrum.json
```

This imports already obtained evidence. It does not run tests, call a model,
write source, append journal events, choose a fixer or close findings. The
ordinary diagnosis and its admitted repair scope stay intact. Imported test
outcomes and filename mappings remain **caller-supplied, untrusted assertions**;
observing matching source hashes does not authenticate their production.

## Prepare evidence from the exact candidate

Run individual tests in a separate evaluator copy of the exact candidate, and
retain the commands, exit status, executable identity and complete profiles.
Do not put new profiles inside an admitted candidate workspace: unadmitted
files can change candidate identity. Use the same instrumentation and source
for every test. Examples, executed explicitly by the operator:

```powershell
# Working directory: evaluator copy of the exact candidate.
go test -run '^TestFailingCase$' -count=1 -covermode=set -coverprofile=C:\private-evidence\failed.out ./parser
go test -run '^TestPassingCase$' -count=1 -covermode=set -coverprofile=C:\private-evidence\passed.out ./parser
```

These are [native Go coverage profiles](https://pkg.go.dev/golang.org/x/tools/cover).
A combined-suite profile cannot identify individual test/block associations.
Do not label launch errors, timeouts, missing profiles or UNKNOWN as failing
tests. At least one individual FAIL observation is required. A passing test is
useful for discrimination but not mandatory. Test IDs should identify the exact
individual command/receipt; the import does not authenticate those receipts.

The JSON artifact wraps profiles without introducing a transformation DSL:

```json
{
  "version": 1,
  "run_id": "<exact 64-character run hash>",
  "candidate_id": "<exact 64-character candidate hash>",
  "sources": [
    {
      "path": "parser/parser.go",
      "profile_path": "example.invalid/project/parser/parser.go",
      "sha256": "<SHA-256 of the complete candidate file>"
    }
  ],
  "tests": [
    {
      "id": "TestFailingCase:receipt-id",
      "outcome": "FAIL",
      "profile": "<complete failed.out text, including mode header>"
    },
    {
      "id": "TestPassingCase:receipt-id",
      "outcome": "PASS",
      "profile": "<complete passed.out text, including mode header>"
    }
  ]
}
```

Placeholders must be replaced; this illustrative artifact is not executable.
Map every profile filename explicitly to its repository-relative candidate
file. The importer does not guess module prefixes or reinterpret absolute paths.
All profiles must share one mode and the exact same block inventory, including
zero counters and statement counts. Missing blocks are not inferred uncovered.
`set`, `count` and `atomic` modes are supported; a positive visit count means
one test executed the block, regardless of how many times it visited it.

## Read the ranking

For each block, the report retains failed/passed test execution counts and
the original source range. With `F` failed tests, `ef` failing tests executing
the block and `ep` passing tests executing it:

```text
Ochiai = ef / sqrt(F * (ef + ep))
Ochiai squared = ef^2 / (F * (ef + ep))
```

Exact numerator/denominator integers preserve canonical JSON and deterministic
comparisons. Zero denominator means no test executed the block, not a successful
test or a known safe location. Equal scores use deterministic path/range order.
These are suspiciousness ranks, not calibrated defect probabilities or proof of
causation. No manually tuned weights are used.

Fabric validates complete source hashes and range bounds under its existing
read lease, brackets source reads with candidate observations and journal heads,
and checks guard ownership when closing the lease. Only
`complete_candidate_bytes_observed` publishes ranked blocks. Candidate drift,
missing/partial files, mismatched hashes, invalid ranges or unavailable read
ownership produce `source_status: unavailable` with no ranked blocks, while
preserving the base diagnosis and explicitly untrusted aggregate counts.
Malformed or foreign-run/candidate artifacts are rejected with bounded errors.

Limits: 1 MiB encoded JSON, 64 tests, 24 source files, 64 KiB per profile,
4,096 blocks per profile, 32 KiB complete source per file. Oversized evidence is
not silently truncated into a valid spectrum. Source/control configuration and
ambiguous paths cannot be admitted through this interface. Keep artifacts and
reports private; filenames/test IDs may contain project information.

The flag is exclusive with `--anchor`, `--previous` and `--closure`. Default
diagnosis serialization and old invocation bytes are unchanged. JEV acquisition
selection, automatic profile collection, graph/ranking fusion and automatic
fixer consumption remain future integrations. Independent verification/review
and exact closure oracles remain required after any repair.

See the [contract](../specifications/repair-diagnosis.md) and
[executed evidence](../evaluation/repair-spectrum-localization.md).
