package pkgb

import "github.com/worldiety/std"

// GetMap1 returns a map[interface]interface{}
func GetMap1() *std.Map {
	myIfaceMap := &std.Map{}
	myIfaceMap.Put(std.NewBox("hello"), std.NewBox("world"))
	myIfaceMap.Put(std.NewBox(4), std.NewBox(6))
	return myIfaceMap
}

// GetMap2 returns a map[string]string
func GetMap2() *std.StrStrMap {
	categories := &std.StrStrMap{}
	categories.Put("gandalf", "mage")
	categories.Put("bilbo", "hobbit")
	categories.Put("sam", "hobbit")
	categories.Put("gimli", "dwarf")
	categories.Put("aragorn", "numenor")
	return categories
}

// GetMap3 returns a map[string]interface{}
func GetMap3() *std.StrMap {
	categories := &std.StrMap{}
	categories.Put("mages", std.NewBox(3))
	categories.Put("bilbo", std.NewBox("hobbit"))
	categories.Put("hasHobbits", std.NewBox(true))
	categories.Put("pie", std.NewBox(3.14))
	return categories
}

// GetSlice1 returns a string slice
func GetSlice1() *std.StrSlice {
	return &std.StrSlice{[]string{"a", "b", "c"}}
}
