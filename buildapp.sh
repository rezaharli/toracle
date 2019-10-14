cd $GOPATH/src/git.eaciitapp.com/rezaharli/toracle

echo "Removing old files if exists"
rm -rf $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/toracle
rm -f $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/toracle.zip

echo "Creating build"
cd $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/
go build main.go

echo "Duplicating project"
cd $GOPATH/src/git.eaciitapp.com/rezaharli/toracle
mkdir toracle
mv $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/main.exe $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/toracle/
cp -R $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/config/ $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/toracle/config

echo "Zipping the build"
tanggal=`date +%Y-%m-%d`
zip -r toracle_$tanggal.zip toracle
mv toracle_$tanggal.zip $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/toracle
# rm -rf $GOPATH/src/git.eaciitapp.com/rezaharli/toracle

cd $GOPATH/src/git.eaciitapp.com/rezaharli/toracle

# echo "Uploading"
# rsync -avz --inplace --progress -e "ssh -i $KEYS/developer.pem" toracle_$tanggal.zip developer@go.eaciit.com:/data/nginx/files/toracle/

# echo "Removing current build files"
# rm -f $GOPATH/src/git.eaciitapp.com/rezaharli/toracle/toracle_$tanggal.zip
# rm -f $GOPATH/src/git.eaciitapp.com/rezaharli/toracle.zip