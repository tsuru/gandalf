# -*- coding: utf-8 -*-

# Copyright 2012 gandalf authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

import os
from fabric.api import abort, cd, env, local, put, run, settings

current_dir = os.path.abspath(os.path.dirname(__file__))
env.user = 'git'
env.gandalf_path = '/home/%s/gandalf' % env.user


def stop():
    with settings(warn_only=True):
        run('killall -KILL gandalf-webserver')


def build():
    goos = local("go env GOOS", capture=True)
    goarch = local("go env GOARCH", capture=True)
    if goos != "linux" or goarch != "amd64":
        abort("gandalf must be built on linux_amd64 for deployment, you're on %s_%s" % (goos, goarch))
    local("mkdir -p dist")
    local("go clean ./...")
    local("go build -a -o dist/gandalf-webserver ./webserver")


def clean():
    local("rm -rf dist")
    local("rm -f dist.tar.gz")


def send():
    local("tar -czf dist.tar.gz dist")
    run("mkdir -p %(gandalf_path)s" % env)
    put(os.path.join(current_dir, "dist.tar.gz"), env.gandalf_path)


def start():
    with cd(env.gandalf_path):
        run("tar -xzf dist.tar.gz")
    run("nohup %s/dist/gandalf-webserver >& ./gandalf-webserver.out < ./webserver.out &" % env.gandalf_path, pty=False)


def deploy():
    build()
    send()
    stop()
    start()
    clean()
