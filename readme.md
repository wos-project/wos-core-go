# wos-core-go README #

wos-core-go is a Golang CDN, cache, and geo index running over IPFS for augmented reality objects stored in the world.  wos-core-api hosts a RESTful API supporting the [wos-protocol](https://github.com/wos-project/wos-protocol).  See swagger docs for interactive help on using the protocol.  wos-core-api requires access to an IPFS node.  v1 wos-core-go supports a variety of AR object types, including ARCs [arc-spec](https://github.com/wos-project/arc-spec), pins, and pinned arcs.

wos-core-go creates objects in IPFS with S3 caching and thumbnail creation.  Use S3 for fast file access to files you've stored in IPFS.  Use RESTFful API to find objects in the world.

### Prerequisites ###
- IPFS node access
- PostgreSQL with postgis extensions
- AWS S3
- ffmpeg

### Linux PostgreSQL setup ###
```Console
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup initdb
sudo systemctl start postgresql
sudo su postgres
createdb wos_dev
psql
CREATE USER wos_user WITH PASSWORD 'insecure';
GRANT ALL PRIVILEGES ON DATABASE wos_dev TO wos_user;

# edit pg_hba.conf 
# host    all             all             127.0.0.1/32            md5
sudo systemctl restart postgresql
```

### Mac PostgreSQL install ###
```Console
brew install postgresql 
createdb wos_dev
psql -d wos_dev

CREATE USER wos_user WITH PASSWORD 'insecure';
GRANT ALL PRIVILEGES ON DATABASE wos_dev TO wos_user;
CREATE EXTENSION postgis;
```

### Mac Postgis install ###
```Console
brew install postgis
psql -d wos_dev

CREATE TABLE global_points (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64),
    location GEOGRAPHY(POINT,4326)
  );

INSERT INTO global_points (name, location) VALUES ('Town', 'SRID=4326;POINT(-110 30)');
SELECT name FROM global_points WHERE ST_DWithin(location, 'SRID=4326;POINT(-110 29)'::geography, 1000000);
```

## Build and Run ##
```Console
GO111MODULE=on

# copy config/config.yaml.default to config/config.yaml and configure secrets

# build swagger docs
swag init -g app/main.go

# build for mac
go build -o wos-core-go ./app

# build for linux
env GOOS=linux GOARCH=amd64 go build -o wos-core-go ./app

# run
./wos-core-go -config app/config.yaml
```

## Let's encrypt ##
```Console
sudo apt-get update
sudo apt-get install certbot
sudo apt-get install python3-certbot-nginx
sudo certbot --nginx -d worldos.earth

scp build/rooted/etc/cron.d/letsencrypt-renew-nginx vm:/rooted/etc/cron.d/
```

## Ffmpeg ##
The API uses ffprobe, which is part of ffmpeg to discover info about mpeg/audio files.  To install on a mac
```Console
brew install ffmpeg
```
The API will run without it, but it cannot determine the duration of audio clips without it.

## Testing ##
```Console
go test -logtostderr ./app/...
```

## Swagger ##
```Console
swag init
go build
./wos-core-go -config app/config.yaml
```
Browse http://localhost:8080/swagger/index.html and enter credentials in config.yaml swagger.users section




