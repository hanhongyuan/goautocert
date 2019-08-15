package models

import (
	"github.com/go-xorm/xorm"
	"time"
)

type Certificate struct {
	Id int `json:"id" xorm:"pk autoincr notnull "`
	// xxx.net.json
	Domain        string `xorm:"varchar(128)  not null" json:"domain"`
	CertUrl       string `xorm:"varchar(256) " json:"cert_url"`
	CertStableUrl string `xorm:"varchar(256) " json:"cert_stable_url"`
	// xxx.net.crt
	Certificate string `xorm:"varchar(5120) " json:"certificate"`
	// xxx.net.issuer.crt
	IssuerCertificate string `xorm:"varchar(5120) " json:"issuer_certificate"`
	// xxx.net.key
	PrivateKey string `xorm:"varchar(5120) " json:"private_key"`

	BaseModel `json:"-" xorm:"-"`

	Created time.Time `json:"created" xorm:"datetime notnull created"`
}

// 新增
func (c *Certificate) Create() (insertId int, err error) {
	_, err = Db.Insert(c)
	if err == nil {
		insertId = c.Id
	}

	return
}

func (c *Certificate) UpdateBean(id int16) (int64, error) {
	return Db.ID(id).Cols("domain,cert_url,cert_stable_url,certificate,issuer_certificate,private_key").Update(c)
}

// 更新
func (c *Certificate) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(c).ID(id).Update(data)
}

// 删除
func (c *Certificate) Delete(id int) (int64, error) {
	return Db.Id(id).Delete(new(Certificate))
}

func (c *Certificate) Find(id int) error {
	_, err := Db.Id(id).Get(c)

	return err
}

func (c *Certificate) DomainExists(domain string, id int16) (bool, error) {
	if id == 0 {
		count, err := Db.Where("domain = ?", domain).Count(c)
		return count > 0, err
	}

	count, err := Db.Where("domain = ? AND id != ?", domain, id).Count(c)
	return count > 0, err
}

func (c *Certificate) List(params CommonMap) ([]Certificate, error) {
	c.parsePageAndPageSize(params)
	list := make([]Certificate, 0)
	session := Db.Desc("id")
	c.parseWhere(session, params)
	err := session.Limit(c.PageSize, c.pageLimitOffset()).Find(&list)

	return list, err
}

func (c *Certificate) AllList() ([]Certificate, error) {
	list := make([]Certificate, 0)
	err := Db.Cols("domain,cert_url,cert_stable_url,certificate,issuer_certificate,private_key").Desc("id").Find(&list)

	return list, err
}

func (c *Certificate) Total(params CommonMap) (int64, error) {
	session := Db.NewSession()
	c.parseWhere(session, params)
	return session.Count(c)
}

// 解析where
func (c *Certificate) parseWhere(session *xorm.Session, params CommonMap) {
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
