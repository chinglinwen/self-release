set -e
echo "compiling..."
go build
echo "kill exist process first"
pkill self-release || :
echo "starting..."
# ./self-release -w http://wechat-notify.devops.haodai.net -agentid 1000003 -secret G5h7CTEqkBw-Fe3luf2JM8UNNJAcYTpbXvpveY7M3lg &> s.log &
./self-release -gitlab-user wenzhenglin -gitlab-pass cKGa3eVAF7tZMvCukdsP &> s.log &
tail -f s.log
