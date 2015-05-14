==================
Installation guide
==================

This document describes how to install Gandalf using the `tsuru PPA
<https://launchpad.net/~tsuru/+archive/ppa>`_.

This document assumes that Gandalf is being installed on Ubuntu. You can use
equivalent packages for git, MongoDB and other gandalf dependencies, and :doc:`build
Gandalf from source </install-from-source>`, if you're planning to run Gandalf on other platforms or
distributions. Please make sure you satisfy minimal version requirements.

Adding the PPA
==============

You can add the PPA and install Gandalf with the following commands:

.. highlight:: bash

::

    $ sudo apt-add-repository ppa:tsuru/ppa
    $ sudo apt-get update
    $ sudo apt-get install gandalf-server

It will install Git automatically, but you'll still need to install MongoDB
manually. After installing MongoDB, or editing ``/etc/gandalf.conf`` to point
to an external MongoDB database (see `Configuring the server`_ for more
details), you can start Gandalf and Git daemon using upstart management
commands:

.. highlight:: bash

::

    $ sudo start git-daemon
    $ sudo start gandalf-server

Configuring the server
----------------------

Before running gandalf, you must configure it. By default, gandalf will look for
the configuration file in the ``/etc/gandalf.conf`` path. You can check a
sample configuration file and documentation for each gandalf setting in the
:doc:`"Configuring gandalf" </config>` page.

You can download the sample configuration file from Github:

.. highlight:: bash

::

    $ [sudo] curl -sL https://raw.github.com/tsuru/gandalf/master/etc/gandalf.conf -o /etc/gandalf.conf

Starting
--------

As stated above, you can start the Gandalf server and Git daemon using upstart
commands:

.. highlight:: bash

::

    $ sudo restart git-daemon
    $ sudo restart gandalf-server
