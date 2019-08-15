package letsencrypt

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/ouqiang/gocron/internal/models"
	"time"
)

func UpCertificate2AliyunSLB(key models.AccessKey, config models.DomainConfig, aslb models.AliyunSLB, certificate models.Certificate) (err error) {
	client, err := slb.NewClientWithAccessKey(aslb.RegionId, key.AccessKeyId, key.AccessKeySecret)
	if err != nil {
		return err
	}
	serverCertificateId := ""
	if cs, err := describeServerCertificates(client); err != nil {
		return err
	} else if len(cs) > 0 {
		// 返回阿里云的证书列表， 从这些列表中找到目标域名的  serverCertificateId， 并判断是否过期如果过期了就需要删除。

		expireTimeStamp := int64(86400 * config.DefaultRenewDay)
		// 剩余过期时间
		var lastExpireTimeStamp int64

		for i := 0; i < len(cs); i++ {
			for j := 0; j < len(cs[i].SubjectAlternativeNames.SubjectAlternativeName); j++ {
				// 查找域名
				if cs[i].SubjectAlternativeNames.SubjectAlternativeName[j] == certificate.Domain {
					serverCertificateId = cs[i].ServerCertificateId
					lastExpireTimeStamp = cs[i].ExpireTimeStamp

					now := time.Now() // current local time
					sec := now.Unix() * 100
					if (lastExpireTimeStamp - sec) > expireTimeStamp {
						return fmt.Errorf("域名正常")
					}

					// 域名的证书已经过期了或快过期了（小于检查时间）

					goto OutFor
				}
			}
		}
	OutFor:
		// 删除证书
		if err = deleteServerCertificate(client, serverCertificateId); err != nil {
			return err
		}
		serverCertificateId = ""
		// 上传证书

	}
	if len(serverCertificateId) == 0 {
		// 上传证书
		if serverCertificateId, err = uploadServerCertificate(client, certificate); err != nil {
			return
		}
	}

	// 监听状态定义参考 https://help.aliyun.com/document_detail/27607.html
	var loadBalancerHTTPSListenerStatus string
	if loadBalancerHTTPSListenerStatus, err = describeLoadBalancerHTTPSListenerAttribute(client, aslb.ListenerPort, aslb.LoadBalancerId); err != nil {
		return
	} else if loadBalancerHTTPSListenerStatus == "" {
		// 监听不存在，就需要新建一个https监听
		if err = createLoadBalancerHTTPSListener(client, aslb.ListenerPort, aslb.BackendServerPort, aslb.LoadBalancerId, serverCertificateId); err != nil {
			return
		}
	} else {
		// 监听存在就直接修改这个监听的证书id
		if err = setLoadBalancerHTTPSListenerAttribute(client, aslb.ListenerPort, aslb.LoadBalancerId, serverCertificateId); err != nil {
			return
		}
	}

	if loadBalancerHTTPSListenerStatus != "running" {
		// 通过工单与官方沟通得知，更换slb的证书不用重启监听。这个消息非常棒！！！

		// 启动监听
		if err = startLoadBalancerListener(client, aslb.ListenerPort, aslb.LoadBalancerId); err != nil {
			return
		}
	}

	return nil
}

// 修改意见存在的监听的 ssl证书id
func setLoadBalancerHTTPSListenerAttribute(client *slb.Client, port int, loadBalancerId, serverCertificateId string) (err error) {
	request := slb.CreateSetLoadBalancerHTTPSListenerAttributeRequest()
	request.Scheme = "https"

	request.ListenerPort = requests.NewInteger(port)
	request.LoadBalancerId = loadBalancerId
	request.ServerCertificateId = serverCertificateId

	response, err := client.SetLoadBalancerHTTPSListenerAttribute(request)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	fmt.Printf("response is %#v\n", response)
	return nil
}

// 启动监听
func startLoadBalancerListener(client *slb.Client, port int, loadBalancerId string) (err error) {
	request := slb.CreateStartLoadBalancerListenerRequest()
	request.Scheme = "https"

	request.ListenerPort = requests.NewInteger(port)
	request.LoadBalancerId = loadBalancerId

	response, err := client.StartLoadBalancerListener(request)
	if err != nil {
		fmt.Print(err.Error())
		return err
	} else if response.IsSuccess() {
		return
	}
	return fmt.Errorf("启动监听(%d/%s)失败,错误：%s", port, loadBalancerId, response.GetHttpContentString())
}

// 上传证书，返回证书ID
func uploadServerCertificate(client *slb.Client, certificate models.Certificate) (serverCertificateId string, err error) {

	request := slb.CreateUploadServerCertificateRequest()
	request.Scheme = "https"

	request.ServerCertificate = certificate.Certificate
	request.PrivateKey = certificate.PrivateKey
	request.ServerCertificateName = certificate.Domain

	response, err := client.UploadServerCertificate(request)
	if err != nil {
		return "", err
	}
	return response.ServerCertificateId, nil
}

// 删除指定的证书
func deleteServerCertificate(client *slb.Client, serverCertificateId string) (err error) {
	request := slb.CreateDeleteServerCertificateRequest()
	request.Scheme = "https"

	response, err := client.DeleteServerCertificate(request)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	fmt.Printf("response is %#v\n", response)
	return nil
}

// 获取负载均衡的指定端口详情
func describeLoadBalancerHTTPSListenerAttribute(client *slb.Client, port int, loadBalancerId string) (exist string, err error) {

	request := slb.CreateDescribeLoadBalancerHTTPSListenerAttributeRequest()
	request.Scheme = "https"

	request.ListenerPort = requests.NewInteger(port)
	request.LoadBalancerId = loadBalancerId

	response, err := client.DescribeLoadBalancerHTTPSListenerAttribute(request)
	if err != nil {
		if cerr, ok := err.(*errors.ServerError); ok && cerr.ErrorCode() == "InvalidParameter" {

			fmt.Print("InvalidParameter")
			return "", nil
		} else {
			return "", err
		}
	}

	fmt.Printf("response is %#v\n", response)
	return response.Status, nil
}

//  对指定负载均衡添加监听；
func createLoadBalancerHTTPSListener(client *slb.Client, port, backendServerPort int, loadBalancerId, serverCertificateId string) (err error) {

	request := slb.CreateCreateLoadBalancerHTTPSListenerRequest()
	request.Scheme = "https"

	request.ListenerPort = requests.NewInteger(port)
	request.ServerCertificateId = serverCertificateId
	request.LoadBalancerId = loadBalancerId

	// 下面这些都是必选参数，这儿都设置一个默认值！参数说明见：https://help.aliyun.com/document_detail/27593.html
	request.Bandwidth = requests.NewInteger(-1)
	request.StickySession = "on"                                       // session会话粘连/会话保持
	request.HealthCheck = "off"                                        // 监控检查
	request.BackendServerPort = requests.NewInteger(backendServerPort) // 后端端口，swarm容器服务器默认使用9080端口。
	request.StickySessionType = "insert"                               // session保存使用的cookie方式
	request.CookieTimeout = requests.NewInteger(86400)                 // cookie超时时间
	// end

	response, err := client.CreateLoadBalancerHTTPSListener(request)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	fmt.Printf("response is %#v\n", response)
	return nil
}

// 查询证书列表
func describeServerCertificates(client *slb.Client) (cs []slb.ServerCertificate, err error) {
	request := slb.CreateDescribeServerCertificatesRequest()
	request.Scheme = "https"

	response, err := client.DescribeServerCertificates(request)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	fmt.Printf("response is %#v\n", response)
	return response.ServerCertificates.ServerCertificate, nil
}
