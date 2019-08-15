package models

import (
	"github.com/go-xorm/xorm"
	"time"
)

// aliyun 负载均衡的基本配置，完整配置数据项太多，自己到界面上去配。
type AliyunSLB struct {
	Id int `json:"id" xorm:"pk autoincr notnull "`

	// regionId 区域，如：杭州、深圳
	RegionId string `xorm:"varchar(64)  not null" json:"region_id"`

	// SLB实例id
	LoadBalancerId string `xorm:"varchar(64)  not null" json:"load_balancer_id"`

	// 监听端口
	ListenerPort int `xorm:"default 443" json:"listener_port"`
	// 后端端口
	BackendServerPort int `xorm:"default 9080 " json:"backend_server_port"`

	BaseModel `json:"-" xorm:"-"`

	Created time.Time `json:"created" xorm:"datetime notnull created"`
}

// 新增
func (c *AliyunSLB) Create() (insertId int, err error) {
	_, err = Db.Insert(c)
	if err == nil {
		insertId = c.Id
	}

	return
}

func (c *AliyunSLB) UpdateBean(id int16) (int64, error) {
	return Db.ID(id).Cols("region_id,load_balancer_id,listener_port,backend_server_port").Update(c)
}

// 更新
func (c *AliyunSLB) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(c).ID(id).Update(data)
}

// 删除
func (c *AliyunSLB) Delete(id int) (int64, error) {
	return Db.Id(id).Delete(new(AliyunSLB))
}

func (c *AliyunSLB) Find(id int) error {
	_, err := Db.Id(id).Get(c)

	return err
}

func (c *AliyunSLB) DomainExists(domain string, id int16) (bool, error) {
	if id == 0 {
		count, err := Db.Where("domain = ?", domain).Count(c)
		return count > 0, err
	}

	count, err := Db.Where("domain = ? AND id != ?", domain, id).Count(c)
	return count > 0, err
}

func (c *AliyunSLB) List(params CommonMap) ([]AliyunSLB, error) {
	c.parsePageAndPageSize(params)
	list := make([]AliyunSLB, 0)
	session := Db.Desc("id")
	c.parseWhere(session, params)
	err := session.Limit(c.PageSize, c.pageLimitOffset()).Find(&list)

	return list, err
}

func (c *AliyunSLB) AllList() ([]AliyunSLB, error) {
	list := make([]AliyunSLB, 0)
	err := Db.Cols("region_id,load_balancer_id,listener_port,backend_server_port").Desc("id").Find(&list)

	return list, err
}

func (c *AliyunSLB) Total(params CommonMap) (int64, error) {
	session := Db.NewSession()
	c.parseWhere(session, params)
	return session.Count(c)
}

// 解析where
func (c *AliyunSLB) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		session.And("id = ?", id)
	}
	regionId, ok := params["RegionId"]
	if ok && regionId.(string) != "" {
		session.And("region_id = ?", regionId)
	}
	balancerId, ok := params["LoadBalancerId"]
	if ok && balancerId.(string) != "" {
		session.And("load_balancer_id = ?", balancerId)
	}
}
