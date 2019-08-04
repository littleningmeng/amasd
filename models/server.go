package models

import (
	"scrapyd-admin/core"
	"github.com/ltachi1/logrus"
	"fmt"
	"time"
	"strconv"
)

type Server struct {
	Base       core.BaseModel `json:"-" xorm:"-"`
	Id         int            `json:"id" xorm:"pk autoincr"`
	Alias      string         `json:"alias"`
	Host       string         `json:"host" binding:"required"`
	Username   string         `json:"-"`
	Password   string         `json:"-"`
	Auth       uint8          `json:"auth"`
	Status     uint8          `json:"status"`
	CreateTime core.Timestamp `json:"create_time" xorm:"created"`
	UpdateTime core.Timestamp `json:"update_time" xorm:"created updated"`
}

var (
	ServerStatusNormal uint8 = 1 //服务器状态正常
	ServerStatusFault  uint8 = 2 //服务器状态故障
	ServerAuthClose    uint8 = 1 //服务器验证关闭
	ServerAuthOpen     uint8 = 2 //服务器验证开启
)

//添加服务器
func (s *Server) InsertOne() (bool, string) {
	if s.Host[len(s.Host) - 1:] == "/" {
		s.Host = s.Host[:len(s.Host) - 1]
	}
	//校验服务器是否已经添加过
	if count, _ := core.Db.Where("host = ?", s.Host).Table(s).Count(); count > 0 {
		return false, "host_repeat_error"
	}
	//检验服务器是否可用
	scrapyd := Scrapyd{Host: s.Host, Auth: s.Auth, Username: s.Username, Password: s.Password}
	if err := scrapyd.DaemonStatus(); err != nil {
		return false, "scrapyd_server_error"
	}

	if _, error := core.Db.InsertOne(s); error != nil {
		core.WriteLog(core.LogTypeServer, logrus.ErrorLevel, logrus.Fields{"host": s.Host}, error)
		return false, "add_error"
	}
	return true, ""
}

//获取所有服务器
func (s *Server) Find() []Server {
	servers := make([]Server, 0)
	core.Db.OrderBy("id asc").Find(&servers, s)
	return servers
}

func (s *Server) FindByIds(ids []int) []Server {
	servers := make([]Server, 0)
	core.Db.In("id", ids).OrderBy("id asc").Find(&servers)
	return servers
}

//根据项目id获取所有服务器
func (s *Server) FindByProjectId(projectId int) []Server {
	servers := make([]Server, 0)
	core.Db.Table("server").Alias("s").Join("INNER", "project_server as ps", "ps.server_id = s.id").Where("ps.project_id = ?", projectId).Find(&servers, s)
	return servers
}

//根据项目id获取所有未拥有此项目的服务器
func (s *Server) FindByProjectIdNotProject(projectId int) []Server {
	servers := make([]Server, 0)
	core.Db.Table("server").Alias("s").Join("INNER", "project_server as ps", "ps.server_id = s.id").Where("ps.project_id != ?", projectId).Find(&servers, s)
	return servers
}

//分页获取服务器数据
func (s *Server) PageList(projectId int, page int, pageSize int) ([]Server, int) {
	servers := make([]Server, 0)
	var totalCount int64 = 0
	if projectId == 0 {
		totalCount, _ = core.Db.Table("server").Count()
		core.Db.OrderBy("update_time desc").Limit(pageSize, (page-1)*pageSize).Find(&servers)
	} else {
		totalCount, _ = core.Db.Table("server").Alias("s").Join("INNER", "project_server as ps", "ps.server_id = s.id").Where("ps.project_id = ?", projectId).Count()
		core.Db.Select("s.*").Table("server").Alias("s").Join("INNER", "project_server as ps", "ps.server_id = s.id").Where("ps.project_id = ?", projectId).Limit(pageSize, (page-1)*pageSize).OrderBy("s.update_time desc").Find(&servers)
	}
	return servers, int(totalCount)
}

func (s *Server) Get(id int) bool {
	ok, _ := core.Db.Id(id).NoAutoCondition().Get(s)
	return ok
}

func (s *Server) Update(id int, data core.B) error {
	data["update_time"] = strconv.Itoa(int(time.Now().Unix()))
	_, error := core.Db.Table(s).ID(id).Update(data)
	return error
}

func (s *Server) Del(id int) (bool, string) {
	if !s.Get(id) {
		return false, "server_info_error"
	}
	//如果项目下有正在运行的定时任务或者普通任务则不允许删除
	if new(Task).HaveRunningByServer(id) || new(SchedulesTask).HaveEnabledByServer(id) {
		return false, "server_del_task_running_error"
	}
	//移除服务器时不做移除项目的处理
	session := core.Db.NewSession()
	defer session.Close()
	session.Begin()
	//删除关联服务器
	if _, err := session.Where("server_id = ?", id).NoAutoCondition().Delete(&ProjectServer{}); err != nil {
		fmt.Println(1)
		return false, "del_error"
	}
	//删除所有定时任务
	if _, err := session.Where("server_id = ?", id).NoAutoCondition().Delete(&SchedulesTask{}); err != nil {
		fmt.Println(2)
		return false, "del_error"
	}
	//删除所有任务
	if _, err := session.Where("server_id = ?", id).NoAutoCondition().Delete(&Task{}); err != nil {
		fmt.Println(3)
		return false, "del_error"
	}
	if _, err := session.Where("id = ?", id).NoAutoCondition().Delete(&Server{}); err != nil {
		return false, "del_error"
	}
	session.Commit()

	return true, ""
}

//检测服务器状态并更新
func (s *Server) DetectionStatus() {
	serverList := s.Find()
	for _, server := range serverList {
		go func(server Server) {
			scrapyd := Scrapyd{
				Host:     server.Host,
				Auth:     server.Auth,
				Username: server.Username,
				Password: server.Password,
			}
			if err := scrapyd.DaemonStatus(); err == nil {
				if server.Status == ServerStatusFault {
					core.WriteLog(core.LogTypeServer, logrus.InfoLevel, logrus.Fields{"host": server.Host}, "scrapyd服务恢复正常")
					if _, error := core.Db.Id(server.Id).Update(&Server{Status: ServerStatusNormal}); error != nil {
						core.WriteLog(core.LogTypeServer, logrus.ErrorLevel, logrus.Fields{"host": server.Host}, fmt.Sprintf("scrapyd服务状态更新失败:%s", error))
					}
				}
			} else {
				if server.Status == ServerStatusNormal {
					core.WriteLog(core.LogTypeServer, logrus.ErrorLevel, logrus.Fields{"host": server.Host}, fmt.Sprintf("scrapyd服务异常:%s", err))
					if _, error := core.Db.Id(server.Id).Update(&Server{Status: ServerStatusFault}); error != nil {
						core.WriteLog(core.LogTypeServer, logrus.ErrorLevel, logrus.Fields{"host": server.Host}, fmt.Sprintf("scrapyd服务状态更新失败:%s", error))
					}
				}
			}
		}(server)
	}
}
