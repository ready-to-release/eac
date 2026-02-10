"""Behave environment hooks."""


def before_scenario(context, scenario):
    """Reset state before each scenario."""
    context.store = None
    context.result = None
    context.delete_result = None
