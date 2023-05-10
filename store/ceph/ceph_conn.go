package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var cephConn *s3.S3

func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	//1.初始化ceph的一些信息
	auth := aws.Auth{
		AccessKey: "XBANCI07289XSUCPKQU2",
		SecretKey: "yKiga0urJTGlhTw8fdyfDisfSTRLFLBvLqvQ8nPD",
	}

	curRegion := aws.Region{
		Name:                 "default",
		EC2Endpoint:          "http://192.168.59.78:9080",
		S3Endpoint:           "http://192.168.59.78:9080",
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}
	//2.创建S3类型的连接
	return s3.New(auth, curRegion)
}

// GetCephBucket GetCephBucket:获取指定的bucket对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}
