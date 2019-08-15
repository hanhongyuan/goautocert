package models

import (
	"crypto"
	"encoding/json"
	"encoding/pem"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/registration"
	"github.com/go-xorm/xorm"
	"time"
)

// implements acme.User
type AcmeUser struct {
	Id int `json:"id" xorm:"pk autoincr notnull "`

	// 邮件地址
	Email string `json:"email" xorm:"varchar(64) notnull unique"`

	// 类似域名注册服务器的账号密码,
	PrivateKey string `xorm:"varchar(2048) " json:"private_key"`

	// 官方返回的账号信息，需要保存下来
	Resource string `xorm:"varchar(2048)  " json:"resource"`

	BaseModel `json:"-" xorm:"-"`

	Created time.Time `json:"created" xorm:"datetime notnull created"`
}

// 新增
func (c *AcmeUser) Create() (insertId int, err error) {
	_, err = Db.Insert(c)
	if err == nil {
		insertId = c.Id
	}

	return
}

func (c *AcmeUser) UpdateBean(id int) (int64, error) {
	return Db.ID(id).Cols("email,private_key,resource").Update(c)
}

// 更新
func (c *AcmeUser) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(c).ID(id).Update(data)
}

// 删除
func (c *AcmeUser) Delete(id int) (int64, error) {
	return Db.Id(id).Delete(new(AcmeUser))
}

func (c *AcmeUser) Find(id int) error {
	_, err := Db.Id(id).Get(c)

	return err
}

func (c *AcmeUser) EmailExists(domain string, id int16) (bool, error) {
	if id == 0 {
		count, err := Db.Where("email = ?", domain).Count(c)
		return count > 0, err
	}

	count, err := Db.Where("email = ? AND id != ?", domain, id).Count(c)
	return count > 0, err
}

func (c *AcmeUser) List(params CommonMap) ([]AcmeUser, error) {
	c.parsePageAndPageSize(params)
	list := make([]AcmeUser, 0)
	session := Db.Desc("id")
	c.parseWhere(session, params)
	err := session.Limit(c.PageSize, c.pageLimitOffset()).Find(&list)

	return list, err
}

func (c *AcmeUser) AllList() ([]AcmeUser, error) {
	list := make([]AcmeUser, 0)
	err := Db.Cols("email,private_key,resource").Desc("id").Find(&list)

	return list, err
}

func (c *AcmeUser) Total(params CommonMap) (int64, error) {
	session := Db.NewSession()
	c.parseWhere(session, params)
	return session.Count(c)
}

// 存储账号的用户名和密钥
func (c *AcmeUser) SaveResourceAndPrivateKey(privateKey crypto.PrivateKey, resource *registration.Resource) (err error) {
	pemKey := certcrypto.PEMBlock(privateKey)
	c.PrivateKey = string(pem.EncodeToMemory(pemKey))
	var d []byte
	if d, err = json.MarshalIndent(resource, "", "\t"); err != nil {
		return
	}
	c.Resource = string(d)
	_, err = c.UpdateBean(c.Id)
	return
}

// 解析where
func (c *AcmeUser) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		session.And("id = ?", id)
	}
	email, ok := params["Email"]
	if ok && email.(string) != "" {
		session.And("email = ?", email)
	}
}
