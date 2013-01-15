===========================
Making a backup for Gandalf
===========================

You can use the misc/backup.bash script to make a backup
for authorized_keys and repositories files created by gandalf.

This backup script sends these data to the s3.

First, let's install the s3cmd.

If you use ubuntu you can use apt-get to install s3cmd.

.. highlight:: bash

::

    $ sudo apt-get install s3cmd

If you use mac you can use homebrew to install s3cmd:

.. highlight:: bash

::

    $ brew install s3cmd


Now let's configure s3cmd, its need to set your amazon access and secret key:

.. highlight:: bash

::
    $ s3cmd --configure


For use backup.bash script to make a backup for authorized_keys:

.. highlight:: bash

::
    $ ./backup.bash s3://mybucket

And, if you wanto make a backup for authorized_keys and repositorie files:

.. highlight:: bash

::
    $ ./backup.bash s3://mybucket /var/repositories
