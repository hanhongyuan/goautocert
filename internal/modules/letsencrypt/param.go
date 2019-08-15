package letsencrypt

import (
	"encoding/json"
	"fmt"
)

// 参数，以json格式存储在 task.command字段中。
type Param struct {
	AcmeUserId     int `json:"acme_user_id"`
	AccessKeyId    int `json:"access_key_id"`
	AliyunSLBId    int `json:"aliyun_slb_id"`
	CertificateId  int `json:"certificate_id"`
	DoaminConfigId int `json:"doamin_config_id"`
}

// 创建证书参数检查
func (p *Param) ValidationObtain() (err error) {
	return
}

// 创建证书参数检查
func (p *Param) ValidationRenew() (err error) {
	return
}

// 创建证书参数检查
func (p *Param) ValidationRevoke() (err error) {
	return
}

// 创建证书申请参数
func CreateObtainParam(command string) (p *Param, err error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("错误：参数无效！")
	}
	p = new(Param)
	if err = json.Unmarshal([]byte(command), &p); err != nil {
		return
	}

	return
}

// 续期证书申请参数
func CreateRenewParam(command string) (p *Param, err error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("错误：参数无效！")
	}
	p = new(Param)
	if err = json.Unmarshal([]byte(command), &p); err != nil {
		return
	}

	return
}

// 注销证书申请参数
func CreateRevokeParam(command string) (p *Param, err error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("错误：参数无效！")
	}
	p = new(Param)
	if err = json.Unmarshal([]byte(command), &p); err != nil {
		return
	}

	return
}
