DIRECTORY ?= $(HOME)/Downloads/

test: 
	go test -v -cover ./...

setup/osquery/mac:
	@./setup-osquery.sh $(DIRECTORY)

logsdb:
	docker stop logsdb; docker rm logsdb; true;
	docker run -d --name logsdb -p 27017:27017 \
	-e MONGO_INITDB_ROOT_USERNAME=user \
	-e MONGO_INITDB_ROOT_PASSWORD=password \
	-v mongo_data:/data/db \
	mongo

start/osqueryd/mac:
	sudo -v
	echo "> staring osquery!"
	sudo /opt/osquery/lib/osquery.app/Contents/MacOS/osqueryd --verbose --disable_events=false --disable_audit=false --disable_endpointsecurity=false --disable_endpointsecurity_fim=false --enable_file_events=true > /dev/null 2>&1 &

stop/osqueryd/mac:
	sudo pkill osqueryd

run/dev/mac:
	sudo $(HOME)/go/bin/wails dev

build/package:
	sudo $(HOME)/go/bin/wails build

run/build/mac:
	sudo ./build/bin/filechangestracker.app/Contents/MacOS/filechangestracker > /dev/null 2>&1 &