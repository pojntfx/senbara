module github.com/pojntfx/senbara/senbara-gtk

go 1.25.0

require (
	github.com/jwijenbergh/puregotk v0.0.0-20251022075221-eae1610c7d83
	github.com/pojntfx/go-gettext v0.1.2
)

require github.com/jwijenbergh/purego v0.0.0-20251017112123-b71757b9ba42 // indirect

replace github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172 => ../../puregotk

replace github.com/jwijenbergh/puregotk => github.com/pojntfx/puregotk v0.0.0-20251125051126-73ef36c6a49c
