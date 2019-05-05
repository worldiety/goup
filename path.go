package main

import (
	"io/ioutil"
	"os"
	"strings"
)

// Path is our helper for real paths, copied from our VFS, we should introduce a module for that alone
type Path string

// StartsWith tests whether the path begins with prefix.
func (p Path) StartsWith(prefix Path) bool {
	return strings.HasPrefix(p.String(), prefix.String())
}

// EndsWith tests whether the path ends with prefix.
func (p Path) EndsWith(suffix Path) bool {
	return strings.HasSuffix(p.String(), suffix.String())
}

// Names splits the path by / and returns all segments as a simple string array.
func (p Path) Names() []string {
	tmp := strings.Split(string(p), "/")
	cleaned := make([]string, len(tmp))
	idx := 0
	for _, str := range tmp {
		str = strings.TrimSpace(str)
		if len(str) > 0 {
			cleaned[idx] = str
			idx++
		}
	}
	return cleaned[0:idx]
}

// NameCount returns how many names are included in this path.
func (p Path) NameCount() int {
	return len(p.Names())
}

// NameAt returns the name at the given index.
func (p Path) NameAt(idx int) string {
	return p.Names()[idx]
}

// Name returns the last element in this path or the empty string if this path is empty.
func (p Path) Name() string {
	tmp := p.Names()
	if len(tmp) > 0 {
		return tmp[len(tmp)-1]
	}
	return ""
}

// Parent returns the parent path of this path.
func (p Path) Parent() Path {
	tmp := p.Names()
	if len(tmp) > 0 {
		return Path(strings.Join(tmp[:len(tmp)-1], "/"))
	}
	return ""
}

// String normalizes the slashes in Path
func (p Path) String() string {
	return "/" + strings.Join(p.Names(), "/")
}

// Child returns a new Path with name appended as a child
func (p Path) Child(name string) Path {
	if len(p) == 0 {
		if strings.HasPrefix(name, "/") {
			return Path(name)
		}
		return Path("/" + name)

	}
	if strings.HasPrefix(name, "/") {
		return Path(p.String() + name)
	}
	return Path(p.String() + "/" + name)
}

// Exists checks if the (absolute) file exists
func (p Path) Exists() bool {
	_, err := os.Stat(p.String())
	return err == nil
}

// IsDir checks if path represents a directory
func (p Path) IsDir() bool {
	stat, err := os.Stat(p.String())
	if err != nil {
		return false
	}
	return stat.IsDir()
}

// TrimPrefix returns a path without the prefix
func (p Path) TrimPrefix(prefix Path) Path {
	tmp := "/" + strings.TrimPrefix(p.String(), prefix.String())
	return Path(tmp)
}

// Normalize removes any . and .. and returns a Path without those elements.
func (p Path) Normalize() Path {
	names := p.Names()
	tmp := make([]string, len(names))[0:0]
	for _, name := range names {
		if name == "." {
			continue
		}
		if name == ".." {
			// walk up -> remove last segment
			if len(tmp) > 0 {
				tmp = tmp[:len(tmp)-1]
			}
			continue
		}
		tmp = append(tmp, name)
	}
	return Path("/" + strings.Join(tmp, "/"))
}

// Resolve takes the base dir and normalizes this path and returns it.
// E.g.
//   1. "/my/path".Resolve("/some/thing") => /my/path
//   2. "./my/path".Resolve("/some/thing") => /some/thing/my/path
//   3. "my/path".Resolve("/some/thing") => /some/thing/my/path
//   4. "my/path".Resolve("/some/thing") => /some/thing/my/path
//   5. "my/path/../../".Resolve("/some/thing") => /some/thing
func (p Path) Resolve(baseDir Path) Path {
	// 1. the absolute filename case
	if strings.HasPrefix(string(p), "/") {
		return p.Normalize()
	}

	// any other case (2-5) needs the prefix of baseDir
	return baseDir.Add(p).Normalize()
}

// Add returns this path concated with the given path. Sadly we have no overloading operator.
func (p Path) Add(path Path) Path {
	return ConcatPaths(p, path)
}

// List returns the list of accessible children
func (p Path) List() (children []Path) {
	files, _ := ioutil.ReadDir(p.String())
	for _, f := range files {
		children = append(children, p.Child(f.Name()))
	}
	return
}

// ConcatPaths merges all paths together
func ConcatPaths(paths ...Path) Path {
	tmp := make([]string, 0)
	for _, path := range paths {
		for _, name := range path.Names() {
			tmp = append(tmp, name)
		}
	}
	return Path("/" + strings.Join(tmp, "/"))
}
