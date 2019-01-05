package presetHandlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

func ScriptHandler(scriptContent string) func(*gin.Context, *http.Response, *string) *string {
	return func(ctx *gin.Context, res *http.Response, body *string) *string {
		r := addScript(*body, scriptContent)
		return &r
	}
}

/*添加<script></script>*/
func addScript(s string, scriptContent string) string {
	r := regexp.MustCompile(`(?i)<head.*?>`)
	if l := r.FindStringIndex(s); len(l) > 0 {
		LINK := `<script>` + scriptContent + `</script>`
		//将其加到<head>之后
		return s[:l[1]] + LINK + s[l[1]:]
	}
	return s
}
