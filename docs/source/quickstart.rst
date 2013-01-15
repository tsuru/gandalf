================
Quickstart Guide
================

Api usage
=========

Create a user:

.. highlight:: bash

::

    $ curl -d '{"name": "username", "keys": [{"content": "ssh-rsa userpubkey user@host", "name": "keyname"}]}' gandalf-host.com/user

You should see the following:

.. highlight:: bash

::

    User "username" successfuly created

Now let's create a repository:

.. highlight:: bash

    $ curl -d '{"name": "myproject", "users": ["username"], "ispublic": true}' gandalf-host.com/repository

You should get the following:

.. highlight:: bash

    Repository "myproject" successfuly created

In order to delete a repository, execute the following:

.. highlight:: bash

::

    $ curl -XDELETE gandalf-host.com/repository/myproject

The output should be:

.. highlight:: bash

::

    Repository "myproject" successfuly removed

To delete a user:

.. highlight:: bash

::

    $ curl -XDELETE gandalf-host.com/user/username

The output should be:

.. highlight:: bash

::

    User "username" successfuly removed

TODO: grant/revoke user access to a repository

Pushing into myproject
======================

Now we already have access to myproject, let's create a git repository locally to test our setup:

.. highlight:: bash

::

    $ mkdir myproject
    $ cd myproject
    $ git init
    $ git remote add gandalf git@gandalf-host.com:myproject.git
    $ touch README
    $ git add .
    $ git commit -m "first commit"
    $ git push gandalf master

You ould see the usual git output.
