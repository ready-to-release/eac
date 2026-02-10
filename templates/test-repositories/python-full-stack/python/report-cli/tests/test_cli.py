"""Unit tests for report CLI."""

from report_cli.cli import generate_report


class TestGenerateReport:
    def test_basic_report(self):
        result = generate_report([10, 20, 30])
        assert "Mean:   20.0" in result
        assert "Median: 20" in result

    def test_empty_data(self):
        result = generate_report([])
        assert "Error" in result
