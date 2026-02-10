"""Step definitions for key-value store operations."""

from behave import given, when, then
from data_store.store import KeyValueStore


@given("an empty data store")
def step_empty_store(context):
    context.store = KeyValueStore()


@when('I set key "{key}" to value "{value}"')
def step_set_key(context, key, value):
    context.store.set(key, value)


@when('I get key "{key}"')
def step_get_key(context, key):
    context.result = context.store.get(key)


@when('I delete key "{key}"')
def step_delete_key(context, key):
    context.delete_result = context.store.delete(key)


@then('getting key "{key}" should return "{expected}"')
def step_verify_value(context, key, expected):
    actual = context.store.get(key)
    assert actual == expected, f"Expected '{expected}', got '{actual}'"


@then("the result should be empty")
def step_result_empty(context):
    assert context.result is None


@then('getting key "{key}" should be empty')
def step_key_empty(context, key):
    assert context.store.get(key) is None


@then("the store should have {count:d} entries")
def step_store_count(context, count):
    assert context.store.count() == count


@then("the delete result should be false")
def step_delete_false(context):
    assert context.delete_result is False
