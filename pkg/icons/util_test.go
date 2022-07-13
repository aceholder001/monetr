package icons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSimpleIconsSlugMarkdown(t *testing.T) {
	data := "<!--\nThis file is automatically generated. If you want to change something, please\nupdate the script at 'scripts/release/update-slugs-table.js'.\n-->\n\n# Simple Icons slugs\n\n| Brand name | Brand slug |\n| :--- | :--- |\n| `.NET` | `dotnet` |\n| `/e/` | `e` |\n| `1001Tracklists` | `1001tracklists` |\n| `1Password` | `1password` |"

	result, err := parseSimpleIconsSlugMarkdown([]byte(data))
	assert.NoError(t, err, "must not return an error")
	assert.NotEmpty(t, result, "should have some brands parsed")
}
