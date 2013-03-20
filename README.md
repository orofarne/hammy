BUILD
=====

    ./build.sh

INSTALL
=======

    sudo ./install.sh

CLEAN
=====

    ./clean.sh


EXAMPLE CONFIGURATION
=====================

    ./build.sh
    ./examples/db_schema.rb | mysql -u root
    ./bin/hammyd -config=examples/config.gcfg &
    ./bin/hammydatad -config=examples/config.gcfg &
    ./bin/hammycid -config=examples/config.gcfg &

