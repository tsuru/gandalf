API Reference
=============

User creation
-------------

Creates an user in the database.

* Method: POST
* URI: /user
* Format: json

User removal
------------

Removes an user from the database.

Key add
-------

Adds a key to an user in the database and writes it in authorized_keys file from the user running Gandalf.

Key removal
-----------

Removes a key from a user in the database and from the authorized_keys file from the user running Gandalf.

Repository creation
-------------------

Creates a repository in the database and an equivalent bare repository in the filesystem.

Repository removal
------------------

Removes a repository from the database and the equivalent bare repository from the filesystem.

Repository retrieval
--------------------

Retrieves information about a repository.

Access grant in repository
--------------------------

Grants an user read and write access into a repository.

Access revoke in repository
---------------------------

Revokes an user read and write access from a repository.

Get file contents
-----------------

Returns the contents for a `path` in the specified `repository` with the given `ref` (commit, tag or branch).

* Method: GET
* URI: /repository/`:name`/contents/`:path`?ref=:ref
* Format: binary

Where:

* `:name` is the name of the repository;
* `:path` is the file path in the repository file system;
* `:ref` is the repository ref (commit, tag or branch). **This is optional**. If not passed this is assumed to be "master".

Get archive
-----------

Returns the compressed archive for the specified `repository` with the given `ref` (commit, tag or branch).

* Method: GET
* URI: /repository/`:name`/archive/`:ref.:format`
* Format: binary

Where:

* `:name` is the name of the repository;
* `:ref` is the repository ref (commit, tag or branch);
* `:format` is the format to return the archive. This can be zip, tar or tar.gz.
