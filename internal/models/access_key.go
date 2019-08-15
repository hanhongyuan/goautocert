package models

import (
	"github.com/go-xorm/xorm"
	"time"
)

// 阿里云的 AccessKey
type AccessKey struct {
	Id int `xorm:"pk autoincr notnull " json:"id"`

	//  key
	AccessKeyId string `xorm:"varchar(64) notnull" json:"access_key_id"`

	// secret
	AccessKeySecret string `xorm:"varchar(128) notnull" json:"access_key_secret"`

	// 备注
	Remark string `xorm:"varchar(128) " json:"remark"`

	BaseModel `json:"-" xorm:"-"`

	Created time.Time `json:"created" xorm:"datetime notnull created"`
}

// 新增
func (c *AccessKey) Create() (insertId int, err error) {
	_, err = Db.Insert(c)
	if err == nil {
		insertId = c.Id
	}

	return
}

func (c *AccessKey) UpdateBean(id int16) (int64, error) {
	return Db.ID(id).Cols("access_key_id,access_key_secret,remark").Update(c)
}

// 更新
func (c *AccessKey) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(c).ID(id).Update(data)
}

// 删除
func (c *AccessKey) Delete(id int) (int64, error) {
	return Db.Id(id).Delete(new(AccessKey))
}

func (c *AccessKey) Find(id int) error {
	_, err := Db.Id(id).Get(c)

	return err
}

func (c *AccessKey) KeyAndSecretExists(key, value string) (bool, error) {
	count, err := Db.Where("access_key_id = ? and access_key_secret = ?", key, value).Count(c)
	return count > 0, err
}

func (c *AccessKey) List(params CommonMap) ([]AccessKey, error) {
	c.parsePageAndPageSize(params)
	list := make([]AccessKey, 0)
	session := Db.Desc("id")
	c.parseWhere(session, params)
	err := session.Limit(c.PageSize, c.pageLimitOffset()).Find(&list)

	return list, err
}

func (c *AccessKey) AllList() ([]AccessKey, error) {
	list := make([]AccessKey, 0)
	err := Db.Cols("access_key_id,access_key_secret,remark").Desc("id").Find(&list)

	return list, err
}

func (c *AccessKey) Total(params CommonMap) (int64, error) {
	session := Db.NewSession()
	c.parseWhere(session, params)
	return session.Count(c)
}

// 解析where
func (c *AccessKey) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		session.And("id = ?", id)
	}
	key, ok := params["AccessKeyId"]
	if ok && key.(string) != "" {
		session.And("access_key_id = ?", key)
	}
	secret, ok := params["AccessKeySecret"]
	if ok && secret.(string) != "" {
		session.And("access_key_secret = ?", secret)
	}
}
