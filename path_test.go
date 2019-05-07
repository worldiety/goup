package main

import (
	"testing"
)

func TestEmptyPath(t *testing.T) {
	cases := []string{"", "/"}
	for _, str := range cases {
		p := Path(str)
		if p.NameCount() != 0 {
			t.Fatal("expected 0 but got", p.NameCount())
		}

		if len(p.Names()) != 0 {
			t.Fatal("expected 0 but got", len(p.Names()))
		}

		if p.Parent().NameCount() != 0 {
			t.Fatal("expected 0 but got", p.Parent().NameCount())
		}

		if p.String() != "/" {
			t.Fatal("expected / but got", p.String())
		}
	}
}

func Test1Path(t *testing.T) {
	cases := []string{"a", "/a", "a/", "/a/"}
	for _, str := range cases {
		p := Path(str)
		if p.NameCount() != 1 {
			t.Fatal("expected 1 but got", p.NameCount(), " => "+str)
		}

		if len(p.Names()) != 1 {
			t.Fatal("expected 1 but got", len(p.Names()), " => "+str)
		}

		if p.NameAt(0) != "a" {
			t.Fatal("expected a but got", p.NameAt(0))
		}

		if p.Parent().NameCount() != 0 {
			t.Fatal("expected 0 but got", p.Parent().NameCount())
		}

		if p.String() != "/a" {
			t.Fatal("expected /a but got", p.String())
		}
	}
}

func Test2Path(t *testing.T) {
	cases := []string{"a/b", "/a/b", "a/b", "/a/b/"}
	for _, str := range cases {
		p := Path(str)
		if p.NameCount() != 2 {
			t.Fatal("expected 1 but got", p.NameCount(), " => "+str)
		}

		if len(p.Names()) != 2 {
			t.Fatal("expected 1 but got", len(p.Names()), " => "+str)
		}

		if p.NameAt(0) != "a" {
			t.Fatal("expected a but got", p.NameAt(0))
		}

		if p.NameAt(1) != "b" {
			t.Fatal("expected b but got", p.NameAt(1))
		}

		if p.Parent().NameCount() != 1 {
			t.Fatal("expected 1 but got", p.Parent().NameCount())
		}

		if p.String() != "/a/b" {
			t.Fatal("expected /a/b but got", p.String())
		}
	}
}

func TestModPath(t *testing.T) {
	p := ConcatPaths("a/b/", "/c")
	if p.String() != "/a/b/c" {
		t.Fatal("expected /a/b/c but got", p)
	}

	p = p.Child("d")
	if p.String() != "/a/b/c/d" {
		t.Fatal("expected /a/b/c/d but got", p)
	}

	p = p.TrimPrefix(Path("a/b/c"))
	if p.String() != "/d" {
		t.Fatal("expected /d but got", p)
	}

	p = p.Child("/x/y/z")
	if p.String() != "/d/x/y/z" {
		t.Fatal("expected /d/x/y/z but got", p)
	}

	p = ""
	p = p.Child("/a/b/c")
	if p.String() != "/a/b/c" {
		t.Fatal("expected /a/b/c but got", p)
	}
}

func TestPath_Normalize(t *testing.T) {
	p := Path("..")
	if p.Normalize().String() != "/" {
		t.Fatal("expected / but got", p.Normalize().String())
	}

	p = Path("../../a")
	if p.Normalize().String() != "/a" {
		t.Fatal("expected /a but got", p.Normalize().String())
	}

	p = Path("a/b/c/..")
	if p.Normalize().String() != "/a/b" {
		t.Fatal("expected /a/b but got", p.Normalize().String())
	}

	p = Path("a/b/c/../d")
	if p.Normalize().String() != "/a/b/d" {
		t.Fatal("expected /a/b/d but got", p.Normalize().String())
	}

	p = Path("a/b/c/../d/./e")
	if p.Normalize().String() != "/a/b/d/e" {
		t.Fatal("expected /a/b/d/e but got", p.Normalize().String())
	}
}

func TestPath_Name(t *testing.T) {
	p := Path("")
	if p.Name() != "" {
		t.Fatal("expected '' but got", p.Name())
	}

	p = Path("/")
	if p.Name() != "" {
		t.Fatal("expected '' but got", p.Name())
	}

	p = Path("/a")
	if p.Name() != "a" {
		t.Fatal("expected 'a' but got", p.Name())
	}

	p = Path("/a/b")
	if p.Name() != "b" {
		t.Fatal("expected 'b' but got", p.Name())
	}
}

func TestPath_StartsWith(t *testing.T) {
	p := Path("")
	if !p.StartsWith("") {
		t.Fatal("expected to start with ''")
	}

	p = Path("a")
	if !p.StartsWith("/a") {
		t.Fatal("expected to start with '/a'")
	}

	p = Path("a/b/c/d")
	if !p.StartsWith("/a") {
		t.Fatal("expected to start with '/a'")
	}
}

func TestPath_EndsWith(t *testing.T) {
	p := Path("")
	if !p.EndsWith("") {
		t.Fatal("expected to end with ''")
	}

	p = Path("a")
	if !p.EndsWith("/a") {
		t.Fatal("expected to end with '/a'")
	}

	p = Path("a/b/c/d")
	if !p.EndsWith("/d") {
		t.Fatal("expected to end with '/d'")
	}
}
