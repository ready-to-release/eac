"""Behave environment hooks."""


def before_scenario(context, scenario):
    """Reset state before each scenario."""
    context.values = None
    context.report = None
