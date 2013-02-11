================
Quickstart Guide
================

Before starting, make sure Gandalf is :doc:`installed </install>`.

Creating an user and a repository
=================================

Create a user:

.. highlight:: bash

::

    $ curl -d '{"name": "username", "keys": {"keyname": "ssh-rsa userpubkey user@host"}}' gandalf-host.com/user

You should see the following:

.. highlight:: bash

::

    User "username" successfully created

Now let's create a repository:

.. highlight:: bash

::

    $ curl -d '{"name": "myproject", "users": ["username"], "ispublic": true}' gandalf-host.com/repository

You should get the following:

.. highlight:: bash

::

    Repository "myproject" successfully created

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

Removing an user and a repository
=================================

In order to delete a repository, execute the following:

.. highlight:: bash

::

    $ curl -XDELETE gandalf-host.com/repository/myproject

The output should be:

.. highlight:: bash

::

    Repository "myproject" successfully removed

To delete a user:

.. highlight:: bash

::

    $ curl -XDELETE gandalf-host.com/user/username

The output should be:

.. highlight:: bash

::

    User "username" successfully removed
