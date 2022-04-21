package main

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"os"
	"time"
)

const LEADER = "leader"

func main() {
	client, errCli := clientv3.New(clientv3.Config{
		Endpoints:   []string{"Etcd:2379", "Etcd:2380"},
		DialTimeout: 5 * time.Second,
	})
	if errCli != nil {
		log.Fatal(errCli)
	}
	defer client.Close()

	node := os.Getenv("NODE_NAME")

	if node == "" {
		log.Fatal("NODE_NAME is not set")
		return
	}

	leaseResponse, errGrant := client.Grant(context.TODO(), 5)
	if errGrant != nil {
		log.Fatal(errGrant)
	}

	for true {
		isLeader := leaderElection(client, node, leaseResponse)

		if isLeader {
			log.Printf("\nLeader [%s]", node)
		} else {
			log.Printf("\nFollower [%s]", node)
		}
		time.Sleep(time.Second * 5)
	}

}

func leaderElection(client *clientv3.Client, node string, leaseResponse *clientv3.LeaseGrantResponse) bool {
	lease := clientv3.WithLease(leaseResponse.ID)

	leader, errLeader := client.Get(context.TODO(), LEADER)
	if errLeader != nil {
		client.Revoke(context.TODO(), leaseResponse.ID)
		return false
	}

	if leader.Count <= 0 {
		_, errPut := client.Put(context.TODO(), LEADER, node, lease)
		if errPut != nil {
			client.Revoke(context.TODO(), leaseResponse.ID)
			return false
		}
		return true
	}

	for _, keyValue := range leader.Kvs {
		log.Printf("\n-> data %s", keyValue)

	}

	return false

}
