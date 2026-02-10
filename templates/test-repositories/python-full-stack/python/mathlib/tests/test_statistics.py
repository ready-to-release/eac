"""Unit tests for statistics module."""

import pytest
from mathlib.statistics import mean, median


class TestMean:
    def test_integers(self):
        assert mean([1, 2, 3]) == 2.0

    def test_single_value(self):
        assert mean([42]) == 42.0

    def test_empty_list(self):
        with pytest.raises(ValueError):
            mean([])


class TestMedian:
    def test_odd_count(self):
        assert median([1, 2, 3]) == 2

    def test_even_count(self):
        assert median([1, 2, 3, 4]) == 2.5

    def test_unsorted(self):
        assert median([3, 1, 2]) == 2

    def test_empty_list(self):
        with pytest.raises(ValueError):
            median([])
