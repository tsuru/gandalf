.. image:: https://secure.travis-ci.org/globocom/gandalf.png
   :target: http://travis-ci.org/globocom/gandalf

Gandalf is an api for authenticate git users per project.

YOU SHALL NOT PASS!
==================


Installation from source
------------------------

Gandalf is built in Go, see http://golang.org/doc/install to install it. Gandalf also uses mongodb:

    $> [sudo] apt-get install mongodb

Get gandalf:

    $> go get github.com/globocom/gandalf

Gandalf will come with a default configuration file, at etc/gandalf.conf, customize it with your needs before running the install script.

The script will build and run gandalf server with the current user, so if you want your
repositories urls to be like `git@host.com` you should create a user called git and change to it before running the script.

So let's run it:

    | $> cd $GOPATH/github.com/globocom/gandalf
    | $> ./setup/install.sh

No output means no error :)

Now test if gandalf server is up and running

    $> ps -ef | grep gandalf

This should output something like the following

    git      27334     1  0 17:30 ?        00:00:00 /home/git/gandalf/dist/gandalf-webserver

Now we're ready to move on!

Api usage
---------

Create a user:

    $> curl -d '{"name": "username", "keys": [{"keyname": "ssh-rsa userpubkey user@host"}]}' gandalf-host.com/user

You should see the following:

    User "username" successfuly created

Now let's create a repository:

    $> curl -d '{"name": "myproject", "users": ["username"], "ispublic": true}' gandalf-host.com/repository

You should get the following:

    Repository "myproject" successfuly created

In order to delete a repository, execute the following:

    $> curl -XDELETE gandalf-host.com/repository/myproject

The output should be:

    Repository "myproject" successfuly removed

To delete a user:

    $> curl -XDELETE gandalf-host.com/user/username

The output should be:

    User "username" successfuly removed

TODO: grant/revoke user access to a repository

Pushing into myproject
""""""""""""""""""""""

Now we already have access to myproject, let's create a git repository locally to test our setup:

    | $> mkdir myproject
    | $> cd myproject
    | $> git init
    | $> git remote add gandalf git@gandalf-host.com:myproject.git
    | $> touch README
    | $> git add .
    | $> git commit -m "first commit"
    | $> git push gandalf master

You should see the usual git output.
