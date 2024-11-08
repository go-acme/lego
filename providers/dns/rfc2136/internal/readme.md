# TSIG Key File

How to generate example:

```console
$ docker run --rm -it -v $(pwd):/app -w /app alpine sh
/app # apk add bind
/app # tsig-keygen lego > sample1.conf
/app # tsig-keygen -a hmac-sha512 lego > sample2.conf
```
