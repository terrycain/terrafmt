PLATFORMS := darwin/386 linux/386 linux/amd64 linux/arm64 linux/arm windows/amd64 darwin/amd64
SIGNED_PLATFORMS := darwin/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
now = $(shell date +'%Y-%m-%dT%T')
version = $(word 3, $(subst /, ,${GITHUB_REF}))

release: $(PLATFORMS)
mac_release: $(SIGNED_PLATFORMS)

build:
	go build -ldflags="-s -w -X main.sha1=${GITHUB_SHA} -X main.buildTime=${now} -X main.version=${version}"

build_amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.sha1=${GITHUB_SHA} -X main.buildTime=${now} -X main.version=${version}"

clean:
	rm -f terrafmt terrafmt-*.tar.gz

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -ldflags="-s -w -X main.sha1=${GITHUB_SHA} -X main.buildTime=${now} -X main.version=${version}" -o 'terrafmt-$(os)-$(arch)'
	chmod +x 'terrafmt-$(os)-$(arch)'
	# Dodgyness so we can cross compile in parallel but end up with a tar.gz with terrafmt inside
	tar --transform='flags=r;s|terrafmt-$(os)-$(arch)|terrafmt|' -czvf 'terrafmt-$(os)-$(arch).tar.gz' 'terrafmt-$(os)-$(arch)'
	rm 'terrafmt-$(os)-$(arch)'

$(SIGNED_PLATFORMS):
	mkdir -p 'terrafmt-$(os)-$(arch)'
	GOOS=$(os) GOARCH=$(arch) go build -ldflags="-s -w -X main.sha1=${GITHUB_SHA} -X main.buildTime=${now} -X main.version=${version}" -o 'terrafmt-$(os)-$(arch)/terrafmt'
	chmod +x 'terrafmt-$(os)-$(arch)/terrafmt'
	cd 'terrafmt-$(os)-$(arch)' && ../gon ../gon_config.hcl
	mv 'terrafmt-$(os)-$(arch)/terrafmt.zip' 'terrafmt-$(os)-$(arch).zip'
	# rm -rf 'terrafmt-$(os)-$(arch)'

