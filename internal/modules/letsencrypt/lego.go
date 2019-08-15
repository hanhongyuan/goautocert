package letsencrypt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/certificate"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"
	"github.com/ouqiang/gocron/internal/models"
	"github.com/ouqiang/gocron/internal/modules/logger"
	"time"

	"log"
)

type tmpUser struct {
	AcmeUser     *models.AcmeUser
	PrivateKey   crypto.PrivateKey
	Registration *registration.Resource
}

func (u *tmpUser) GetEmail() string {
	return u.AcmeUser.Email
}
func (u *tmpUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *tmpUser) GetPrivateKey() crypto.PrivateKey {
	return u.PrivateKey
}

// user类型转换
// newUser bool 表示是否是新用户，如果是新用户就需要创建private key
func getUserByAcmeUser(au *models.AcmeUser, newUser bool) (u *tmpUser, err error) {
	var pk crypto.PrivateKey
	var r registration.Resource
	if len(au.PrivateKey) > 0 {
		if pk, err = certcrypto.ParsePEMBundle([]byte(au.PrivateKey)); err != nil {
			logger.Warnf("Error while loading the certificate for domain %s\n\t%v", au.Email, err)
			return
		}
	} else if newUser { // 新用户
		// Create a tmpUser. New accounts need an email and private key to start.
		if pk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
			logger.Error(err)
			return
		}
	} else {
		if len(au.Resource) > 0 {
			if err = json.Unmarshal([]byte(au.Resource), &r); err != nil {
				return
			}
		} else {
			return nil, fmt.Errorf("缺少用户的Account数据，请检查数据是否完整！！！")
		}
	}
	u = &tmpUser{au, pk, &r}
	return
}
func newConfig(user registration.User) *lego.Config {
	vconfig := lego.NewConfig(user)
	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	// config.CADirURL = "http://192.168.99.100:4000/directory"
	// vconfig.CADirURL = "https://acme-v02.api.letsencrypt.org/directory" // 正式api， 有限制：https://letsencrypt.org/docs/rate-limits/
	vconfig.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory" // letsencrypt官方测试api，说明：https://letsencrypt.org/docs/staging-environment/
	vconfig.Certificate.KeyType = certcrypto.RSA2048
	return vconfig
}

// 申请新证书
func ObtainCertificate(au *models.AcmeUser, config models.DomainConfig, ak models.AccessKey) (result *models.Certificate, err error) {
	myUser, err1 := getUserByAcmeUser(au, true)
	if err1 != nil {
		return nil, err1
	}

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(newConfig(myUser))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	provider, err := newDNSChallengeProviderByName(config, ak)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// We specify an http port of 5002 and an tls port of 5001 on all interfaces
	// because we aren't running as root and can't bind a listener to port 80 and 443
	// (used later when we attempt to pass challenges). Keep in mind that you still
	// need to proxy challenge traffic to port 5002 and 5001.
	// err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	myUser.Registration = reg
	// 存储用户的密钥信息
	_ = au.SaveResourceAndPrivateKey(myUser.PrivateKey, reg)

	request := certificate.ObtainRequest{
		Domains: []string{"*.scweiqu.com"},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	// fmt.Printf("%#v\n", certificates)
	result = &models.Certificate{
		Domain:            certificates.Domain,
		CertUrl:           certificates.CertURL,
		CertStableUrl:     certificates.CertStableURL,
		Certificate:       string(certificates.Certificate),
		PrivateKey:        string(certificates.PrivateKey),
		IssuerCertificate: string(certificates.IssuerCertificate),
	}
	return
}

// 证书续期
// defaultRenewDay int  续期天数
// reuseKey bool  实现重用现有的私钥
// bundle bool  ???
// mustStaple bool  ???
func RenewCertificate(au *models.AcmeUser, config models.DomainConfig, ak models.AccessKey, certificateData models.Certificate) (result *models.Certificate, err error) {
	domain := config.Domain

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certificates, err := certcrypto.ParsePEMBundle([]byte(certificateData.Certificate))
	if err != nil {
		log.Fatalf("Error while loading the certificate for domain %s\n\t%v", domain, err)
		return nil, err
	}

	cert := certificates[0]

	if !needRenewal(cert, domain, config.DefaultRenewDay) {
		return nil, fmt.Errorf("时间还充足，不允许续期证书！")
	}

	// This is just meant to be informal for the tmpUser.
	timeLeft := cert.NotAfter.Sub(time.Now().UTC())
	logger.Debugf("[%s] acme: Trying renewal with %d hours remaining", domain, int(timeLeft.Hours()))

	certDomains := certcrypto.ExtractDomains(cert)

	var privateKey crypto.PrivateKey
	if config.HasReuseKey() {
		if privateKey, err = certcrypto.ParsePEMPrivateKey([]byte(certificateData.PrivateKey)); err != nil {
			return nil, err
		}
	} else {
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	request := certificate.ObtainRequest{
		Domains:    certDomains,
		Bundle:     config.HasBundle(),
		PrivateKey: privateKey,
		MustStaple: config.HasMustStaple(),
	}

	myUser, err1 := getUserByAcmeUser(au, false)
	if err1 != nil {
		return nil, err1
	}

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(newConfig(myUser))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 提供dns-01 challenge
	provider, err := newDNSChallengeProviderByName(config, ak)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// end

	certRes, errR := client.Certificate.Obtain(request)
	if errR != nil {
		log.Fatal(errR)
		return nil, errR
	}
	result = &models.Certificate{
		Domain:            certRes.Domain,
		CertUrl:           certRes.CertURL,
		CertStableUrl:     certRes.CertStableURL,
		Certificate:       string(certRes.Certificate),
		PrivateKey:        string(certRes.PrivateKey),
		IssuerCertificate: string(certRes.IssuerCertificate),
	}
	return

}

// 证书注销
func RevokeCertificate(au *models.AcmeUser, config *models.DomainConfig, certificateData models.Certificate) (result *models.Certificate, err error) {

	log.Printf("Trying to revoke certificate for domain %s", certificateData.Domain)

	myUser, err1 := getUserByAcmeUser(au, false)
	if err1 != nil {
		return nil, err1
	}
	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(newConfig(myUser))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = client.Certificate.Revoke([]byte(certificateData.Certificate))
	if err != nil {
		log.Fatalf("Error while revoking the certificate for domain %s\n\t%v", certificateData.Domain, err)
	}

	log.Println("Certificate was revoked.")
	return
}

func needRenewal(x509Cert *x509.Certificate, domain string, days int) bool {
	if x509Cert.IsCA {
		log.Fatalf("[%s] Certificate bundle starts with a CA certificate", domain)
	}

	if days >= 0 {
		notAfter := int(time.Until(x509Cert.NotAfter).Hours() / 24.0)
		if notAfter > days {
			log.Printf("[%s] The certificate expires in %d days, the number of days defined to perform the renewal is %d: no renewal.",
				domain, notAfter, days)
			return false
		}
	}

	return true
}
func merge(prevDomains []string, nextDomains []string) []string {
	for _, next := range nextDomains {
		var found bool
		for _, prev := range prevDomains {
			if prev == next {
				found = true
				break
			}
		}
		if !found {
			prevDomains = append(prevDomains, next)
		}
	}
	return prevDomains
}
