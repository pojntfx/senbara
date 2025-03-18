package locales

import "embed"

//go:generate find . -type f -name '*.po' -exec sh -c 'msgfmt {} -o $(echo {} | sed -e s/.po$//).mo' \;
//go:embed *
var FS embed.FS
