"""CLI entry point for report generation."""

import sys
from mathlib.statistics import mean, median


def generate_report(values: list[float]) -> str:
    """Generate a summary report for the given values."""
    if not values:
        return "Error: No data provided"
    return (
        f"Summary Report\n"
        f"==============\n"
        f"Count:  {len(values)}\n"
        f"Mean:   {mean(values)}\n"
        f"Median: {median(values)}\n"
    )


def main():
    """CLI entry point."""
    if len(sys.argv) < 2:
        print("Usage: report <value1> <value2> ...")
        sys.exit(1)
    values = [float(v) for v in sys.argv[1:]]
    print(generate_report(values))
