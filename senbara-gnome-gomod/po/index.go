package po

import "embed"

//go:generate sh -c "find ../src -name '*.go' -o -name '*.blp' | xgettext --language=C++ --keyword=_ --keyword=L --omit-header -o senbara-gnome-gomod.pot --files-from=-"
//go:generate sh -c "find . -name 'senbara-gnome-gomod.po' -print0 | xargs -0 -I {} msgmerge --update --backup=none \"{}\" senbara-gnome-gomod.pot"
//go:generate sh -c "find . -type f -name '*.po' -print0 | xargs -0 -I {} sh -c 'msgfmt -o \"{}.mo\" \"{}\"' && find . -type f -name '*.po.mo' -exec sh -c 'mv \"{}\" \"$(echo \"{}\" | sed s/\\.po\\.mo/.mo/)\"' \\;"
//go:generate sh -c "ls -d */ | sed s@/@@g > LINGUAS"
//go:embed *
var FS embed.FS
