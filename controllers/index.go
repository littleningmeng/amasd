package controllers

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"scrapyd-admin/config"
	"scrapyd-admin/core"
	"scrapyd-admin/models"
)

type Index struct {
	core.BaseController
}

func (i *Index) Index(c *gin.Context) {
	sm := new(models.SystemMenu)
	passport := core.GetPassportInstance()
	c.HTML(http.StatusOK, "index/index", gin.H{
		"menuStr":  template.HTML(sm.GetMenuStr()),
		"realName": passport.RealName,
		"id":       passport.Id,
	})
}

func (i *Index) Main(c *gin.Context) {
	c.HTML(http.StatusOK, "index/main", gin.H{})
}

func (i *Index) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "index/login", gin.H{})
}

//登录
func (i *Index) DoLogin(c *gin.Context) {
	var admin models.Admin
	if err := c.ShouldBind(&admin); err == nil {
		ok, code := admin.Login()
		if !ok {
			i.Fail(c, config.PromptMsg[code])
			return
		}
		i.Success(c)
	} else {
		i.Fail(c, config.PromptMsg["system_error"])
	}
}

//退出登录
func (i *Index) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func (i *Index) Error(c *gin.Context) {

}
