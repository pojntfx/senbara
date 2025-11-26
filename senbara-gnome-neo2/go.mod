module github.com/pojntfx/senbara/senbara-gnome-neo

go 1.24.0

replace (
	github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172 => ../../puregotk
	github.com/pojntfx/senbara/senbara-gtk-go v0.0.0-20250826075235-cbb2c7573805 => ../senbara-gtk-go
)

require (
	github.com/jwijenbergh/puregotk v0.0.0-20251022075221-eae1610c7d83
	github.com/pojntfx/senbara/senbara-gtk-go v0.0.0-20251011063231-959fe0be4948
)

require github.com/jwijenbergh/purego v0.0.0-20251017112123-b71757b9ba42 // indirect

replace github.com/jwijenbergh/puregotk => github.com/pojntfx/puregotk v0.0.0-20251125051126-73ef36c6a49c
