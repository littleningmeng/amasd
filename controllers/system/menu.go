package system

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"scrapyd-admin/config"
	"scrapyd-admin/core"
	"scrapyd-admin/models"
	"strconv"
)

type Menu struct {
	core.BaseController
}

func (a *Menu) Index(c *gin.Context) {
	if core.IsAjax(c) {
		c.JSON(http.StatusOK, gin.H{
			"data": new(models.SystemMenu).TreeMenus(),
		})
	} else {
		c.HTML(http.StatusOK, "menu/index", gin.H{})
	}
}

//添加菜单
func (a *Menu) Add(c *gin.Context) {
	if core.IsAjax(c) {
		var menu models.SystemMenu
		if err := c.ShouldBind(&menu); err == nil {
			if menu.Insert() {
				a.Success(c)
			} else {
				a.Fail(c, config.PromptMsg["add_error"])
			}
		} else {
			a.Fail(c, config.PromptMsg["add_error"])
		}
	} else {
		parentId, _ := strconv.Atoi(c.DefaultQuery("parent_id", "0"))
		c.HTML(http.StatusOK, "menu/add", gin.H{
			"parentId":  parentId,
			"treeMenus": new(models.SystemMenu).TreeMenus(),
		})
	}
}

//编辑菜单
func (a *Menu) Edit(c *gin.Context) {
	if core.IsAjax(c) {
		id, _ := strconv.Atoi(c.DefaultPostForm("id", "0"))
		parentId, _ := strconv.Atoi(c.DefaultPostForm("parent_id", ""))
		name := c.DefaultPostForm("name", "")
		app := c.DefaultPostForm("app", "")
		controller := c.DefaultPostForm("controller", "")
		action := c.DefaultPostForm("action", "")
		parameter := c.DefaultPostForm("parameter", "")
		icon := c.DefaultPostForm("icon", "")
		status, _ := strconv.Atoi(c.DefaultPostForm("status", ""))
		if id == 0 {
			a.Fail(c, config.PromptMsg["parameter_error"])
			return
		}
		if name == "" {
			a.Fail(c, config.PromptMsg["system_menu_name_error"])
			return
		}
		if app == "" {
			a.Fail(c, config.PromptMsg["system_menu_app_error"])
			return
		}
		if controller == "" {
			a.Fail(c, config.PromptMsg["system_menu_controller_error"])
			return
		}
		if action == "" {
			a.Fail(c, config.PromptMsg["system_menu_action_error"])
			return
		}
		if !(status == models.MenuStatusEnable || status == models.MenuStatusDisable) {
			a.Fail(c, config.PromptMsg["system_menu_status_error"])
			return
		}

		err := new(models.SystemMenu).Update(id, core.A{
			"parent_id":  parentId,
			"name":       name,
			"app":        app,
			"controller": controller,
			"action":     action,
			"parameter":  parameter,
			"icon":       icon,
			"status":     status,
		})
		if err == nil {
			a.Success(c)
		} else {
			a.Fail(c, config.PromptMsg["update_error"])
		}
	} else {
		id, _ := strconv.Atoi(c.DefaultQuery("id", "0"))
		if id == 0 {
			core.Error(c, config.PromptMsg["parameter_error"])
			return
		}
		sm := new(models.SystemMenu)
		if ok := sm.Get(id); ok && sm.Id == id {
			c.HTML(http.StatusOK, "menu/edit", gin.H{
				"info":      sm,
				"treeMenus": sm.TreeMenus(),
			})
		} else {
			core.Error(c, config.PromptMsg["system_menu_info_error"])
		}

	}
}

//修改状态
func (a *Menu) EditStatus(c *gin.Context) {
	if core.IsAjax(c) {
		id, _ := strconv.Atoi(c.DefaultQuery("id", "0"))
		status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
		if !(status == models.MenuStatusEnable || status == models.MenuStatusDisable) {
			a.Fail(c, config.PromptMsg["parameter_error"])
			return
		}
		if error := new(models.SystemMenu).Update(id, core.A{"status": status}); error == nil {
			a.Success(c)
		} else {
			a.Fail(c, config.PromptMsg["update_error"])
		}
	}
}

//删除菜单
func (a *Menu) Del(c *gin.Context) {
	id, _ := strconv.Atoi(c.DefaultQuery("id", "0"))
	if id == 0 {
		a.Fail(c, config.PromptMsg["parameter_error"])
		return
	}
	sm := new(models.SystemMenu)
	sm.Id = id
	if ok := sm.DeleteById(); ok {
		a.Success(c)
	} else {
		a.Fail(c, config.PromptMsg["del_error"])
	}

}
