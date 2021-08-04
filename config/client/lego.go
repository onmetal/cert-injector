/*
Copyright (c) 2021 T-Systems International GmbH, SAP SE or an SAP affiliate company. All right reserved
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"log"
)

const secret = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAycJwJ3EPeBI49xt5BVD1unliiGftWZEiBKaG8WoGxOuKCkKS
griyL/BSdQd9+LSEn1g02q2cdOSlD+lYSv1s5gvundpBiJQXdZgHhPCN7rmpW2Xv
pZ0oHbF6qm6pgPRTurUf4Qud1bkUXGrUUAy6AtMPwkTOBX3ta3IFCPV9osFMsghp
Onng0hQ8cbxhk3YFZxx0D2Dfbc66zPMS+CqfOVxP46EcbneemRTF7MyZx5Cv3plw
KscZVGi6sztCpJIgNJDKjIv2UjoUAvzRVqzaymt/hwozKIl5+Qs+01jV9bWHJOoD
LHyIOACZCFu1XNoQQ69hJVfk9mKOb3qrpCJUfQIDAQABAoIBAQC3mJUgfxS5ibOG
wdw1xz9k6hKM2C23JIeVPchsJLR2O3RI892I0PNtBj6yuhea2wIYUlb+a5+FC49c
1FWBH+4ZxN/live5hjF209p70b8GbrK7Nh6GUWVw59EdCEh8zVjn/Ow+iKifFKV/
l8MN+RbHfTLI8H2dp8MF1CLazTH/iFMvCMv0Ke7FajlUCcqSqCYjqNAh2YlD4Q0Q
Rqc2/M+N86XGe5+q7l6Q/SzbAiWSLUoOBif/c8gVVJqivago1AoRfFE447i4XtY2
A9dhvlD8oNcC4gAurvtsmYSyHy3tYFOPvEgUsNKJhftIi+WJNEcaRMgCTgqqWp4O
dDE1C9PBAoGBAN0BQDGC2g1xt9Pgh4lu6vfB43lniwMhPThAEoiTfW3B0e7QGkVg
npW5Awm2DWjLYgcrNWzrbrAMyKcb38bhllDDFlKSKDzziNiiIQCyhkSfo/HkVKdj
iepcBZoRNIoH662HrnezeMhMcuU71PcME6FheeKZYh7g8PHwVzDbkRgtAoGBAOm1
DI+OxHn8K7xARKsr0OvPyMZXukVyCZeqFMVz8gGUTfy9V3uoPJMSJfZ5BVbRqg71
ZuWxvr26dX99TvHwrqRcMIZjOGRANB/1JuDu/Wd414PMTOvXELbCfQmoiWe4pzM9
abwAR8UUEcq6Wb0y8kbdhZtLTMKINMu3IFslTQ+RAoGAQF2P45uXhBjdkBCxiL5M
IpJOfNpCK0wv90T54NsLyb6MNMBZFmGYbkSu9NIXv7CUQUA9VBaRayad/cVpfBPR
Yn4e7zdwDqhi76zwbbKQ1kWkStvUJ9gen6njW8atBZJe+nAsyOH1SGizgb3WPYk/
4l1wUSWY5SNgKSZ1Tl50OJUCgYAe8UahdzCSSg3sVcIBu8JkhlU51YGnEiss9mrb
nbdL+Du/G76Kc8LZYgy+rlVDomzWoC0oejkb26UU5R1fsRMeVcpi8J4Vv95m4Mlt
/JZ2baxzGciRbR8cY3G0pqjSn8MbaKUoLA1UjYyxf6zD/QvQ0CGRZw3Zr7j1w+A3
0R970QKBgQDSLwzIGt/RrIcuk8aQoZMcERQLOlRhUt/E3/5mTexZVXUaVvv5NwOt
Z4FdJihE8Z5wGvlH27atTIKCfAs9+w0ZOz1R+frP/M8iHfRMNte1BV5g7tl4aCHd
Lxl3aP4wPQZkSfgK8OX7oPGNto5pUTDn9HUobNABmhhNmtooeEWlxQ==
-----END RSA PRIVATE KEY-----`

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func main() {

	// Create a user. New accounts need an email and private key to start.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	//key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}
	//s := []byte(secret)
	//key, err := certcrypto.ParsePEMPrivateKey(s)
	//if err != nil {
	//	log.Fatal(err)
	//}

	myUser := MyUser{
		Email: "you@yours.com",
		key:   key,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = "https://localhost:14000/dir"

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "8080"))
	if err != nil {
		log.Fatal(err)
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"kubernetes.docker.internal"},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	//fmt.Printf("%#v\n", certificates)
	k, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Fatal(err)
	}
	renew(*certificates, k)
}

func renew(certificat certificate.Resource, k []byte) {
	log.Println("renewing in process")
	//k := []byte(secret)
	key, err := x509.ParseECPrivateKey(k)
	//key, err := certcrypto.ParsePEMPrivateKey(k)
	if err != nil {
		log.Fatal(err)
	}

	myUser := MyUser{
		Email: "you@yours.com",
		key:   key,
		//Registration: reg,
	}

	config := lego.NewConfig(&myUser)
	config.CADirURL = "https://localhost:14000/dir"

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	newRega, err := client.Registration.ResolveAccountByKey()
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = newRega
	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "8080"))
	if err != nil {
		log.Fatal(err)
	}
	var cert certificate.Resource
	cert.PrivateKey = certificat.PrivateKey
	cert.Certificate = certificat.Certificate
	newCert, err := client.Certificate.Renew(cert, true, false, "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v\n", newCert.CertURL)
}
