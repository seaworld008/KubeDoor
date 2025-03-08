
     go-bindata -o=asset/asset.go   -pkg=asset version
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kubedoor-agent-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o kubedoor-agent-arm64 .

	docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	docker buildx build --push --platform=linux/amd64,linux/arm64  --tag hub.glodon.com/glodon-pub/kube-agent:v1.0.0  -f Dockerfile-local .
	docker buildx rm project-v3-builder
