package golangfuse

import "regexp"

func JinjaToGoTemplate(prompt string) string {
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)
	return re.ReplaceAllString(prompt, "{{.$1}}")
}
