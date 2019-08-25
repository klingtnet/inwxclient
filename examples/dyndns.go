package main

import (
	"log"
	"net/http"
	"os"

	"github.com/klingtnet/inwxclient"
	"github.com/mitchellh/mapstructure"
)

type NameserverInfo struct {
	Cleanup struct {
		Status string
		TStamp int
	}
	Domain string
	Count  int
	Record []struct {
		ID   int
		Name string
	}
}

func main() {
	cl, err := inwxclient.NewDOMRobot(inwxclient.ProdAPI, http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cl.Do("account.login", map[string]string{"user": os.Getenv("INWX_USER"), "pass": os.Getenv("INWX_PASS")})
	if err != nil {
		log.Fatal("login", err)
	}

	resp, err = cl.Do("nameserver.info", map[string]string{"domain": "example.domain", "name": "subdomain", "type": "A"})
	if err != nil {
		log.Fatal("info", err)
	}
	var data NameserverInfo
	mapstructure.Decode(resp.Data, &data)
	if data.Count != 1 {
		log.Fatalf("%#v", resp)
	}

	resp, err = cl.Do("nameserver.updateRecord", map[string]interface{}{"id": data.Record[0].ID, "content": "127.0.0.1"})
	if err != nil {
		log.Fatal("update", err)
	}

	resp, err = cl.Do("account.logout", nil)
	if err != nil {
		log.Fatal("logout", err)
	}
}
