package presetHandlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func WidthLimitHandler() func(*gin.Context, *http.Response, *string) *string {
	return func(ctx *gin.Context, res *http.Response, body *string) *string {
		r := *body
		//判断是否为max-width-limit-[0-9]+.css, 是的话返回一个css文件
		limit, ok := ctx.GetQuery("limit")
		if !ok {
			println(limit)
			return &r
		}
		//r = addStyle(r,fmt.Sprintf(`p{max-width:%spx!important;}`, limit))
		r = addScript(r, `
//适配iphone上iframe宽度不是chengk页面的宽度不受父页面
window.addEventListener('load',function(){
     //alert("网页可见区域宽："+document.body.clientWidth+"\n 屏幕可用工作区宽度："+ window.screen.availWidth+"\n");
 
     if (!navigator.userAgent.match(/iPad|iPhone/i)){
         //alert("非ios");
         return false;
     }
     //如果是iphone,ipad，则重新修改body宽度值
     document.body.style.width = Math.min(window.screen.availWidth/document.body.clientWidth, document.body.clientWidth/window.screen.availWidth)*100+'%';
     document.body.style.overflow = 'scroll';
})
`)
		return &r
	}
}