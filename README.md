
## build
```
go build
```

## config
config file 
```
[default]
access_key = xxxx
host_base = s3.xxx.corp
host_bucket = s3.xxx.corp/%(bucket)
secret_key = xxxxx
use_https = False
region= default
```

### usage

```
#s3-cli ls [s3://BUCKET[/PREFIX]]
#s3-cli put FILE [FILE...] s3://BUCKET[/PREFIX]
#s3-cli get s3://BUCKET/OBJECT LOCAL_FILE
#s3-cli rm s3://BUCKET/OBJECT
#s3-cli sync LOCAL_DIR s3://BUCKET[/PREFIX] or s3://BUCKET[/PREFIX] LOCAL_DIR
#s3-cli du s3://BUCKET[/PREFIX]
#s3-cli info s3://BUCKET[/OBJECT]
```