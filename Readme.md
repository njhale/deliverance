# deliverance

WARNING: proof of concept, not for general consumption

## Run it locally

```sh
# install
$ go get github.com/ecordell/bndlr

# start a registry
$ docker run -it --rm -p 5000:5000 registry

# push manifests
$ /bndlr push ./manifests localhost:5000/ecordell/testbndlr:test
  Pushing to localhost:5000/ecordell/testbndlr:test...
  Uploading d213f9ccc4e4 manifests
  DEBU[0000] push                                          digest="sha256:d213f9ccc4e47682afb20622d9ad6dc91a6207252cd262c4983947c48beaddef" mediatype=application/vnd.docker.image.rootfs.diff.tar.gzip size=8308
  DEBU[0000] push                                          digest="sha256:92d28d80396d5490334847b54c87a38d1bb0f606c8989787c47c208ee45d9ab8" mediatype=application/vnd.docker.container.image.v1+json size=153
  DEBU[0000] do request                                    digest="sha256:d213f9ccc4e47682afb20622d9ad6dc91a6207252cd262c4983947c48beaddef" mediatype=application/vnd.docker.image.rootfs.diff.tar.gzip request.headers="map[Accept:[application/vnd.docker.image.rootfs.diff.tar.gzip, *]]" request.method=HEAD size=8308 url="http://localhost:5000/v2/ecordell/testbndlr/blobs/sha256:d213f9ccc4e47682afb20622d9ad6dc91a6207252cd262c4983947c48beaddef"
  DEBU[0000] do request                                    digest="sha256:92d28d80396d5490334847b54c87a38d1bb0f606c8989787c47c208ee45d9ab8" mediatype=application/vnd.docker.container.image.v1+json request.headers="map[Accept:[application/vnd.docker.container.image.v1+json, *]]" request.method=HEAD size=153 url="http://localhost:5000/v2/ecordell/testbndlr/blobs/sha256:92d28d80396d5490334847b54c87a38d1bb0f606c8989787c47c208ee45d9ab8"
  DEBU[0000] fetch response received                       digest="sha256:92d28d80396d5490334847b54c87a38d1bb0f606c8989787c47c208ee45d9ab8" mediatype=application/vnd.docker.container.image.v1+json response.headers="map[Accept-Ranges:[bytes] Cache-Control:[max-age=31536000] Content-Length:[153] Content-Type:[application/octet-stream] Date:[Fri, 11 Oct 2019 20:52:21 GMT] Docker-Content-Digest:[sha256:92d28d80396d5490334847b54c87a38d1bb0f606c8989787c47c208ee45d9ab8] Docker-Distribution-Api-Version:[registry/2.0] Etag:[\"sha256:92d28d80396d5490334847b54c87a38d1bb0f606c8989787c47c208ee45d9ab8\"] X-Content-Type-Options:[nosniff]]" size=153 status="200 OK" url="http://localhost:5000/v2/ecordell/testbndlr/blobs/sha256:92d28d80396d5490334847b54c87a38d1bb0f606c8989787c47c208ee45d9ab8"
  DEBU[0000] fetch response received                       digest="sha256:d213f9ccc4e47682afb20622d9ad6dc91a6207252cd262c4983947c48beaddef" mediatype=application/vnd.docker.image.rootfs.diff.tar.gzip response.headers="map[Accept-Ranges:[bytes] Cache-Control:[max-age=31536000] Content-Length:[8308] Content-Type:[application/octet-stream] Date:[Fri, 11 Oct 2019 20:52:21 GMT] Docker-Content-Digest:[sha256:d213f9ccc4e47682afb20622d9ad6dc91a6207252cd262c4983947c48beaddef] Docker-Distribution-Api-Version:[registry/2.0] Etag:[\"sha256:d213f9ccc4e47682afb20622d9ad6dc91a6207252cd262c4983947c48beaddef\"] X-Content-Type-Options:[nosniff]]" size=8308 status="200 OK" url="http://localhost:5000/v2/ecordell/testbndlr/blobs/sha256:d213f9ccc4e47682afb20622d9ad6dc91a6207252cd262c4983947c48beaddef"
  DEBU[0000] push                                          digest="sha256:77f6d184053070e89a37284dcf4a312c7f3794c3aec6cd31d7999395e5915a2d" mediatype=application/vnd.docker.distribution.manifest.v2+json size=611
  DEBU[0000] do request                                    digest="sha256:77f6d184053070e89a37284dcf4a312c7f3794c3aec6cd31d7999395e5915a2d" mediatype=application/vnd.docker.distribution.manifest.v2+json request.headers="map[Accept:[application/vnd.docker.distribution.manifest.v2+json, *]]" request.method=HEAD size=611 url="http://localhost:5000/v2/ecordell/testbndlr/manifests/test"
  DEBU[0000] fetch response received                       digest="sha256:77f6d184053070e89a37284dcf4a312c7f3794c3aec6cd31d7999395e5915a2d" mediatype=application/vnd.docker.distribution.manifest.v2+json response.headers="map[Content-Length:[611] Content-Type:[application/vnd.docker.distribution.manifest.v2+json] Date:[Fri, 11 Oct 2019 20:52:21 GMT] Docker-Content-Digest:[sha256:77f6d184053070e89a37284dcf4a312c7f3794c3aec6cd31d7999395e5915a2d] Docker-Distribution-Api-Version:[registry/2.0] Etag:[\"sha256:77f6d184053070e89a37284dcf4a312c7f3794c3aec6cd31d7999395e5915a2d\"] X-Content-Type-Options:[nosniff]]" size=611 status="200 OK" url="http://localhost:5000/v2/ecordell/testbndlr/manifests/test"
  Pushed  with digest sha256:77f6d184053070e89a37284dcf4a312c7f3794c3aec6cd31d7999395e5915a2d
```
