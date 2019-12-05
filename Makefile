version="0.2.0"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"

clean:
	-rm tpflow

build-go:
	go build -o conbee-ad src/service.go

build-go-arm:
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o conbee-ad service.go;cd ../

build-go-amd:
	cd ./src;GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o conbee-ad service.go;cd ../


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf conbee-ad_$(version).tar.gz conbee-ad VERSION

package-deb-doc-1:
	@echo "Packaging application as debian package"
	chmod a+x package/debian/DEBIAN/*
	cp ./src/conbee-ad package/debian/opt/thingsplex/conbee-ad
	cp VERSION package/debian/opt/thingsplex/conbee-ad
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"

tar-arm: build-js build-go-arm package-deb-doc-1
	@echo "The application was packaged into tar archive "

deb-arm : clean configure-arm build-go-arm package-deb-doc-1
	mv package/debian.deb package/build/conbee-ad_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc-1
	mv debian.deb conbee-ad_$(version)_amd64.deb

run :
	go run src/service.go -c testdata/var/config.json


.phony : clean
