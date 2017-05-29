#!/usr/bin/env bash

os=$(uname)
sudo=''
if [ "$os" = 'Darwin' ]; then
    brew install libyaml
elif [ "$os" = 'Linux' ]; then
    if [ $(lsb_release -is) = "Debian" ]; then
        apt-get install -y python-dev libyaml-dev python-pip build-essential libsqlite3-dev
    else
        sudo apt-get -y install libyaml-dev build-essential libsqlite3-dev
        sudo='sudo'
    fi
fi

# libffi-dev
#cmd="$sudo pip install -Iv cffi==1.5.2"
#$cmd

# cmd="$sudo pip install --upgrade pathtools Jinja2"
# $cmd

#cmd="$sudo pip install --no-use-wheel tuf"
#$cmd
#
#cmd="$sudo pip install tuf[tools]"
#$cmd

# https://github.com/ellisonbg/antipackage
pip install git+https://github.com/ellisonbg/antipackage.git#egg=antipackage
