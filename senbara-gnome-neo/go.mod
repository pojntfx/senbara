module github.com/pojntfx/senbara/senbara-gnome-neo

go 1.24.0

replace (
	github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172 => github.com/pojntfx/puregotk v0.0.0-20251011060225-c87603a2de88
	github.com/pojntfx/senbara/senbara-gtk-go v0.0.0-20250826075235-cbb2c7573805 => ../senbara-gtk-go
)

require (
	github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172
	github.com/pojntfx/senbara/senbara-gtk-go v0.0.0-20250826075235-cbb2c7573805
)

require github.com/jwijenbergh/purego v0.0.0-20250812133547-b5852df1402b // indirect
