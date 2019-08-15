package service

import "github.com/ouqiang/gocron/internal/models"

// 证书申请执行任务
type CertificateObtainHandler struct{}

func (h *CertificateObtainHandler) Run(taskModel models.Task, taskUniqueId int64) (result string, err error) {
	return
}

// 证书续期执行任务
type CertificateRenewHandler struct{}

func (h *CertificateRenewHandler) Run(taskModel models.Task, taskUniqueId int64) (result string, err error) {
	return
}

// 证书注销执行任务
type CertificateRevokeHandler struct{}

func (h *CertificateRevokeHandler) Run(taskModel models.Task, taskUniqueId int64) (result string, err error) {
	return
}
