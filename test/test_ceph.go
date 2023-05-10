package main

import (
	"net_disk/store/ceph"
	"os"
)

func main() {
	bucket := ceph.GetCephBucket("userfile")
	d, _ := bucket.Get("/ceph/e5a44e032d10e094a8203ee5abc078504d661717")
	tmpFile, _ := os.Create("D:/tmp/test_file.png")
	tmpFile.Write(d)
	return
	//bucket := ceph.GetCephBucket("testbucket1")
	////创建一个新的bucket
	//err := bucket.PutBucket(s3.PublicRead)
	//fmt.Printf("create bucket err:%+v\n", err)
	////查询这个bucket下面指定条件的object
	//res, err := bucket.List("", "", "", 100)
	//fmt.Printf("object keys:%+v\n", res)
	////新上传一个对象
	//bucket.Put("/testupload/a.txt", []byte("just for test"), "octet-stream", s3.PublicRead)
	//fmt.Printf("upload err:%+v\n", err)
	////查询这个bucket下面指定条件的object keys
	//res, err = bucket.List("", "", "", 100)
	//fmt.Printf("object keys:%+v\n", res)

}
