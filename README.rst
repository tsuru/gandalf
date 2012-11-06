.. image:: https://secure.travis-ci.org/globocom/gandalf.png
   :target: http://travis-ci.org/globocom/gandalf

Gandalf is an api for authenticate git users per project.

YOU SHALL NOT PASS!
==================


Installation
------------

Clone gandalf

    $> git clone git://github.com/globocom/gandalf

Now run the install script (from gandalf root)

    $> cd gandalf
    $> ./setup/install.sh

No output means no error :)

Now test if gandalf server is up and running

    $> ps -ef | grep gandalf

This should output something like the following

    git      27334     1  0 17:30 ?        00:00:00 /home/git/gandalf/dist/gandalf-webserver

Now we're ready to move on!

Api usage
---------

Create a user:

    $> curl -d '{"name": "username", "keys": ["ssh-rsa userpubkey user@host"]}' gandalf-host.com/user

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
