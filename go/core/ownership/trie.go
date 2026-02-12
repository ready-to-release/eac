package ownership

import "strings"

// pathTrie is a prefix trie keyed on path segments (split by "/").
// Each node stores the ownerEntry values whose root exactly matches
// the path to that node.  Walking the trie from root to the deepest
// matching segment collects all candidate entries efficiently.
type pathTrie struct {
	children map[string]*pathTrie
	entries  []ownerEntry // entries whose root == path to this node
	catchAll []ownerEntry // entries with root "/" (global catch-all)
}

// newPathTrie builds a trie from the non-file entries in the resolver.
func newPathTrie(entries []ownerEntry) *pathTrie {
	root := &pathTrie{}
	for _, e := range entries {
		if e.isFile {
			continue
		}
		if e.root == "/" {
			root.catchAll = append(root.catchAll, e)
			continue
		}
		segments := strings.Split(e.root, "/")
		node := root
		for _, seg := range segments {
			if seg == "" {
				continue
			}
			if node.children == nil {
				node.children = make(map[string]*pathTrie)
			}
			child, ok := node.children[seg]
			if !ok {
				child = &pathTrie{}
				node.children[seg] = child
			}
			node = child
		}
		node.entries = append(node.entries, e)
	}
	return root
}

// findCandidatesTrie walks the trie following the segments of path,
// collecting the deepest matching entries.  Returns all entries at the
// most-specific (deepest) matching depth, which mirrors the behaviour
// of the linear findCandidates.
func (t *pathTrie) findCandidatesTrie(path string) []ownerEntry {
	segments := strings.Split(path, "/")

	var bestDepth int = -1
	var candidates []ownerEntry

	// Walk the trie.  At each node, if there are entries, they are
	// deeper than (or equal to) any previously seen entries because
	// we walk top-down; we only keep the deepest.
	node := t
	for _, seg := range segments {
		if node.children == nil {
			break
		}
		child, ok := node.children[seg]
		if !ok {
			break
		}
		node = child
		if len(node.entries) > 0 {
			// These entries are at a deeper depth than anything before.
			depth := node.entries[0].depth
			if depth > bestDepth {
				bestDepth = depth
				candidates = node.entries
			}
		}
	}

	// If no prefix entries matched, fall back to catch-all.
	if bestDepth == -1 && len(t.catchAll) > 0 {
		return t.catchAll
	}

	return candidates
}
