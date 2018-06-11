package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"flag"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("insufficient arguments")
	}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	bucketName := fs.String("bucket", "", "bucket name")
	objectName := fs.String("object", "", "object name")
	fs.Parse(os.Args[2:])

	switch os.Args[1] {
	case "cat":
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatal(err)
		}
		bucket := client.Bucket(*bucketName)
		r, err := bucket.Object(*objectName).NewReader(ctx)
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(b))
	default:
		log.Fatal("unknown command")
	}
}
