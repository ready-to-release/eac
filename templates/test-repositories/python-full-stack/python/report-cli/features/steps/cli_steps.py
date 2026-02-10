"""Step definitions for report generation."""

from behave import given, when, then
from report_cli.cli import generate_report


@given("a dataset with values {values}")
def step_dataset(context, values):
    context.values = [float(v.strip()) for v in values.split(",")]


@given("an empty dataset")
def step_empty_dataset(context):
    context.values = []


@when("I generate a summary report")
def step_generate_report(context):
    context.report = generate_report(context.values)


@then("the report should contain the mean value {expected}")
def step_check_mean(context, expected):
    assert f"Mean:   {expected}" in context.report


@then("the report should contain the median value {expected}")
def step_check_median(context, expected):
    assert f"Median: {expected}" in context.report


@then("the report should show an error message")
def step_check_error(context):
    assert "Error" in context.report
