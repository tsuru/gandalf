===========================
Install Gandalf from source
===========================

Clone gandalf

.. highlight:: bash

::

    $ git clone git://github.com/globocom/gandalf

Now run the install script (from gandalf root)

.. highlight:: bash

::

    $ cd gandalf
    $ ./setup/install.sh

No output means no error :)

Now test if gandalf server is up and running

.. highlight:: bash

::

    $ ps -ef | grep gandalf

This should output something like the following

.. highlight:: bash

::

    git      27334     1  0 17:30 ?        00:00:00 /home/git/gandalf/dist/gandalf-webserver

Now we're ready to move on!
