module github.com/pojntfx/senbara/senbara-gtk

go 1.24.0

require (
	github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172
	github.com/pojntfx/senbara/libsenbara-gtk-go v0.0.0-20250826075235-cbb2c7573805
)

require github.com/jwijenbergh/purego v0.0.0-20250812133547-b5852df1402b // indirect

replace (
	github.com/jwijenbergh/purego v0.0.0-20250812133547-b5852df1402b => ../../purego
	github.com/jwijenbergh/puregotk v0.0.0-20250812133623-7203178b5172 => ../../puregotk
)
