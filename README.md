# ancientcitadelgifs

Part of the ancientcitadel.com infrastructure. It takes `.gif` urls, downloads them,
and converts them to `.webm`, `.mp4` and `.jpg` (thumbnail) formats,
playable on all major browsers. The video files are then pushed to an S3 bucket.

You'll need an S3 bucket, and credentials that allow write access to said bucket.

## Deployment

```
heroku create -b https://github.com/ddollar/heroku-buildpack-multi.git
heroku config:set S3_BUCKET_HOST=
heroku config:set S3_BUCKET_NAME=
heroku config:set AWS_ACCESS_KEY_ID=
heroku config:set AWS_SECRET_ACCESS_KEY=
git push heroku master
```

## Development

```
git clone github.com/AndrewVos/ancientcitadelgifs
cd ancientcitadelgifs
mkdir vendor
curl -L --silent http://johnvansickle.com/ffmpeg/releases/ffmpeg-release-64bit-static.tar.xz | tar xvJ -C vendor
export S3_BUCKET_HOST=
export S3_BUCKET_NAME=
export AWS_ACCESS_KEY_ID=
export AWS_SECRET_ACCESS_KEY=
go build && ./ancientcitadelgifs
```

## Uploading

```
$ curl localhost:9090/upload?u=http%3A%2F%2Fmedia.giphy.com%2Fmedia%2FObXgWWGHzMlVe%2Fgiphy.gif
{
	"mp4url":  "{S3_BUCKET_HOST}/ffbbcc7fb8acaca2e3839414bc3a61bd.mp4",
	"webmurl": "{S3_BUCKET_HOST}/ffbbcc7fb8acaca2e3839414bc3a61bd.webm",
	"jpgurl":  "{S3_BUCKET_HOST}/ffbbcc7fb8acaca2e3839414bc3a61bd.jpg",
	"width":   450,
	"height":  253
}
```
