==================
Backing up Gandalf
==================

You can use the misc/backup.bash script to make a backup of authorized_keys and
repositories files created by gandalf, and misc/mongodb/backup.bash to make a
backup of the database. Both scripts store archives in S3 buckets.

Dependencies
============

The backups script sends these data to the s3 using the `s3cmd
<http://s3tools.org/s3cmd>`_ tool.

First, make sure you have installed s3cmd. You can install it using your
preferred package manager. For more details, refer to its `download
documentation <http://s3tools.org/download>`_.

Now let's configure s3cmd, it requires your amazon access and secret key:

.. highlight:: bash

::

    $ s3cmd --configure

authorized_keys and bare repositories
=====================================

In order to make backups, use the ``backup.bash`` script. It's able to backup
the authorized_keys file and all repositories. For backing up only the
authorized_keys file, execute it with only one parameter:

.. highlight:: bash

::

    $ ./misc/backup.bash s3://mybucket

This parameter is the bucket to which you want to send the file.

To include all bare repositories, use a second parameter, indicating the path
to the repositories:

.. highlight:: bash

::

    $ ./misc/backup.bash s3://mybucket /var/repositories
