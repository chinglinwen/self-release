set -e
echo "compiling..."
go build
echo "kill exist process first"
# pkill nfssvc || :
echo "starting..."
./nfssvc -p 8006 -path /tmp/exports