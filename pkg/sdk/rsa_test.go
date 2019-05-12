package sdk

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/pjoc-team/base-service/pkg/sign"
	"github.com/pjoc-team/base-service/pkg/util"
	"reflect"
	"testing"
)

func TestPkcs8(t *testing.T) {
	key := `MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAJZF1WY+G91d5pew
9mTH5sMpQfF6k5pvMILc7pRriEVectB/QZ1EFXXrD9Nru3gcBq1DTCIEazR2m8WF
/Yubf6PtZikGSb9tM/kiPZNnp2Wgmn85rzlBkIMjG2UAmN+pTSLk5ztSjCwRPMQp
4uLVjHmEVR3mmM1H4XtL8uzJUdDhAgMBAAECgYBO068PhQEE7B7r744wa5QnR9sp
ms0Ws8DUxKP6AzZmfRbpO/flUTOYuYeBtf+PD9SIysaDCaJa0OUBhjnsI9OeEj7t
3Zwjt8nYRVdXykbKe6/jAAvgRwKHsqyzkVyQiMW+knk3HfIgvBjiJHkMJA2b5fVk
+I70nivqUwAJ+eIb4QJBAMY2CEcb2sSHa68NONy4/VS3kgbW7OtntghUNE8ITdkO
q/1MLIMbAb6Bf+SMObBSy92snFegnpG+Gs0Ph2H65fUCQQDCFc+LeEo7OhFV8QEY
wV1U2XMHohHI3HQBFZJRgJcEIUdRtO4zSsAlxYqOj0uQHbG8J4neh5xNb0i77td8
i/+9AkAeORTwCs5D00ZXLdPyy/5M0aThiBoeFvVJtdU4C9Ma+sK838WVxCNy8foX
Vk5hlW5igbRhJCupm2wowmppRUGVAkEAlRLvoS65xZgqbJp6vyr2pw+GnRxNELzT
lWmeQ1/DnvZ4szeHpnoJ8Hk0nZ9O6NkGBYFREk2TLp8FfORNO2rE+QJAFhyVRkyt
tSv1rC96v5GI0OPltMB6njy5UsafWpJXnxc7NNOsd8rxFidOZyJNAMgUcGbcXZe1
d/dD0K5NaP6UBQ==`
	publicKey := `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCWRdVmPhvdXeaXsPZkx+bDKUHx
epOabzCC3O6Ua4hFXnLQf0GdRBV16w/Ta7t4HAatQ0wiBGs0dpvFhf2Lm3+j7WYp
Bkm/bTP5Ij2TZ6dloJp/Oa85QZCDIxtlAJjfqU0i5Oc7UowsETzEKeLi1Yx5hFUd
5pjNR+F7S/LsyVHQ4QIDAQAB`
	bytes, _ := base64.StdEncoding.DecodeString(key)
	key2, err := x509.ParsePKCS8PrivateKey(bytes)
	fmt.Println(err)
	fmt.Println(key2)
	fmt.Println(reflect.TypeOf(key2))
	s := util.RandString(64)
	source := []byte(s)
	pkcs8, e := SignPKCS8(source, key, crypto.MD5)
	if e != nil{
		fmt.Println("err: ", e.Error())
		return
	}
	err = sign.VerifyPKCS1v15WithStringKey(source, pkcs8, publicKey, crypto.MD5)
	if err != nil{
		fmt.Println("Verify error! err: ", err.Error())
	} else {
		fmt.Println("Verify ok!")
	}
}

func TestPkcsv15(t *testing.T) {
	key := `MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAKwcjUGTInfreHjA
q0vdLjUthr4FzPw74XleWm7HBazX3PXP2YuUS2wBwTXq/7TXnhkhZK6nNJ+jwzXW
ddzNt3a+l89pPQ3sH2/y8ZymJbuvOCcMcS1GZ3wejpEKcWFj4qvUCJo5UbJkOAo3
zU1yC03fIXLG2v9oCWsCgk6icgs9AgMBAAECgYEAmjbTEuilP9JLFdd9JPLADoIG
c4l7DJ8S/s7eNNg7a43XvKFKidiMY/CGkKtKB14TmOzk6+GCM3Bm33yUCw6AzXom
zbNjo6e/tPN9iMa/BNlRTYZ0o9wbl3wKaXNgwAQTIcawJu4lS8Z1tkRtX39+8mQ0
7PIdV0Mxy7SX/zIVTbkCQQDgVQblblbywws38AsdOFEkiHYsB/uLJnqYmlvldIkK
VrX0b2OEkfQue3X340Ol5/K8W/LjxET4MkTH855amVuHAkEAxGhc7f/paptQ2IkR
8Gft5sa22acmg/gFQ+zOrdvCS7RftvoMkwXRf1zOTygcCKwqxmMttJs1Tap11PtU
MJB8GwJBAJT0YEvfZCR1lfFilj6kslxLsAu0kTXoKudBN1u4cXZH9TGE9NOGBlkG
6WHbqYNz7B1eAH10SUd0OW97N954jekCQBjmtmmCWaCngm+BqWGWe6YXqZUcyRhQ
OLKs8mM+inq4qq3zM+bsBOyl4o5MoolC8sVmS6QrTW3EX94Z40mo9+MCQGex95+t
GAPfOs4nJqfP07Zp7+TC2fdqMoCranzqJ1WqzubE5RmfBVccV2wMcrFGHJZhPrX0
9olHCsgRIKqQhmg=`
	//bytes, _ := base64.StdEncoding.DecodeString(key)
	key2 := ParsePrivateKey(key)
	fmt.Println(key2)
}

