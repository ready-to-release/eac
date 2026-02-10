"""Unit tests for geometry module."""

import math
import pytest
from mathlib.geometry import circle_area, rectangle_area


class TestCircleArea:
    def test_unit_circle(self):
        assert circle_area(1) == pytest.approx(math.pi)

    def test_zero_radius(self):
        assert circle_area(0) == 0

    def test_negative_radius(self):
        with pytest.raises(ValueError):
            circle_area(-1)


class TestRectangleArea:
    def test_square(self):
        assert rectangle_area(5, 5) == 25

    def test_rectangle(self):
        assert rectangle_area(3, 4) == 12

    def test_negative_dimension(self):
        with pytest.raises(ValueError):
            rectangle_area(-1, 5)
