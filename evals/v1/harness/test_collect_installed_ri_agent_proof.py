import contextlib
import importlib.util
import io
import json
from pathlib import Path
from types import SimpleNamespace
import sys
import tempfile
import unittest
from unittest import mock


COLLECTOR_PATH = Path(__file__).with_name("collect-installed-ri-agent-proof.py")
spec = importlib.util.spec_from_file_location("collect_installed_ri_agent_proof", COLLECTOR_PATH)
collector = importlib.util.module_from_spec(spec)
spec.loader.exec_module(collector)


class RunIDBoundaryTests(unittest.TestCase):
    def test_malformed_run_ids_never_invoke_fabric_or_open_controller(self):
        malformed = ["--export-jsonl", "../journal", "A" * 64, "a" * 63, "g" * 64]
        with tempfile.TemporaryDirectory() as directory:
            installed = Path(directory) / "ri.exe"
            installed.write_bytes(b"installed")
            for run_id in malformed:
                with self.subTest(run_id=run_id), \
                     mock.patch.object(collector, "controller_path") as controller_path, \
                     mock.patch("subprocess.run") as run:
                    argv = ["collector", "--run-id=" + run_id, "--repository-root", directory,
                            "--installed-ri", str(installed), "--fabric-cli", str(installed)]
                    with mock.patch.object(sys, "argv", argv), contextlib.redirect_stdout(io.StringIO()):
                        collector.main()
                    controller_path.assert_not_called()
                    self.assertEqual(run.call_count, 0)

    def test_sink_rejects_non_string_run_id_before_cli(self):
        with mock.patch("subprocess.run") as run:
            self.assertIsNone(collector.inspect_with_fabric("missing-cli", "missing-root", 7, "0" * 64))
        self.assertEqual(run.call_count, 0)

    def test_valid_run_id_preserves_inspect_and_export_arguments(self):
        run_id = "a" * 64
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / "repo"
            root.mkdir()
            binary = Path(directory) / "fabric.exe"
            binary.write_bytes(b"fabric")
            payload = {}
            event = {"version": 1, "sequence": 1, "previous": collector.ZERO,
                     "kind": "run.created", "payload": payload}
            event["hash"] = collector.domain_hash("harness.event.v1", {
                "version": event["version"], "sequence": event["sequence"],
                "previous": event["previous"], "kind": event["kind"], "payload": payload})
            exported = (collector.canon(event) + "\n").encode("utf-8")
            completed = [SimpleNamespace(returncode=0, stdout=exported),
                         SimpleNamespace(returncode=0, stdout=json.dumps({"run_id": run_id}).encode("utf-8"))]
            with mock.patch("subprocess.run", side_effect=completed) as run:
                snapshot = collector.inspect_with_fabric(binary, root, run_id, event["hash"])
            self.assertEqual(snapshot, {"run_id": run_id})
            resolved_binary = str(binary.resolve(strict=True))
            resolved_root = str(root.resolve(strict=True))
            self.assertEqual([call.args[0] for call in run.call_args_list], [
                [resolved_binary, "--root", resolved_root, "inspect", run_id, "--export-jsonl"],
                [resolved_binary, "--root", resolved_root, "inspect", run_id],
            ])


if __name__ == "__main__":
    unittest.main()
