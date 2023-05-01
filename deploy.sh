#/bin/sh

addr=$1
if [ -z "$1" ]; then
    echo no address set
    exit
fi

echo deploying to $addr ...
echo copying files...
rsync -av --delete-after . root@$addr:/fakes
echo starting docker-compose...
ssh root@$addr << EOF
    cd /fakes
    docker-compose stop
    docker-compose up --build --force-recreate -d
    docker image prune -f
EOF