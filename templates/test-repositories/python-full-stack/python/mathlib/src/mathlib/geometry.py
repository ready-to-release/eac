"""Geometry calculations."""

import math


def circle_area(radius: float) -> float:
    """Calculate the area of a circle."""
    if radius < 0:
        raise ValueError("Radius cannot be negative")
    return math.pi * radius ** 2


def rectangle_area(width: float, height: float) -> float:
    """Calculate the area of a rectangle."""
    if width < 0 or height < 0:
        raise ValueError("Dimensions cannot be negative")
    return width * height
