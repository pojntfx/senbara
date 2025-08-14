package po

import "embed"

//go:generate go tool github.com/leonelquinteros/gotext/cli/xgotext -in .. -out .
//go:generate sh -c "find ../web/templates -name '*.html' -print0 | head -z -n1 | xargs -0 -I {} sh -c 'xgettext --keyword=_ --from-code=UTF-8 --add-location=file --join-existing --omit-header -o default.pot <(sed -z \"s|\\$\\.Locale\\.Get[[:space:]]*\\\"\\([^\\\"]*\\)\\\"|\\n_(\\\"\\1\\\");\\n|g; 1i#line 1 \\\"{}\\\"\" \"{}\")' && find ../web/templates -name '*.html' -print0 | tail -z -n+2 | xargs -0 -I {} sh -c 'xgettext --keyword=_ --from-code=UTF-8 --add-location=file --omit-header --join-existing -o default.pot <(sed -z \"s|\\$\\.Locale\\.Get[[:space:]]*\\\"\\([^\\\"]*\\)\\\"|\\n_(\\\"\\1\\\");\\n|g; 1i#line 1 \\\"{}\\\"\" \"{}\")'"
//go:generate sh -c "find . -name 'default.po' -print0 | xargs -0 -I {} msgmerge --update --backup=none \"{}\" default.pot"
//go:embed *
var FS embed.FS
