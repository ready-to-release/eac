"""In-memory key-value store implementation."""


class KeyValueStore:
    """Simple in-memory key-value store."""

    def __init__(self):
        self._data: dict[str, str] = {}

    def set(self, key: str, value: str) -> None:
        """Store a key-value pair."""
        self._data[key] = value

    def get(self, key: str) -> str | None:
        """Retrieve a value by key. Returns None if not found."""
        return self._data.get(key)

    def delete(self, key: str) -> bool:
        """Delete a key. Returns True if the key existed."""
        if key in self._data:
            del self._data[key]
            return True
        return False

    def count(self) -> int:
        """Return the number of stored entries."""
        return len(self._data)
