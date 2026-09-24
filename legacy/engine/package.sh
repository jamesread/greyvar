buildid -n
TAG=`buildid -k tag`
DIR=pkg/boleas-${TAG}/

rm -rf pkg
mkdir -p $DIR

cp bin/boleas $DIR
cp -r res $DIR

cd pkg

zip -r boleas-$TAG.zip boleas-$TAG/

