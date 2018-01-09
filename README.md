# In-flight deb packages of Go

Notice
------

This is a fixed up version of godeb from [niemeyer/godeb](https://github.com/niemeyer/godeb) containing all necessary PRs from the original repository to produce a working version of godeb that produces non-broken debian packages also when compiled with modern go versions.

Introduction
------------

For details of how this tool works and context for why it was built,
please refer to the following blog post:

  * [In-flight deb packages of Go](http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go)


Installation and usage
----------------------

If you want to compile it yourself you currently need to check out the repository in your GOPATH under github.com/niemeyer/godeb and then execute

    CGO_ENABLED=0 go build -tags netgo

Otherwise, there are pre-built binaries available for all architectures currently supported by go, the generic URL is `https://storage.googleapis.com/godeb/godeb-$GOARCH$GOARM.tar.gz`, there is also a signature for every released file ending in `.asc`.

  * [amd64](https://storage.googleapis.com/godeb/godeb-amd64.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-amd64.tar.gz.asc))
  * [386](https://storage.googleapis.com/godeb/godeb-386.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-386.tar.gz.asc))
  * [arm GOARCH=5](https://storage.googleapis.com/godeb/godeb-arm5.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-arm5.tar.gz.asc))
  * [arm GOARCH=6](https://storage.googleapis.com/godeb/godeb-arm6.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-arm6.tar.gz.asc))
  * [arm GOARCH=7](https://storage.googleapis.com/godeb/godeb-arm7.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-arm7.tar.gz.asc))
  * [arm64](https://storage.googleapis.com/godeb/godeb-arm64.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-arm64.tar.gz.asc))
  * [ppc64le](https://storage.googleapis.com/godeb/godeb-ppc64le.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-ppc64le.tar.gz.asc))
  * [s390x](https://storage.googleapis.com/godeb/godeb-s390x.tar.gz) ([sign](https://storage.googleapis.com/godeb/godeb-s390x.tar.gz.asc))

Create a release
----------------

Notes how we create a release

```
./build.sh
for i in *.tar.gz; do gpg2 --sign --detach-sign -u mgebetsroither@mgit.at --armor $i; done
for i in *.asc; do gpg $i; done
gsutil -m cp godeb-* gs://godeb
```

License
-------

The godeb code is licensed under the LGPL with an exception that allows it to be linked statically. Please see the LICENSE file for details.
