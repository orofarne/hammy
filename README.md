BUILD
=====
    . ./env.sh
    go run bootstrap.go
    go test hammy && go install hammy hammyd hammycid hammydatad

INSTALL
=======

    ./install.sh
