API Reference
=============

User creation
-------------

Creates a user in the database.

* Method: POST
* URI: /user
* Format: json

User removal
------------

Removes a user from the database.

Key add
-------

Adds a key to a user in the database and writes it in authorized_keys file from the user running Gandalf.

Key removal
-----------

Removes a key from a user in the database and from the authorized_keys file from the user running Gandalf.

Repository creation
-------------------

Creates a repository in the database and an equivalent bare repository in the filesystem.

* Method: POST
* URI: /repository
* Format: JSON

Example URL (http://gandalf-server omitted for clarity)::

    $ curl -XPOST /repository \                  # POST to /repository
        -d '{"name": "myrepository", \           # Name of the repository
            "users": ["myuser"], \               # Users with read/write access
            "readonlyusers": ["alice", "bob"]}'  # Users with read-only access

Repository removal
------------------

Removes a repository from the database and the equivalent bare repository from the filesystem.

Repository retrieval
--------------------

Retrieves information about a repository.

Access set in repository
--------------------------

Redefines collections of users with read and write access into a repository. Specify ``readonly=yes`` if you'd like to set read-only access.

* Method: PUT
* URI: /repository/set
* Format: JSON

Example URL for **read/write** access (http://gandalf-server omitted for clarity)::

    $ curl -XPUT /repository/set \                  # PUT to /repository/set
        -d '{"repositories": ["myrepo"], \          # Collection of repositories
            "users": ["john", "james"]}'            # Users with read/write access

Example URL for **read-only** access (http://gandalf-server omitted for clarity)::

    $ curl -XPUT /repository/set?readonly=yes \     # PUT to /repository/set
        -d '{"repositories": ["myrepo"], \          # Collection of repositories
            "users": ["bob", "alice"]}'             # Users with read-only access

Access grant in repository
--------------------------

Grants a user read and write access into a repository. Specify ``readonly=yes`` if you'd like to grant read-only access.

* Method: POST
* URI: /repository/grant
* Format: JSON

Example URL for **read/write** access (http://gandalf-server omitted for clarity)::

    $ curl -XPOST /repository/grant \               # POST to /repository/grant
        -d '{"repositories": ["myrepo"], \          # Collection of repositories
            "users": ["john", "james"]}'            # Users with read/write access

Example URL for **read-only** access (http://gandalf-server omitted for clarity)::

    $ curl -XPOST /repository/grant?readonly=yes \  # POST to /repository/grant
        -d '{"repositories": ["myrepo"], \          # Collection of repositories
            "users": ["bob", "alice"]}'             # Users with read-only access

Access revoke in repository
---------------------------

Revokes a user both read **and** write access from a repository.

* Method: DELETE
* URI: /repository/revoke
* Format: JSON

Example URL (http://gandalf-server omitted for clarity)::

    $ curl -XDELETE /repository/revoke \            # DELETE to /repository/grant
        -d '{"repositories": ["myrepo"], \          # Collection of repositories
            "users": ["john", "james"]}'            # Users with read-only access

Get file contents
-----------------

Returns the contents for a `path` in the specified `repository` with the given `ref` (commit, tag or branch).

* Method: GET
* URI: /repository/`:name`/contents?ref=:ref&path=:path
* Format: binary

Where:

* `:name` is the name of the repository;
* `:path` is the file path in the repository file system;
* `:ref` is the repository ref (commit, tag or branch). **This is optional**. If not passed this is assumed to be "master".

Example URLs (http://gandalf-server omitted for clarity)::

    $ curl /repository/myrepository/contents?ref=0.1.0&path=/some/path/in/the/repo.txt
    $ curl /repository/myrepository/contents?path=/some/path/in/the/repo.txt  # gets master

Get tree
--------

Returns a list of all the files under a `path` in the specified `repository` with the given `ref` (commit, tag or branch).

* Method: GET
* URI: /repository/`:name`/tree?ref=:ref&path=:path
* Format: JSON

Where:

* `:name` is the name of the repository;
* `:path` is the file path in the repository file system. **This is optional**. If not passed this is assumed to be ".";
* `:ref` is the repository ref (commit, tag or branch). **This is optional**. If not passed this is assumed to be "master".

Example result::

    [{
        filetype: "blob",
        hash: "6767b5de5943632e47cb6f8bf5b2147bc0be5cf8",
        path: ".gitignore",
        permission: "100644",
        rawPath: ".gitignore"
    }, {
        filetype: "blob",
        hash: "fbd8b6db62282a8402a4fc5503e9a886b4fb8b4b",
        path: ".travis.yml",
        permission: "100644",
        rawPath: ".travis.yml"
    }]

`rawPath` contains exactly the value returned from git (with escaped characters, quotes, etc), while `path` is somewhat cleaner (spaces removed, quotes removed from the left and right).

Example URLs (http://gandalf-server omitted for clarity)::

    $ curl /repository/myrepository/tree                                 # gets master and root path(.)
    $ curl /repository/myrepository/tree?ref=0.1.0                       # gets 0.1.0 tag and root path(.)
    $ curl /repository/myrepository/tree?ref=0.1.0&path=/myrepository    # gets 0.1.0 tag and files under /myrepository

Get archive
-----------

Returns the compressed archive for the specified `repository` with the given `ref` (commit, tag or branch).

* Method: GET
* URI: /repository/`:name`/archive?ref=:ref&format=:format
* Format: binary

Where:

* `:name` is the name of the repository;
* `:ref` is the repository ref (commit, tag or branch);
* `:format` is the format to return the archive. This can be zip, tar or tar.gz.

Example URLs (http://gandalf-server omitted for clarity)::

    $ curl /repository/myrepository/archive?ref=master&format=zip        # gets master and zip format
    $ curl /repository/myrepository/archive?ref=master&format=tar.gz     # gets master and tar.gz format
    $ curl /repository/myrepository/archive?ref=0.1.0&format=zip         # gets 0.1.0 tag and zip format

Get branches
------------

Returns a list of all the branches of the specified `repository`.

* Method: GET
* URI: /repository/`:name`/branches
* Format: JSON

Where:

* `:name` is the name of the repository.

Example result::

    [{
        ref: "6767b5de5943632e47cb6f8bf5b2147bc0be5cf8",
        name: "master",
        subject: "much WOW",
        createdAt: "Mon Jul 28 10:13:27 2014 -0300"
        author: {
            name: "Author name",
            email: "<author@email.com>",
            date: "Mon Jul 28 10:13:27 2014 -0300"
        },
        committer: {
            name: "Committer name",
            email: "<committer@email.com>",
            date: "Tue Jul 29 13:43:57 2014 -0300"
        },
        tagger: {
            date: "",
            email: "",
            name: ""
        },
        _links: {
            zipArchive: "/repository/myrepository/branch/archive?ref=master&format=zip",
            tarArchive: "/repository/myrepository/branch/archive?ref=master&format=tar.gz"
        }
    }]

Example URL (http://gandalf-server omitted for clarity)::

    $ curl /repository/myrepository/branches                  # gets list of branches

Get tags
--------

Returns a list of all the tags of the specified `repository`.

* Method: GET
* URI: /repository/`:name`/tags
* Format: JSON

Where:

* `:name` is the name of the repository.

Example result::

    [{
        ref: "6767b5de5943632e47cb6f8bf5b2147bc0be5cf8",
        name: "0.1",
        subject: "much WOW",
        createdAt: "Mon Jul 28 10:13:27 2014 -0300"
        author: {
            name: "Author name",
            email: "<author@email.com>",
            date: "Mon Jul 28 10:13:27 2014 -0300"
        },
        committer: {
            name: "Committer name",
            email: "<committer@email.com>",
            date: "Tue Jul 29 13:43:57 2014 -0300"
        },
        tagger: {
            name: "",
            email: "",
            date: ""
        },
        _links: {
            zipArchive: "/repository/myrepository/branch/archive?ref=0.1&format=zip",
            tarArchive: "/repository/myrepository/branch/archive?ref=0.1&format=tar.gz"
        }
    }]

Example result for an `annotated tag <https://git-scm.com/book/en/v2/Git-Basics-Tagging#Annotated-Tags>`_::

    [{
        ref: "6767b5de5943632e47cb6f8bf5b2147bc0be5cf8",
        name: "0.2",
        subject: "much WOW",
        createdAt: "Tue Jul 29 13:43:57 2014 -0300"
        author: {
            name: "",
            email: "",
            date: ""
        },
        committer: {
            name: "",
            email: "",
            date: ""
        },
        tagger: {
            name: "Tagger name",
            email: "<tagger@email.com>",
            date: "Tue Jul 29 13:43:57 2014 -0300"
        },
        _links: {
            zipArchive: "/repository/myrepository/branch/archive?ref=0.2&format=zip",
            tarArchive: "/repository/myrepository/branch/archive?ref=0.2&format=tar.gz"
        }
    }]

Example URL (http://gandalf-server omitted for clarity)::

    $ curl /repository/myrepository/tags                      # gets list of tags

Add repository hook
-------------------

Create a repository hook.

* Method: POST
* URI: /hook/`:name`

Where:

* `:name` is the name of the hook.

    - Supported hook names:

        * `post-receive`
        * `pre-receive`
        * `update`

Example URL for bare repository (http://gandalf-server omitted for clarity)::

    $ curl -d '{"content": "content of my post-receive hook"}' localhost:8000/hook/post-receive

You should see the following:

.. highlight:: bash

::

    hook post-receive successfully created


Example URL for one or more repositories (http://gandalf-server omitted for clarity)::

    $ curl -d '{"repositories": ["some-repo"], "content": "content of my update hook"}' localhost:8000/hook/update

You should see the following:

.. highlight:: bash

::

    hook update successfully created for some-repo

Commit
------

Commits a ZIP file into `repository`.

* Method: POST
* URI: /repository/`:name`/commit
* Format: MULTIPART

Where:

* `:name` is the name of the repository.

Expects a multipart form with the following fields:

* `message`: The commit message
* `author-name`: The name of the author
* `author-email`: The email of the author
* `committer-name`: The name of the committer
* `committer-email`: The email of the committer
* `branch`: The name of the branch this commit will be applied to
* `zipfile`: A ZIP file with files and directory structure for this commit. These
  files will copied on top of current repository contents.

Due to files being added over current existing repository contents, it's not
possible to remove exiting files from the repository. It's only possible to add or
modify existing ones.

Example URL (http://gandalf-server omitted for clarity)::

    # commit `scaffold.zip` into `myrepository`:
    $ curl -XPOST /repository/myrepository/commit \
        -F "message=Repository scaffold" \
        -F "author-name=Author Name" \
        -F "author-email=author@email.com" \
        -F "committer-name=Committer Name" \
        -F "committer-email=committer@email.com" \
        -F "branch=master" \
        -F "zipfile=@scaffold.zip"

Example result::

    {
        ref: "6767b5de5943632e47cb6f8bf5b2147bc0be5cf8",
        name: "master",
        subject: "Repository scaffold",
        createdAt: "Mon Jul 28 10:13:27 2014 -0300"
        author: {
            name: "Author Name",
            email: "<author@email.com>",
            date: "Mon Jul 28 10:13:27 2014 -0300"
        },
        committer: {
            name: "Committer Name",
            email: "<committer@email.com>",
            date: "Tue Jul 29 13:43:57 2014 -0300"
        },
        tagger: {
            date: "",
            email: "",
            name: ""
        },
        _links: {
            tarArchive: "/repository/myrepository/archive?ref=master&format=tar.gz",
            zipArchive: "/repository/myrepository/archive?ref=master&format=zip",
        }
    }

Logs
----

Returns a list of all commits into `repository`.

* Method: GET
* URI: /repository/`:name`/logs?ref=:ref&total=:total
* Format: JSON

Where:

* `:name` is the name of the repository;
* `:ref` is the repository ref (commit, tag or branch);
* `:total` is the maximum number of items to retrieve

Example URL (http://gandalf-server omitted for clarity)::

    $ curl /repository/myrepository/logs?ref=HEAD&total=1

Example result::

    {
        commits: [{
            ref: "6767b5de5943632e47cb6f8bf5b2147bc0be5cf8",
            subject: "much WOW",
            createdAt: "Mon Jul 28 10:13:27 2014 -0300"
            author: {
                name: "Author name",
                email: "<author@email.com>",
                date: "Mon Jul 28 10:13:27 2014 -0300"
            },
            committer: {
                name: "Committer name",
                email: "<committer@email.com>",
                date: "Tue Jul 29 13:43:57 2014 -0300"
            },
            parent: [
                "a367b5de5943632e47cb6f8bf5b2147bc0be5cf8"
            ]
        }],
        next: "1267b5de5943632e47cb6f8bf5b2147bc0be5cf123"
    }

Namespaces
----------

Gandalf supports namespaces for repositories and must be informed in the name of the repository followed by a single slash and the actual name of the repository, i.e. `mynamespace/myrepository`. Examples of usage:

* Creates a repository in a namespace:

    * Method: POST
    * URI: /repository
    * Format: JSON

    Example URL (http://gandalf-server omitted for clarity)::

        $ curl -XPOST /repository \
            -d '{"name": "mynamespace/myrepository", \
                "users": ["myuser"], \
                "readonlyusers": ["alice", "bob"]}'

* Returns a list of all the branches of the specified `mynamespace/myrepository`.

    * Method: GET
    * URI: //repository/`:name`/branches
    * Format: JSON

    Where:

    * `:name` is the name of the repository.

    Example URL (http://gandalf-server omitted for clarity)::

        $ curl /repository/mynamespace/myrepository/branches  # gets list of branches
