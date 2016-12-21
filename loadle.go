package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"log"
	"sync"

	"github.com/xenolf/lego/acme"
)

var n = flag.Int("n", 1, "Number of goroutines to spawn")
var m = flag.Int("m", 1, "Number of clients to create")

type MyUser struct {
	Registration *acme.RegistrationResource
	key          crypto.PrivateKey
}

func (u MyUser) GetEmail() string {
	return ""
}
func (u MyUser) GetRegistration() *acme.RegistrationResource {
	return u.Registration
}
func (u MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func main() {
	flag.Parse()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	acme.UserAgent = "loadle load tester"

	myUser := &MyUser{
		key: privateKey,
	}

	stagingURL := "https://acme-staging.api.letsencrypt.org/directory"
	client, err := acme.NewClient(stagingURL, myUser, acme.RSA2048)
	if err != nil {
		log.Fatal(err)
	}

	reg, err := client.Register()
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	err = client.AgreeToTOS()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	var clients []*acme.Client
	for i := 0; i < *m; i++ {
		// Use independent clients to avoid locking on nonce pool.
		client, err := acme.NewClient(stagingURL, myUser, acme.RSA2048)
		if err != nil {
			log.Fatal(err)
		}
		clients = append(clients, client)
	}

	for i := 0; i < *n; i++ {
		wg.Add(1)
		go authz(clients[i%len(clients)], &wg)
	}
	wg.Wait()
}

func authz(client *acme.Client, wg *sync.WaitGroup) {
	client.Authz("loadtest.le-test.hoffman-andrews.com")
	wg.Done()
}
