"""Exercise process resource measurements without invoking a model provider."""

import os
from pathlib import Path
import sys
import unittest

from benchmark_startup import sample_command


class ProcessResourceTests(unittest.TestCase):
    @unittest.skipUnless(os.name == "nt", "Windows process counters")
    def test_completed_process_reports_cpu_io_and_lifetime_peak(self):
        metrics = {}
        stdout, stderr, elapsed, peak = sample_command(
            [sys.executable, "-c", "print(sum(i*i for i in range(100000)))"],
            Path(__file__).resolve().parent,
            resource_metrics=metrics,
        )
        self.assertTrue(stdout.strip().isdigit())
        self.assertEqual(stderr, b"")
        self.assertGreater(elapsed, 0)
        self.assertGreater(peak, 0)
        self.assertEqual(metrics["peak_working_set_bytes"], peak)
        self.assertEqual(metrics["scope"], "process_only_excludes_children")
        self.assertGreaterEqual(metrics["kernel_cpu_ns"], 0)
        self.assertGreaterEqual(metrics["user_cpu_ns"], 0)
        self.assertGreater(metrics["kernel_cpu_ns"] + metrics["user_cpu_ns"], 0)
        self.assertEqual(set(metrics["io"]), {
            "read_operations", "write_operations", "other_operations",
            "read_bytes", "write_bytes", "other_bytes",
        })
        for value in metrics["io"].values():
            self.assertIsInstance(value, int)
            self.assertGreaterEqual(value, 0)

    def test_resource_collection_is_optional(self):
        stdout, stderr, elapsed, _ = sample_command(
            [sys.executable, "-c", "print('measurement-fixture')"],
            Path(__file__).resolve().parent,
        )
        self.assertEqual(stdout.strip(), b"measurement-fixture")
        self.assertEqual(stderr, b"")
        self.assertGreater(elapsed, 0)


if __name__ == "__main__":
    unittest.main()
