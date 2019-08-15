package models

import (
	"github.com/go-xorm/xorm"
	"time"
)

type Bool int8

const (
	False Bool = iota     // false
	True  Bool = iota + 1 // true
)

// Domains Struct
type DomainConfig struct {
	Id int `json:"id" xorm:"pk autoincr notnull "`

	// 域名，支持多个，以空格隔开；如：  *.a.com  *.c.com bb.cn
	Domain string `xorm:"varchar(100) not null" json:"domain"`

	// Provider 名称，目前只支持dns-01：如：alidns
	ProviderName string `xorm:"varchar(32)  not null" json:"provider_name"`

	// 默认的续期时间。 如 小于多少天就续期
	DefaultRenewDay int `xorm:"default 90" json:"default_renew_day"`

	// 实现重用现有的私钥
	ReuseKey   Bool `xorm:"tinyint notnull default 0 " json:"reuse_key"`
	Bundle     Bool `xorm:"tinyint notnull default 0 " json:"bundle"`
	MustStaple Bool `xorm:"tinyint notnull default 0 " json:"must_staple"`

	BaseModel `json:"-" xorm:"-"`

	Created time.Time `json:"created" xorm:"datetime notnull created"`
}

// 新增
func (d *DomainConfig) Create() (insertId int, err error) {
	_, err = Db.Insert(d)
	if err == nil {
		insertId = d.Id
	}

	return
}

func (d *DomainConfig) HasReuseKey() bool {
	return d.ReuseKey == True
}

func (d *DomainConfig) HasBundle() bool {
	return d.Bundle == True
}

func (d *DomainConfig) HasMustStaple() bool {
	return d.MustStaple == True
}

func (d *DomainConfig) UpdateBean(id int16) (int64, error) {
	return Db.ID(id).Cols("domain,provider_name,default_renew_day,reuse_key,bundle,must_staple").Update(d)
}

// 更新
func (d *DomainConfig) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(d).ID(id).Update(data)
}

// 删除
func (d *DomainConfig) Delete(id int) (int64, error) {
	return Db.Id(id).Delete(new(DomainConfig))
}

func (d *DomainConfig) Find(id int) error {
	_, err := Db.Id(id).Get(d)

	return err
}

func (d *DomainConfig) DomainExists(domain string, id int16) (bool, error) {
	if id == 0 {
		count, err := Db.Where("domain = ?", domain).Count(d)
		return count > 0, err
	}

	count, err := Db.Where("domain = ? AND id != ?", domain, id).Count(d)
	return count > 0, err
}

func (d *DomainConfig) List(params CommonMap) ([]DomainConfig, error) {
	d.parsePageAndPageSize(params)
	list := make([]DomainConfig, 0)
	session := Db.Desc("id")
	d.parseWhere(session, params)
	err := session.Limit(d.PageSize, d.pageLimitOffset()).Find(&list)

	return list, err
}

func (d *DomainConfig) AllList() ([]DomainConfig, error) {
	list := make([]DomainConfig, 0)
	err := Db.Cols("domain,provider_name,default_renew_day,reuse_key,bundle,must_staple").Desc("id").Find(&list)

	return list, err
}

func (d *DomainConfig) Total(params CommonMap) (int64, error) {
	session := Db.NewSession()
	d.parseWhere(session, params)
	return session.Count(d)
}

// 解析where
func (d *DomainConfig) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		session.And("id = ?", id)
	}
	name, ok := params["Domain"]
	if ok && name.(string) != "" {
		session.And("domain = ?", name)
	}
}
