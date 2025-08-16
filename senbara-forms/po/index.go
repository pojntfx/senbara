package po

import "embed"

//go:generate rm -f default.pot
//go:generate go tool github.com/leonelquinteros/gotext/cli/xgotext -in .. -out .
//go:generate sh -c "(cat default.pot; echo ''; find .. -name '*.html' -exec go tool github.com/unDocUMeantIt/tgotext/cmd/tgotext parse {} --object '$.Locale' \\;) | msguniq --sort-output -o default.pot"
//go:generate sh -c "find . -name 'default.po' -print0 | xargs -0 -I {} msgmerge --update --backup=none \"{}\" default.pot"
//go:embed *
var FS embed.FS
