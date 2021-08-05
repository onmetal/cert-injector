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
	if err != nil {
		log.Fatal(err)
	}

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

	key, err := x509.ParseECPrivateKey(k)
	if err != nil {
		log.Fatal(err)
	}

	myUser := MyUser{
		Email: "you@yours.com",
		key:   key,
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
