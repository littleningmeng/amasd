package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"scrapyd-admin/config"
	"scrapyd-admin/core"
	"scrapyd-admin/models"
	"strconv"
)

type Server struct {
	core.BaseController
}

func (s *Server) Index(c *gin.Context) {
	if core.IsAjax(c) {
		projectId, _ := strconv.Atoi(c.DefaultPostForm("project_id", "0"))
		page, _ := strconv.Atoi(c.DefaultPostForm("pagination[page]", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultPostForm("pagination[perpage]", "10"))
		servers, totalCount := new(models.Server).PageList(projectId, page, pageSize)
		c.JSON(http.StatusOK, gin.H{
			"data": servers,
			"meta": gin.H{
				"page":    page,
				"total":   totalCount,
				"pages":   core.CalculationPages(totalCount, pageSize),
				"perpage": pageSize,
			},
		})
	} else {
		c.HTML(http.StatusOK, "server/index", gin.H{
			"projects": new(models.Project).Find(),
		})
	}

}

func (s *Server) Add(c *gin.Context) {
	if core.IsAjax(c) {
		var server models.Server
		if err := c.ShouldBind(&server); err == nil {
			if server.Auth == models.ServerAuthClose {
				server.Username, server.Password = "", ""
			}
			if server.Auth == models.ServerAuthOpen && (server.Username == "" || server.Password == "") {
				s.Fail(c, config.PromptMsg["server_username_error"])
				return
			}
			if ok, error := server.InsertOne(); !ok {
				s.Fail(c, config.PromptMsg[error])
				return
			}
		} else {
			s.Fail(c, config.PromptMsg["parameter_error"])
			return
		}
		s.Success(c)
	} else {
		c.HTML(http.StatusOK, "server/add", gin.H{
			"projects": new(models.Project).Find(),
		})
	}
}
