set -e
echo compiling...
go build

export GODEBUG=on
./buildsvc -gitlab-user wenzhenglin -gitlab-pass cKGa3eVAF7tZMvCukdsP -repoDir /home/wen/t/repos \
  -harbor-user devuser -harbor-pass Ln28ohyDn