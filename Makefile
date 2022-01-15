version="0.3.4"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"
remote_host = "fh@cube.local"
reprepo_host = ""

clean:
	-rm ./package/buid/conbee

build-go:
	go build -o conbee src/service.go

build-go-arm:
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o ../package/build/conbee service.go;cd ../

build-go-amd:
	cd ./src;GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o conbee service.go;cd ../


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf conbee_$(version).tar.gz conbee VERSION

package-deb-doc:
	@echo "Packaging application as debian package"
	chmod a+x package/debian/DEBIAN/*
	cp ./package/build/conbee package/debian/opt/thingsplex/conbee
	cp VERSION package/debian/opt/thingsplex/conbee
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"

tar-arm: build-js build-go-arm package-deb-doc
	@echo "The application was packaged into tar archive "

deb-arm : clean configure-arm build-go-arm package-deb-doc
	mv package/debian.deb package/build/conbee_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc
	mv debian.deb conbee_$(version)_amd64.deb

upload :
	scp package/build/conbee_$(version)_armhf.deb $(remote_host):~/

upload-install : upload
	ssh -t $(remote_host) "sudo dpkg -i conbee_$(version)_armhf.deb"

remote-install : deb-arm upload
	ssh -t $(remote_host) "sudo dpkg -i conbee_$(version)_armhf.deb"

run :
	go run src/service.go -c testdata/var/config.json

publish-reprepo:
	scp package/build/conbee_$(version)_armhf.deb $(reprepo_host):~/apps

.phony : clean
