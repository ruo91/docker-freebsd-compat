# Porting to FreeBSD
I'm trying to make docker work on freebsd

Major milestones for porting docker on FreeBSD are:

* make it compile (DONE)
* make it start as a daemon (DONE)
* load an image and create the container (aka working graphdriver) (DONE)
* run the container (DONE)
* working top\start\stop\kill (aka working execdriver) (IN PROGRESS)
* working networking aka NAT
* working port forward (aka working networkdriver)
* working volumes and links
* working limits
* major code cleanup and steps to push code to docker project

(See the bigger list below)

# Running
We dont have working docker image on freebsd, and cross-compile doesn't work wery well, so now we need to compile on FreeBSD directly

Prereqesites

    pkg install go
    pkg install git
    pkg install sqlite3
    pkg install ca_root_nss # use this if pull command is not working

First we get the sources

    setenv GOPATH `pwd`
    mkdir -p src/github.com/docker    
    git clone https://github.com/kvasdopil/docker src/github.com/docker/docker
    cd src/github.com/docker/docker
    git checkout freebsd-compat
    
Now build the docker

    sh hack/make/.go-autogen
    cd $GOPATH
    cp -rp src/github.com/docker/docker/vendor/* .

    # Now sure how to do this properly for golang
    setenv CC clang # for FreeBSD 10.1
    ln -s /usr/local/include/sqlite3.h /usr/include/
    ln -s /usr/local/lib/libsqlite3.so* /usr/lib/

    go build -tags daemon github.com/docker/docker/docker

This should build the docker executable in current directory. You can run docker with command:
    
    zfs create -o mountpoint=/dk zroot/docker # mounpoint should be short
    ./docker -d -b none -e jail -s zfs -g /dk -D

After the daemon is started we can pull the image and start the container

    ./docker pull kazuyoshi/freebsd-minimal
    ./docker run kazuyoshi/freebsd-minimal echo hello world
   
Interactive mode works too

    ./docker run -it kazuyoshi/freebsd-minimal csh

# List of working commands and features

Commands:
* attach    - ok
* build
* commit    - bug
* cp        - not working on running containers, 'filename too long' bug on stopped containers
* create    - ok
* diff      - ok (on stopped containers)
* events    - ok
* exec      - ok
* export    - ok
* history   - ok
* images    - ok
* import    - ok
* info      - bug
* inspect   - ok
* kill      - ok
* load      - not working
* login     - ok
* logout    - ok
* logs      - ok
* pause     - not working (not supported on freebsd)
* port      - not working
* ps        - ok
* pull      - ok
* push      - ok
* rename    - ok
* restart   - ok
* rm        - ok
* rmi       - ok
* run       - ok
* save      - ok
* search    - ok
* start     - ok
* stats     - not working
* stop      - ok
* tag       - ok
* top       - ok
* unpause   - not working (not supported on freebsd)
* version   - ok
* wait      - ok

Features:
* image loading         - ok
* container creation    - ok
* container stop\start  - ok
* build on FreeBSD 10.1 - ok
* NAT                   - not working
* port forward          - not working
* volumes               - not working
* links                 - not working
* limits                - not working
