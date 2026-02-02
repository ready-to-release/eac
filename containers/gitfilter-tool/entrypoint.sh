#!/bin/bash
set -e

echo "=== git-filter-repo container ==="
git filter-repo --version
echo ""

if [ ! -f /mailmap ]; then
    echo "ERROR: /mailmap file not mounted"
    exit 1
fi

echo "=== Mailmap contents ==="
cat /mailmap
echo ""

echo "=== Before: Author distribution ==="
git log --format='%an <%ae>' | sort | uniq -c | sort -rn
echo ""

echo "=== Running git-filter-repo ==="
git filter-repo --mailmap /mailmap --force

echo ""
echo "=== After: Author distribution ==="
git log --format='%an <%ae>' | sort | uniq -c | sort -rn
echo ""

echo "=== Done! Ready for: git push --force --mirror origin ==="
