package gs

import (
	"cloud.google.com/go/storage"
	"context"
	"io/ioutil"
	"log"
)

func Cat(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bucket := client.Bucket(bucketName)
	r, err := bucket.Object(objectName).NewReader(ctx)
	if err != nil {
		return []byte{}, err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}
