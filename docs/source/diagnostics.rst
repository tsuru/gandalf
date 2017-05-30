===========
Diagnostics
===========

Gandalf uses `gops <github.com/google/gops>`_ for instrumenting its go processes, see their documentation for more.


To start the webserver with the diagnostics agent, use the ``-diagnostic`` flag.

.. note::

        Gops will start a server you can connect to, exposing it on a randomly availabe port.
        Make sure you're on a safe network before running it.
