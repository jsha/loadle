package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"log"
	"sync"

	"github.com/jsha/lego/acme"
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
	ch := make(chan int)
	for i := 0; i < *m; i++ {
		go func() {
			// Use independent clients to avoid locking on nonce pool.
			client, err := acme.NewClient(stagingURL, myUser, acme.RSA2048)
			if err != nil {
				log.Fatal(err)
			}
			for _ = range ch {
				client.Authz("loadtest.le-test.hoffman-andrews.com")
				wg.Done()
			}
		}()
	}
	for i := 0; i < *n; i++ {
		wg.Add(1)
		ch <- 1
	}
	wg.Wait()
}
